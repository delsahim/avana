package main

import (
	"avana/internal/config"
	"avana/internal/events"
	"avana/internal/users"
)


func init() {
	config.ConnectToDb()
}

func main() {
	config.DB.AutoMigrate(
		&users.User{},
		&events.Event{},
		&events.Attendee{},
		&events.Ticket{},
	)
}