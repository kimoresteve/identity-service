package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateOTP() (string, error) {
	// Set the maximum number (exclusive) - 1,000,000 for 6 digits
	max := big.NewInt(1000000)

	// Generate random number
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %v", err)
	}

	// Format as 6-digit string with leading zeros
	otp := fmt.Sprintf("%06d", n)
	return otp, nil
}
