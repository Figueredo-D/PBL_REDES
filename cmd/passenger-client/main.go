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
	fmt.Println("        SISTEMA DE CARONAS - PASSAGEIRO           ")
	fmt.Println("==================================================")
	fmt.Print("Digite seu Token de Passageiro: ")
	scanner.Scan()
	passengerToken := strings.TrimSpace(scanner.Text())
	if passengerToken == "" {
		passengerToken = "passageiro_default"
	}

	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = "localhost"
	}

	client, err := NewClient(serverHost + ":8080")
	if err != nil {
		fmt.Printf("\n[ERRO CONEXÃO] Não foi possível conectar ao servidor (%s:8080): %v\n", serverHost, err)
		os.Exit(1)
	}
	defer client.Close()

	for {
		fmt.Println("\n--------------------------------------------------")
		fmt.Printf("   PAINEL DO PASSAGEIRO | Usuário: [%s]\n", passengerToken)
		fmt.Println("--------------------------------------------------")
		fmt.Println("  [1] Consultar e Reservar Itinerários (BFS)")
		fmt.Println("  [2] Minhas Reservas")
		fmt.Println("  [3] Cancelar uma Reserva")
		fmt.Println("  [0] Sair")
		fmt.Print("Escolha uma opção: ")

		scanner.Scan()
		option := strings.TrimSpace(scanner.Text())

		switch option {
		case "1":
			handleSearchAndBook(client, scanner, passengerToken)
		case "2":
			handleListBookings(client, passengerToken)
		case "3":
			handleCancelBooking(client, scanner, passengerToken)
		case "0":
			fmt.Println("\nSaindo...")
			return
		default:
			fmt.Println("\n[!] Opção inválida.")
		}
	}
}

func handleSearchAndBook(client *Client, scanner *bufio.Scanner, token string) {
	fmt.Println("\n==================================================")
	fmt.Println("             CONSULTAR VIAGENS                   ")
	fmt.Println("==================================================")

	fmt.Print("Origem: ")
	scanner.Scan()
	origin := strings.TrimSpace(scanner.Text())

	fmt.Print("Destino: ")
	scanner.Scan()
	destination := strings.TrimSpace(scanner.Text())

	fmt.Print("Data (AAAA-MM-DD): ")
	scanner.Scan()
	date := strings.TrimSpace(scanner.Text())

	searchPayload := protocol.SearchItineraryData{
		Origin:      origin,
		Destination: destination,
		Date:        date,
	}

	resp, err := client.Send(protocol.ActionSearchItinerary, token, searchPayload)
	if err != nil || resp.Status != "SUCCESS" {
		fmt.Printf("\n[ERRO]: %s\n", resp.Message)
		return
	}

	var itineraries []domain.ItineraryResult
	json.Unmarshal([]byte(resp.Data), &itineraries)

	if len(itineraries) == 0 {
		fmt.Println("\nNenhum itinerário encontrado para essa data/rota.")
		return
	}

	fmt.Printf("\n========================================================================================\n")
	fmt.Printf("               ITINERÁRIOS ENCONTRADOS (%d OPÇÕES)                                      \n", len(itineraries))
	fmt.Printf("========================================================================================\n")

	for idx, itin := range itineraries {
		tipo := "DIRETA"
		if len(itin.Drivers) > 1 {
			tipo = fmt.Sprintf("COMBINADA (%d TROCAS)", len(itin.Drivers)-1)
		}

		fmt.Printf("\n+--------------------------------------------------------------------------------------+\n")
		fmt.Printf("  OPÇÃO #%d  |  Tipo: %s\n", idx+1, tipo)
		fmt.Printf("  Saída: %s hrs  |  Vagas Livres: %d  |  PREÇO TOTAL: R$ %.2f\n",
			itin.DepartureTime, itin.AvailableSeats, itin.TotalPrice)
		fmt.Printf("  Motoristas: %s\n", strings.Join(itin.Drivers, " -> "))
		fmt.Printf("+--------------------------------------------------------------------------------------+ \n")

		fmt.Println("  +----+------------------------+------------------------+-------------------+----------+---------+")
		fmt.Println("  | #  | ORIGEM                 | DESTINO                | MOTORISTA         | SAÍDA    | PREÇO   |")
		fmt.Println("  +----+------------------------+------------------------+-------------------+----------+---------+")
		for j, seg := range itin.Segments {
			fmt.Printf("  | %-2d | %-22s | %-22s | %-17s | %-8s | R$%-6.2f |\n",
				j+1, truncate(seg.From, 22), truncate(seg.To, 22), truncate(seg.DriverToken, 17), seg.DepartureTime, seg.Price)
		}
		fmt.Println("  +----+------------------------+------------------------+-------------------+----------+---------+")
	}

	fmt.Print("\nDeseja reservar alguma opção acima? (Digite o número da opção 1..N ou 0 para cancelar): ")
	scanner.Scan()
	choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || choice < 1 || choice > len(itineraries) {
		fmt.Println("Operação cancelada.")
		return
	}

	selectedItin := itineraries[choice-1]
	bookResp, err := client.Send(protocol.ActionBookItinerary, token, selectedItin)
	if err != nil {
		fmt.Printf("\n[ERRO TCP]: %v\n", err)
		return
	}

	if bookResp.Status == "SUCCESS" {
		fmt.Printf("\n[SUCESSO] %s\n", bookResp.Message)
	} else {
		fmt.Printf("\n[FALHA] [%s] %s\n", bookResp.Code, bookResp.Message)
	}
}

func handleListBookings(client *Client, token string) {
	resp, err := client.Send(protocol.ActionListBookings, token, nil)
	if err != nil || resp.Status != "SUCCESS" {
		fmt.Println("\n[ERRO] Não foi possível carregar reservas.")
		return
	}

	var bookings []*domain.Booking
	json.Unmarshal([]byte(resp.Data), &bookings)

	if len(bookings) == 0 {
		fmt.Println("\nNenhuma reserva encontrada.")
		return
	}

	fmt.Println("\n==========================================================================================")
	fmt.Println("                                MINHAS RESERVAS                                           ")
	fmt.Println("==========================================================================================")
	for _, b := range bookings {
		fmt.Printf("\n  ID DA RESERVA: %s  |  Status: %s  |  Data da Reserva: %s  |  Valor Total: R$ %.2f\n",
			b.ID, b.Status, b.Date, b.Itinerary.TotalPrice)

		fmt.Println("  +----+------------------------+------------------------+-------------------+----------+---------+")
		fmt.Println("  | #  | ORIGEM                 | DESTINO                | MOTORISTA         | SAÍDA    | PREÇO   |")
		fmt.Println("  +----+------------------------+------------------------+-------------------+----------+---------+")
		for j, seg := range b.Itinerary.Segments {
			fmt.Printf("  | %-2d | %-22s | %-22s | %-17s | %-8s | R$%-6.2f |\n",
				j+1, truncate(seg.From, 22), truncate(seg.To, 22), truncate(seg.DriverToken, 17), seg.DepartureTime, seg.Price)
		}
		fmt.Println("  +----+------------------------+------------------------+-------------------+----------+---------+")
	}
}

func handleCancelBooking(client *Client, scanner *bufio.Scanner, token string) {
	handleListBookings(client, token)

	fmt.Print("\nDigite o ID da Reserva que deseja CANCELAR (ex: BOOK_0001): ")
	scanner.Scan()
	bookingID := strings.TrimSpace(scanner.Text())

	if bookingID == "" {
		return
	}

	cancelPayload := protocol.CancelBookingData{BookingID: bookingID}
	resp, err := client.Send(protocol.ActionCancelBooking, token, cancelPayload)

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