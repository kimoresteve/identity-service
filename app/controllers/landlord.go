package controllers

import (
	"github.com/kimoresteve/identity-service/app/models"
	"github.com/kimoresteve/identity-service/app/utils"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

//type LandlordInput struct {
//	models.ClientInput
//	Address  string `json:"address"`
//	AgencyID *int   `json:"agency_id,omitempty"`
//}

// RegisterLandlord creates a landlord account.
// @Summary Landlord Account Creation
// @Description Registers a new landlord with contact, email, password, and address. Optionally can be associated with an agency.
// @Tags Landlord
// @Accept json
// @Produce json
// @Param landlord body LandlordInput true "Landlord Registration Input"
// @Success 201 {object} map[string]interface{} "Landlord created successful"
// @Failure 400 {string} string "Invalid request or missing fields"
// @Failure 409 {string} string "Email or contact already exists"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/register/landlord [post]
func (c *Controller) RegisterLandlord(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input LandlordInput
	if err := utils.DecodeAndValidateJSONBody(w, r, &input); err != nil {
		var apiErr *utils.APIError
		if errors.As(err, &apiErr) {
			http.Error(w, apiErr.Message, apiErr.Status)
		}
		return
	}

	// Validate required fields for landlord
	if input.Name == "" || input.Contact == "" || input.Password == "" || input.Email == "" {
		http.Error(w, "Name, Email, contact and password are required", http.StatusBadRequest)
		return
	}

	if input.Address == "" {
		http.Error(w, "Address is required for landlords", http.StatusBadRequest)
		return
	}

	// Validate agency exists if AgencyID is provided
	if input.AgencyID != nil {
		var count int
		err := c.DB.QueryRow("SELECT COUNT(*) FROM agencies WHERE id = ?", *input.AgencyID).Scan(&count)
		if err != nil || count == 0 {
			http.Error(w, "Invalid agency ID", http.StatusBadRequest)
			return
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to secure password", http.StatusInternalServerError)
		return
	}

	// Begin database transaction
	tx, err := c.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	uuid := c.generateUUID()

	// Create base client record
	client := models.Client{
		Name:      input.Name,
		Contact:   input.Contact,
		Email:     input.Email,
		UUID:      uuid,
		Type:      "landlord",
		Password:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}

	result, err := tx.Exec(`
        INSERT INTO clients (name, contact, email, password, type, uuid, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		client.Name, client.Contact, client.Email,
		client.Password, client.Type, client.UUID,
		client.CreatedAt, client.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			http.Error(w, "Email or contact already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create landlord account", http.StatusInternalServerError)
		return
	}

	clientID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get client ID", http.StatusInternalServerError)
		return
	}

	// Insert landlord-specific data
	_, err = tx.Exec(`
        INSERT INTO landlords (id, name, address, agency_id)
        VALUES (?, ?, ?, ?)`,
		clientID, input.Name, input.Address, input.AgencyID)

	if err != nil {
		http.Error(w, "Failed to create landlord profile", http.StatusInternalServerError)
		return
	}

	// Generate and send OTP
	if err := c.sendOTP(tx, clientID, input.Contact, now); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to complete registration", http.StatusInternalServerError)
		return
	}

	c.sendSuccessResponse(w, clientID, client.Name, client.Contact, "Landlord created successfully")
}
