package models

import "time"

type ClientType string

const (
	ClientTypeAgency   ClientType = "agency"
	ClientTypeLandlord ClientType = "landlord"
	ClientTypeUser     ClientType = "user"
)

type Client struct {
	ID         int        `json:"id"`
	UUID       string     `json:"uuid"`
	Type       ClientType `json:"type"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Contact    string     `json:"contact"`
	Password   string     `json:"password"`
	IsVerified bool       `json:"is_verified"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Agency struct {
	Client
	Name    string `json:"name"`
	Address string `json:"address"`
	TaxID   string `json:"tax_id"`
	LogoURL string `json:"logo_url"`
}

type Landlord struct {
	Client
	Name     string `json:"name"`
	Address  string `json:"address"`
	AgencyID *int   `json:"agency_id,omitempty"`
}

type User struct {
	Client
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
	OwnerID   int    `json:"owner_id"`
	OwnerType string `json:"owner_type"` // "agency" or "landlord"
}

type ClientInput struct {
	Type     ClientType `json:"type"`
	Email    string     `json:"email"`
	Name     string     `json:"name"`
	Contact  string     `json:"contact"`
	Password string     `json:"password"`
}

type LoginInput struct {
	Contact  string `json:"contact"`
	Password string `json:"password"`
}

type Verify struct {
	Otp      string `json:"otp"`
	ClientID int    `json:"id"`
}

type ResendOTPInput struct {
	LandlordID int    `json:"id"`
	Purpose    string `json:"purpose"`
}

type ForgotPasswordInput struct {
	Contact string `json:"contact"`
}

type ResetPasswordInput struct {
	Password string `json:"password"`
	OTP      string `json:"otp"`
	ID       int    `json:"id"`
}
