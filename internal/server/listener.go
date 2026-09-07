package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"github.com/Figueredo-D/PBL_REDES/internal/protocol"
)

type Server struct {
	address  string
	listener net.Listener
	mu       sync.Mutex
	active   bool
}

func NewServer(address string) *Server {
	return &Server{address: address}
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("falha ao abrir socket na porta %s: %w", s.address, err)
	}
	s.listener = l
	s.active = true

	fmt.Printf("[SERVER] Escutando em TCP %s\n", s.address)

	for s.active {
		conn, err := l.Accept()
		if err != nil {
			if !s.active {
				break
			}
			fmt.Printf("[SERVER] Erro ao aceitar conexão: %v\n", err)
			continue
		}

		go s.handleConnection(conn)
	}

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()
	fmt.Printf("[SERVER] Conexão aberta: %s\n", remoteAddr)

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("[SERVER] Erro de leitura de %s: %v\n", remoteAddr, err)
			}
			break
		}

		var req protocol.Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendResponse(conn, protocol.Response{
				Status:  "ERROR",
				Code:    "MALFORMED_ENVELOPE",
				Message: "Envelope JSON malformado.",
			})
			continue
		}
		resp := s.dispatch(req)
		s.sendResponse(conn, resp)
	}

	fmt.Printf("[SERVER] Conexão fechada: %s\n", remoteAddr)
}

func (s *Server) dispatch(req protocol.Request) protocol.Response {
	switch req.Action {

	case protocol.ActionOfferRide:
		var data protocol.OfferRideData
		if err := json.Unmarshal([]byte(req.Data), &data); err != nil {
			return protocol.Response{Status: "ERROR", Code: "INVALID_DATA", Message: "Payload de oferta de carona inválido."}
		}
		// TODO:LÓGICA DE NEGÓCIO DA CARONA
		return protocol.Response{
			Status:  "SUCCESS",
			Message: fmt.Sprintf("Carona de %s para %s registrada com sucesso.", data.Route[0], data.Route[len(data.Route)-1]),
		}

	case protocol.ActionSearchItinerary:
		var data protocol.SearchItineraryData
		if err := json.Unmarshal([]byte(req.Data), &data); err != nil {
			return protocol.Response{Status: "ERROR", Code: "INVALID_DATA", Message: "Payload de busca inválido."}
		}
		//TODO: LÓGICA DE BUSCA DE ITINERÁRIO
		return protocol.Response{
			Status:  "SUCCESS",
			Message: fmt.Sprintf("Busca realizada para %s -> %s em %s.", data.Origin, data.Destination, data.Date),
		}

	case protocol.ActionBookItinerary:
		var data protocol.BookItineraryData
		if err := json.Unmarshal([]byte(req.Data), &data); err != nil {
			return protocol.Response{Status: "ERROR", Code: "INVALID_DATA", Message: "Payload de reserva inválido."}
		}
		//TODO:LÓGICA DE RESERVA 
		return protocol.Response{Status: "SUCCESS", Message: "Itinerário reservado com sucesso."}

	default:
		return protocol.Response{
			Status:  "ERROR",
			Code:    "UNKNOWN_ACTION",
			Message: fmt.Sprintf("Ação '%s' desconhecida pelo protocolo.", req.Action),
		}
	}
}

func (s *Server) sendResponse(conn net.Conn, resp protocol.Response) {
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	conn.Write(data)
}