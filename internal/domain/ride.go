package domain

import (
	"fmt"
	"sync"
)

type Segment struct {
	From           string  `json:"from"`
	To             string  `json:"to"`
	Price          float64 `json:"price"`
	AvailableSeats int     `json:"available_seats"`
}

type Ride struct {
	ID          string     `json:"id"`
	DriverToken string     `json:"driver_token"`
	Route       []string   `json:"route"`
	Date        string     `json:"date"`
	Time        string     `json:"time"`
	TotalSeats  int        `json:"total_seats"`
	Segments    []*Segment `json:"segments"`
	mu          sync.RWMutex
}

func NewRide(id, driverToken string, route []string, date, timeStr string, totalSeats int, prices []float64) (*Ride, error) {
	if len(route) < 2 {
		return nil, fmt.Errorf("a rota precisa de pelo menos 2 cidades")
	}

	if len(prices) != len(route)-1 {
		return nil, fmt.Errorf("quantidade de preços (%d) difere da quantidade de trechos (%d)", len(prices), len(route)-1)
	}

	if totalSeats <= 0 {
		return nil, fmt.Errorf("a quantidade de assentos deve ser maior que zero")
	}

	var segments []*Segment
	for i := 0; i < len(route)-1; i++ {
		segments = append(segments, &Segment{
			From:           route[i],
			To:             route[i+1],
			Price:          prices[i],
			AvailableSeats: totalSeats,
		})
	}

	return &Ride{
		ID:          id,
		DriverToken: driverToken,
		Route:       route,
		Date:        date,
		Time:        timeStr,
		TotalSeats:  totalSeats,
		Segments:    segments,
	}, nil
}