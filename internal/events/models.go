package events

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model

	// other fields
	Name string				`gorm:"unique;not null"`
	Organiser string		`gorm:"not null"`
	Location string			`gorm:"not null"`
	IsPaidEvent bool		`gorm:"not null"`
	Description string		`gorm:"not null;type:TEXT"`
	IsLimited bool			`gorm:"not null"`
	MaxUnitReservation uint		`gorm:"not null;default:1"`
	EventDate time.Time		`gorm:"not null"`
	RegistrationExpirationDate time.Time	`gorm:"not null"`
	UserID uint
}

type Ticket struct {
	gorm.Model

	// other fields 
	Name string			`gorm:"not null"`
	Price float64		`gorm:"not null"`
	TotalAvailable uint		`gorm:"not null"`
	SingleLimit uint	`gorm:"not null"`
	ExpiryTime time.Time	`gorm:"not null"`
	EventID uint

}

type Attendee struct {
	gorm.Model

	//other fields
	UserID uint
	Units uint		`gorm:"not null"`
	Attended bool			`gorm:"not null;default:false"`
	Review string 			
	Rating uint
	TicketID uint
}