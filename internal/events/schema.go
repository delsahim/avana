package events

type CreateEventSchema struct {
	Name string
	Location string
	Organiser string
	IsPaidEvent bool
	IsLimitedEvent bool
	Description string
	MaxUnitReservation uint
	EventDate string
	RegistrationExpirationDate string
	TotalTicketLimit uint
	Tickets []TicketSchema
}


type TicketSchema struct {
	Name string
	Price float64
	TotalAvailable uint
	SingleLimit uint
	ExpiryTime string
}


type UpdateTicketSchema struct {
	Name          string   `binding:"omitempty"` 
	Price         *float64   `binding:"omitempty"` 
	TotalAvailable *uint      `binding:"omitempty"` 
	SingleLimit   *uint     `binding:"omitempty"` 
	ExpiryTime    string `binding:"omitempty"` 
}

type UpdateEventSchema struct {
	Name                      string `binding:"omitempty"`                      // optional
	Location                  string `binding:"omitempty"`                  // optional
	Organiser                 string `binding:"omitempty"`                 // optional
	IsPaidEvent               *bool  `binding:"omitempty"`             // optional
	IsLimitedEvent 			  *bool  `binding:"omitempty"`
	Description               string `binding:"omitempty"`               // optional
	MaxUnitReservation        *uint  `binding:"omitempty"`      // optional
	EventDate                 string `binding:"omitempty"`            // optional
	RegistrationExpirationDate string `binding:"omitempty"` // optional
	TotalTicketLimit 		   *uint	`binding:"omitempty"`		// optional
}

type BuyTicketScema struct {
	Units uint
}

type GetAllAttendees struct {
	Email string
	TicketType string
	Amount float64
}

type GetAllReviewSchema struct {

}

