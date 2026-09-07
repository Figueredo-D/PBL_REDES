package protocol

const (
	ActionAuthenticate    = "AUTHENTICATE"
	ActionOfferRide       = "OFFER_RIDE"
	ActionListRides       = "LIST_RIDES"
	ActionCancelRide      = "CANCEL_RIDE"
	ActionSearchItinerary = "SEARCH_ITINERARY"
	ActionBookItinerary   = "BOOK_ITINERARY"
	ActionCancelBooking   = "CANCEL_BOOKING"
	ActionListBookings    = "LIST_BOOKINGS"
)

type Request struct {
	Action string `json:"action"`
	Token  string `json:"token,omitempty"`
	Data   string `json:"data,omitempty"` 
}

type Response struct {
	Status  string `json:"status"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    string `json:"data,omitempty"`
}

type AuthData struct {
	Username string `json:"username"`
	UserType string `json:"user_type"`
}

type OfferRideData struct {
	Route         []string  `json:"route"`
	Date          string    `json:"date"`
	Time          string    `json:"time"`
	TotalSeats    int       `json:"total_seats"`
	SegmentPrices []float64 `json:"segment_prices"`
}

type SearchItineraryData struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Date        string `json:"date"`
}

type SegmentInfo struct {
	RideID    string  `json:"ride_id"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	Driver    string  `json:"driver"`
	Price     float64 `json:"price"`
	SeatsLeft int     `json:"seats_left"`
}

type ItineraryOption struct {
	ItineraryID string        `json:"itinerary_id"`
	Segments    []SegmentInfo `json:"segments"`
	TotalPrice  float64       `json:"total_price"`
}

type BookSegment struct {
	RideID string `json:"ride_id"`
	From   string `json:"from"`
	To     string `json:"to"`
}

type BookItineraryData struct {
	Segments []BookSegment `json:"segments"`
}

type CancelBookingData struct {
	BookingID string `json:"booking_id"`
}

type CancelRideData struct {
	RideID string `json:"ride_id"`
}