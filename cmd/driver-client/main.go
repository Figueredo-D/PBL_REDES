package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/Figueredo-D/PBL_REDES/internal/domain"
	"github.com/Figueredo-D/PBL_REDES/internal/protocol"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (c *Client) Send(action, token string, payload interface{}) (*protocol.Response, error) {
	var payloadString string
	if payload != nil {
		bytes, _ := json.Marshal(payload)
		payloadString = string(bytes)
	}

	req := protocol.Request{Action: action, Token: token, Data: payloadString}
	reqBytes, _ := json.Marshal(req)
	reqBytes = append(reqBytes, '\n')

	c.conn.Write(reqBytes)

	respBytes, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var resp protocol.Response
	json.Unmarshal(respBytes, &resp)
	return &resp, nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("==================================================")
	fmt.Println("        SISTEMA DE CARONAS - MOTORISTA            ")
	fmt.Println("==================================================")
	fmt.Print("Digite seu Token de Motorista: ")
	scanner.Scan()
	driverToken := strings.TrimSpace(scanner.Text())
	if driverToken == "" {
		driverToken = "driver_default"
	}

	client, err := NewClient("localhost:8080")
	if err != nil {
		fmt.Printf("[ERRO] Não foi possível conectar: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	for {
		fmt.Println("\n--------------------------------------------------")
		fmt.Printf("   PAINEL DO MOTORISTA | Usuário: [%s]\n", driverToken)
		fmt.Println("--------------------------------------------------")
		fmt.Println("  [1] Ofertar Nova Carona")
		fmt.Println("  [2] Minhas Caronas Cadastradas")
		fmt.Println("  [3] Cancelar uma Carona")
		fmt.Println("  [0] Sair")
		fmt.Print("Escolha uma opção: ")

		scanner.Scan()
		option := strings.TrimSpace(scanner.Text())

		switch option {
		case "1":
			handleOfferRide(client, scanner, driverToken)
		case "2":
			handleListRides(client, driverToken)
		case "3":
			handleCancelRide(client, scanner, driverToken)
		case "0":
			fmt.Println("\nSaindo...")
			return
		default:
			fmt.Println("\n[!] Opção inválida.")
		}
	}
}

func handleOfferRide(client *Client, scanner *bufio.Scanner, token string) {
	fmt.Println("\n==================================================")
	fmt.Println("             CADASTRAR CARONA                     ")
	fmt.Println("==================================================")

	fmt.Print("Cidades da rota (separadas por vírgula): ")
	scanner.Scan()
	rawCities := strings.Split(scanner.Text(), ",")
	var route []string
	for _, c := range rawCities {
		city := strings.TrimSpace(c)
		if city != "" {
			route = append(route, city)
		}
	}

	if len(route) < 2 {
		fmt.Println("[ERRO] A rota deve ter pelo menos 2 cidades.")
		return
	}

	fmt.Print("Data (AAAA-MM-DD): ")
	scanner.Scan()
	date := strings.TrimSpace(scanner.Text())

	fmt.Print("Horário de Saída (HH:MM): ")
	scanner.Scan()
	timeStr := strings.TrimSpace(scanner.Text())

	fmt.Print("Quantidade de Assentos: ")
	scanner.Scan()
	seats, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || seats <= 0 {
		fmt.Println("[ERRO] Assentos inválidos.")
		return
	}

	numSegments := len(route) - 1
	var prices []float64
	fmt.Printf("\nPreço individual dos %d trechos:\n", numSegments)

	for i := 0; i < numSegments; i++ {
		fmt.Printf(" Trecho %s -> %s (R$): ", route[i], route[i+1])
		scanner.Scan()
		price, err := strconv.ParseFloat(strings.TrimSpace(scanner.Text()), 64)
		if err != nil || price <= 0 {
			fmt.Println("[ERRO] Preço inválido.")
			return
		}
		prices = append(prices, price)
	}

	offerPayload := protocol.OfferRideData{
		Route:         route,
		Date:          date,
		Time:          timeStr,
		TotalSeats:    seats,
		SegmentPrices: prices,
	}

	resp, err := client.Send(protocol.ActionOfferRide, token, offerPayload)
	if err != nil {
		fmt.Printf("[ERRO TCP]: %v\n", err)
		return
	}

	if resp.Status == "SUCCESS" {
		fmt.Printf("\n[SUCESSO] %s\n", resp.Message)
	} else {
		fmt.Printf("\n[FALHA] [%s] %s\n", resp.Code, resp.Message)
	}
}

func handleListRides(client *Client, token string) {
	resp, err := client.Send(protocol.ActionListRides, token, nil)
	if err != nil || resp.Status != "SUCCESS" {
		fmt.Println("\n[ERRO] Não foi possível consultar suas caronas.")
		return
	}

	var rides []*domain.Ride
	json.Unmarshal([]byte(resp.Data), &rides)

	if len(rides) == 0 {
		fmt.Println("\nNenhuma carona cadastrada.")
		return
	}

	fmt.Println("\n==========================================================================================")
	fmt.Println("                               MINHAS CARONAS OFERTADAS                                   ")
	fmt.Println("==========================================================================================")
	for _, ride := range rides {
		fmt.Printf("\n  ID DA CARONA: %s  |  Data: %s  |  Saída: %s hrs  |  Vagas Totais: %d\n",
			ride.ID, ride.Date, ride.Time, ride.TotalSeats)

		fmt.Println("  +----+------------------------+------------------------+----------+-------------------+")
		fmt.Println("  | #  | ORIGEM                 | DESTINO                | PREÇO    | VAGAS DISPONÍVEIS |")
		fmt.Println("  +----+------------------------+------------------------+----------+-------------------+")
		for j, seg := range ride.Segments {
			fmt.Printf("  | %-2d | %-22s | %-22s | R$%-6.2f | %-17d |\n",
				j+1, truncate(seg.From, 22), truncate(seg.To, 22), seg.Price, seg.AvailableSeats)
		}
		fmt.Println("  +----+------------------------+------------------------+----------+-------------------+")
	}
}

func handleCancelRide(client *Client, scanner *bufio.Scanner, token string) {
	handleListRides(client, token)

	fmt.Print("\nDigite o ID da Carona que deseja CANCELAR (ex: RIDE_0001): ")
	scanner.Scan()
	rideID := strings.TrimSpace(scanner.Text())

	if rideID == "" {
		return
	}

	cancelPayload := protocol.CancelRideData{RideID: rideID}
	resp, err := client.Send(protocol.ActionCancelRide, token, cancelPayload)

	if err != nil {
		fmt.Printf("\n[ERRO TCP]: %v\n", err)
		return
	}

	if resp.Status == "SUCCESS" {
		fmt.Printf("\n[SUCESSO] %s\n", resp.Message)
	} else {
		fmt.Printf("\n[FALHA] [%s] %s\n", resp.Code, resp.Message)
	}
}

func truncate(str string, max int) string {
	if len(str) > max {
		return str[:max-3] + "..."
	}
	return str
}