package config

import (
	     	"gorm.io/driver/postgres"
	     	"log"
			"gorm.io/gorm"
)

var DB *gorm.DB


func ConnectToDb() {
	var err error
	dsn := "host=localhost user=postgres password=subomi7205 dbname=avana port=5432 sslmode=disable TimeZone=Africa/Lagos"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
	log.Fatal("Error to connect to database")
    	}
}
