package model

type Monitoring struct {
	TotalReservations int `json:"total_reservations"`
	AvailableSpots    int `json:"available_spots"`
	ErrorsLogged      int `json:"errors_logged"`
}
