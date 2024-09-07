package users

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email string			`gorm:"unique; not null"`
	Password string			`gorm:"not null"`
	FirstName string		`gorm:"not null"`
	LastName string			`gorm:"not null"`
	Otp string
	OtpExpires time.Time
	OtpVerified bool 		`gorm:"default:false"`

}