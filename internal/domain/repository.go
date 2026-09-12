package domain

import (
	"fmt"
	"sync"
	"time"
)

type Booking struct {
	ID             string
	PassengerToken string
	Itinerary      ItineraryResult
	Date           string
	Status         string
}

type RideRepository struct {
	mu           sync.RWMutex
	rides        map[string]*Ride
	bookings     map[string]*Booking
	rideCount    int
	bookingCount int
}

func NewRideRepository() *RideRepository {
	return &RideRepository{
		rides:    make(map[string]*Ride),
		bookings: make(map[string]*Booking),
	}
}

func (r *RideRepository) Add(ride *Ride) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rideCount++
	ride.ID = fmt.Sprintf("RIDE_%04d", r.rideCount)
	r.rides[ride.ID] = ride
	return ride.ID
}

func (r *RideRepository) ListByDriver(driverToken string) []*Ride {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Ride
	for _, ride := range r.rides {
		if ride.DriverToken == driverToken {
			result = append(result, ride)
		}
	}
	return result
}

func (r *RideRepository) CancelRide(rideID, driverToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ride, exists := r.rides[rideID]
	if !exists {
		return fmt.Errorf("carona não encontrada")
	}

	if ride.DriverToken != driverToken {
		return fmt.Errorf("você não tem permissão para cancelar esta carona")
	}

	// Deleta a carona do mapa
	delete(r.rides, rideID)

	// Cancela reservas associadas a essa carona
	for _, b := range r.bookings {
		for _, seg := range b.Itinerary.Segments {
			if seg.RideID == rideID {
				b.Status = "CANCELLED_BY_DRIVER"
			}
		}
	}

	return nil
}

func (r *RideRepository) ListByDate(date string) []*Ride {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Ride
	for _, ride := range r.rides {
		if ride.Date == date {
			result = append(result, ride)
		}
	}
	return result
}

func (r *RideRepository) SearchItineraries(filter SearchFilter) []ItineraryResult {
	ridesOnDate := r.ListByDate(filter.Date)
	return FindItineraries(ridesOnDate, filter)
}

func (r *RideRepository) BookItinerary(passengerToken string, itinerary ItineraryResult) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Valida e decrementa vagas de forma atômica
	for _, segDetail := range itinerary.Segments {
		ride, exists := r.rides[segDetail.RideID]
		if !exists {
			return "", fmt.Errorf("uma das caronas do itinerário não existe mais")
		}

		ride.mu.Lock()
		var foundSeg *Segment
		for _, s := range ride.Segments {
			if s.From == segDetail.From && s.To == segDetail.To {
				foundSeg = s
				break
			}
		}

		if foundSeg == nil || foundSeg.AvailableSeats <= 0 {
			ride.mu.Unlock()
			return "", fmt.Errorf("não há assentos disponíveis no trecho %s -> %s", segDetail.From, segDetail.To)
		}
		ride.mu.Unlock()
	}

	// Decrementa assentos após validar todos os trechos
	for _, segDetail := range itinerary.Segments {
		ride := r.rides[segDetail.RideID]
		ride.mu.Lock()
		for _, s := range ride.Segments {
			if s.From == segDetail.From && s.To == segDetail.To {
				s.AvailableSeats--
			}
		}
		ride.mu.Unlock()
	}

	r.bookingCount++
	bookingID := fmt.Sprintf("BOOK_%04d", r.bookingCount)
	booking := &Booking{
		ID:             bookingID,
		PassengerToken: passengerToken,
		Itinerary:      itinerary,
		Date:           time.Now().Format("2006-01-02"),
		Status:         "CONFIRMED",
	}

	r.bookings[bookingID] = booking
	return bookingID, nil
}

func (r *RideRepository) CancelBooking(bookingID, passengerToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	booking, exists := r.bookings[bookingID]
	if !exists {
		return fmt.Errorf("reserva não encontrada")
	}

	if booking.PassengerToken != passengerToken {
		return fmt.Errorf("você não tem permissão para cancelar esta reserva")
	}

	if booking.Status == "CANCELLED" {
		return fmt.Errorf("esta reserva já está cancelada")
	}

	// Devolve os assentos reservantes nos trechos
	for _, segDetail := range booking.Itinerary.Segments {
		if ride, exists := r.rides[segDetail.RideID]; exists {
			ride.mu.Lock()
			for _, s := range ride.Segments {
				if s.From == segDetail.From && s.To == segDetail.To {
					s.AvailableSeats++
				}
			}
			ride.mu.Unlock()
		}
	}

	booking.Status = "CANCELLED"
	return nil
}

func (r *RideRepository) ListBookingsByPassenger(passengerToken string) []*Booking {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Booking
	for _, b := range r.bookings {
		if b.PassengerToken == passengerToken {
			result = append(result, b)
		}
	}
	return result
}