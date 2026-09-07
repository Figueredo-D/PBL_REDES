package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"

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

// Send monta o envelope genérico e insere o payload interno na propriedade Data
func (c *Client) Send(action string, payload interface{}) (*protocol.Response, error) {
	var payloadString string

	if payload != nil {
		bytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar payload interno: %w", err)
		}
		payloadString = string(bytes)
	}

	// Envelope Genérico
	req := protocol.Request{
		Action: action,
		Data:   payloadString,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar envelope: %w", err)
	}
	reqBytes = append(reqBytes, '\n')

	if _, err := c.conn.Write(reqBytes); err != nil {
		return nil, err
	}

	respBytes, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var resp protocol.Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func main() {
	client, err := NewClient("localhost:8080")
	if err != nil {
		fmt.Printf("[ERRO] Não foi possível conectar: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("=== TESTE DE ENVELOPE GENÉRICO ===")

	searchData := protocol.SearchItineraryData{
		Origin:      "Salvador",
		Destination: "Vitória da Conquista",
		Date:        "2026-09-20",
	}

	resp, err := client.Send(protocol.ActionSearchItinerary, searchData)
	if err != nil {
		fmt.Printf("[ERRO]: %v\n", err)
		return
	}

	fmt.Printf("[RESPOSTA] Status: %s | Msg: %s\n", resp.Status, resp.Message)
}