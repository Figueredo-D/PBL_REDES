package domain

import (
	"time"
)

type SearchFilter struct {
	Origin      string
	Destination string
	Date        string
}

// detalha o trecho individual no itinerário para o passageiro!!!!!!!!!
type ItinerarySegmentDetail struct {
	RideID         string  `json:"ride_id"`
	DriverToken    string  `json:"driver_token"`
	From           string  `json:"from"`
	To             string  `json:"to"`
	DepartureTime  string  `json:"departure_time"`
	Price          float64 `json:"price"`
	AvailableSeats int     `json:"available_seats"`
}

//representa a rota completa encontrada pelo algoritmo (direta ou combinada)
type ItineraryResult struct {
	ItineraryID    string                   `json:"itinerary_id"`
	Segments       []ItinerarySegmentDetail `json:"segments"`
	TotalPrice     float64                  `json:"total_price"`
	DepartureTime  string                   `json:"departure_time"`
	AvailableSeats int                      `json:"available_seats"` // Vagas gargalo (menor valor entre os trechos)
	Drivers        []string                 `json:"drivers"`         // Lista de motoristas na rota
}

//representa uma aresta no Grafo de viagens
type GraphEdge struct {
	Ride        *Ride
	Segment     *Segment
	From        string
	To          string
	DriverToken string
}

//é um nó utilizado na travessia BFS para rastrear o caminho percorrido
type PathNode struct {
	City  string
	Edges []GraphEdge
}

//utiliza Busca em Grafo (BFS) para mapear trajetos diretos e combinados
func FindItineraries(rides []*Ride, filter SearchFilter) []ItineraryResult {
	var results []ItineraryResult
	now := time.Now()

	graph := make(map[string][]GraphEdge)

	for _, ride := range rides {
		if ride.Date != filter.Date {
			continue
		}

		if ride.Date == now.Format("2006-01-02") {
			departureDateTime, err := time.Parse("2006-01-02 15:04", ride.Date+" "+ride.Time)
			if err == nil && departureDateTime.Before(now) {
				continue
			}
		}

		ride.mu.RLock()
		for _, seg := range ride.Segments {
			if seg.AvailableSeats > 0 {
				graph[seg.From] = append(graph[seg.From], GraphEdge{
					Ride:        ride,
					Segment:     seg,
					From:        seg.From,
					To:          seg.To,
					DriverToken: ride.DriverToken,
				})
			}
		}
		ride.mu.RUnlock()
	}
	queue := []PathNode{{City: filter.Origin, Edges: []GraphEdge{}}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.City == filter.Destination && len(current.Edges) > 0 {
			res := buildItineraryResult(current.Edges)
			results = append(results, res)
			continue
		}

		// Limite de conexões/trocas de carona para evitar trajetos gigantes ou loops
		if len(current.Edges) >= 4 {
			continue
		}

		for _, edge := range graph[current.City] {
			if hasVisitedCity(current.Edges, edge.To) {
				continue
			}

			if len(current.Edges) > 0 {
				lastEdge := current.Edges[len(current.Edges)-1]
				if edge.Ride.ID != lastEdge.Ride.ID && edge.Ride.Time < lastEdge.Ride.Time {
					continue // nao permite conexão com horário anterior ao trecho de partida
				}
			}

			newEdges := append([]GraphEdge{}, current.Edges...)
			newEdges = append(newEdges, edge)

			queue = append(queue, PathNode{
				City:  edge.To,
				Edges: newEdges,
			})
		}
	}

	return results
}

func hasVisitedCity(edges []GraphEdge, city string) bool {
	for _, e := range edges {
		if e.From == city || e.To == city {
			return true
		}
	}
	return false
}

func buildItineraryResult(edges []GraphEdge) ItineraryResult {
	var segments []ItinerarySegmentDetail
	var totalPrice float64
	var drivers []string
	driverSet := make(map[string]bool)

	minSeats := edges[0].Segment.AvailableSeats
	firstDeparture := edges[0].Ride.Time

	for _, e := range edges {
		totalPrice += e.Segment.Price

		if e.Segment.AvailableSeats < minSeats {
			minSeats = e.Segment.AvailableSeats
		}

		if !driverSet[e.DriverToken] {
			driverSet[e.DriverToken] = true
			drivers = append(drivers, e.DriverToken)
		}

		segments = append(segments, ItinerarySegmentDetail{
			RideID:         e.Ride.ID,
			DriverToken:    e.DriverToken,
			From:           e.Segment.From,
			To:             e.Segment.To,
			DepartureTime:  e.Ride.Time,
			Price:          e.Segment.Price,
			AvailableSeats: e.Segment.AvailableSeats,
		})
	}

	return ItineraryResult{
		ItineraryID:    edges[0].Ride.ID,
		Segments:       segments,
		TotalPrice:     totalPrice,
		DepartureTime:  firstDeparture,
		AvailableSeats: minSeats,
		Drivers:        drivers,
	}
}