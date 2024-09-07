package utils

import (
	"fmt"
	"math/rand"
	"time"
)


func GenerateOtp() string {
    // Seed the random number generator with the current time
    source := rand.NewSource(time.Now().UnixNano())
    rng := rand.New(source)

    // Generate a random 5-digit number
    otp := rng.Intn(100000)
    otpString := fmt.Sprintf("%05d", otp)

    return otpString
}
