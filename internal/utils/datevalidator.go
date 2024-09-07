package utils

import (
	"errors"
	"time"
)

func ValidateDate(dateString string) (time.Time, error) {
	// the format is in YY-MM-DD HH:MM:SS
	// the hour is 24 hour format
	format := "2006-01-02 15:04:05"

  
	// Parse the DoB string into a time.Time instance
	t, err := time.Parse(format, dateString)
  
	// Check for parsing errors
	if err != nil {
	  return time.Time{}, err
	}
	if !t.After(time.Now()) {
		return time.Time{}, errors.New("invalid time")
	}
  
	// Return the parsed time.Time instance
	return t, nil
}

