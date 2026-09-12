package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Figueredo-D/PBL_REDES/internal/domain"
	"github.com/Figueredo-D/PBL_REDES/internal/protocol"
)

func sendTCPRequest(address string, req protocol.Request) (*protocol.Response, error) {
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	reqBytes, _ := json.Marshal(req)
	reqBytes = append(reqBytes, '\n')
	conn.Write(reqBytes)

	reader := bufio.NewReader(conn)
	respBytes, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var resp protocol.Response
	json.Unmarshal(respBytes, &resp)
	return &resp, nil
}

func main() {
	serverAddr := "localhost:8080"
	driverToken := "driver_stress"
	totalSeats := 3       // Apenas 3 vagas disponíveis
	totalClients := 30    // 30 passageiros tentando reservar ao mesmo tempo

	fmt.Println("==================================================")
	fmt.Println("       TESTE DE ESTRESSE E CONCORRÊNCIA TCP       ")
	fmt.Println("==================================================")
	fmt.Printf("Configuração: %d vagas disponíveis para %d passageiros concorrentes.\n\n", totalSeats, totalClients)

	// Step 1: Motorista cadastra uma carona de apenas 3 vagas
	fmt.Println("[1] Motorista criando carona com 3 assentos...")
	offerPayload := protocol.OfferRideData{
		Route:         []string{"Salvador", "Feira de Santana"},
		Date:          "2026-10-10",
		Time:          "08:00",
		TotalSeats:    totalSeats,
		SegmentPrices: []float64{35.0},
	}

	offerPayloadBytes, _ := json.Marshal(offerPayload)
	reqOffer := protocol.Request{
		Action: protocol.ActionOfferRide,
		Token:  driverToken,
		Data:   string(offerPayloadBytes),
	}

	respOffer, err := sendTCPRequest(serverAddr, reqOffer)
	if err != nil || respOffer.Status != "SUCCESS" {
		fmt.Printf("[ERRO] Falha ao cadastrar carona inicial: %v\n", err)
		return
	}
	fmt.Printf("[OK] Carona criada com sucesso! Resposta: %s\n\n", respOffer.Message)

	// Step 2: Fazer a busca do itinerário para obter o payload do trecho
	fmt.Println("[2] Consultando itinerário para obter os dados do trecho...")
	searchPayload := protocol.SearchItineraryData{
		Origin:      "Salvador",
		Destination: "Feira de Santana",
		Date:        "2026-10-10",
	}
	searchBytes, _ := json.Marshal(searchPayload)
	reqSearch := protocol.Request{
		Action: protocol.ActionSearchItinerary,
		Token:  "passageiro_teste",
		Data:   string(searchBytes),
	}

	respSearch, err := sendTCPRequest(serverAddr, reqSearch)
	if err != nil || respSearch.Status != "SUCCESS" {
		fmt.Printf("[ERRO] Falha ao buscar itinerário: %v\n", err)
		return
	}

	var itineraries []domain.ItineraryResult
	json.Unmarshal([]byte(respSearch.Data), &itineraries)

	if len(itineraries) == 0 {
		fmt.Println("[ERRO] Nenhum itinerário encontrado.")
		return
	}

	selectedItinerary := itineraries[0]
	itineraryBytes, _ := json.Marshal(selectedItinerary)

	// Step 3: Disparo simultâneo de 30 Goroutines tentando reservar
	fmt.Printf("[3] Disparando %d requisições simultâneas de reserva em paralelo...\n", totalClients)

	var wg sync.WaitGroup
	var successCount int32
	var failCount int32

	startSignal := make(chan struct{})

	for i := 1; i <= totalClients; i++ {
		wg.Add(1)
		passengerToken := fmt.Sprintf("passageiro_concorrente_%02d", i)

		go func(token string) {
			defer wg.Done()

			// Espera o sinal de partida para que TODAS as conexões disparem exatamente no mesmo instante
			<-startSignal

			reqBook := protocol.Request{
				Action: protocol.ActionBookItinerary,
				Token:  token,
				Data:   string(itineraryBytes),
			}

			respBook, err := sendTCPRequest(serverAddr, reqBook)
			if err == nil && respBook.Status == "SUCCESS" {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
		}(passengerToken)
	}

	// Libera todas as goroutines simultaneamente
	close(startSignal)

	// Aguarda todas terminarem
	wg.Wait()

	fmt.Println("\n==================================================")
	fmt.Println("             RESULTADO DO TESTE                   ")
	fmt.Println("==================================================")
	fmt.Printf("Reservas CONFIRMADAS (Sucesso): %d\n", successCount)
	fmt.Printf("Reservas REJEITADAS (Sem vaga): %d\n", failCount)

	if successCount == int32(totalSeats) && failCount == int32(totalClients-totalSeats) {
		fmt.Println("\n[TESTE APROVADO] O servidor controlou a concorrência com perfeição!")
		fmt.Println("Zero casos de Overbooking. Apenas o número limite de vagas foi preenchido.")
	} else {
		fmt.Println("\n[TESTE FALHOU] Ocorreu inconsistência na contagem de vagas.")
	}
}