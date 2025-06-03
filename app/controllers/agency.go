package controllers

import (
	"fmt"
	"github.com/kimoresteve/identity-service/app/models"
	"github.com/kimoresteve/identity-service/app/utils"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

// Updated input models }

// RegisterAgency creates an agency account.
// @Summary Agency Account Creation
// @Description Registers a new agency with contact, email, password, address, and tax ID
// @Tags Agency
// @Accept json
// @Produce json
// @Param agency body AgencyInput true "Agency Registration Input"
// @Success 201 {object} map[string]interface{} "Agency created successful"
// @Failure 400 {string} string "Invalid request or missing fields"
// @Failure 409 {string} string "Email or contact already exists"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/register/agency [post]
func (c *Controller) RegisterAgency(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input AgencyInput
	if err := utils.DecodeAndValidateJSONBody(w, r, &input); err != nil {
		var apiErr *utils.APIError
		if errors.As(err, &apiErr) {
			http.Error(w, apiErr.Message, apiErr.Status)
		}
		return
	}

	// Validate required fields for agency
	if input.Name == "" || input.Contact == "" || input.Password == "" || input.Email == "" {
		http.Error(w, "Name, Email, contact and password are required", http.StatusBadRequest)
		return
	}

	//if input.Address == "" || input.TaxID == "" {
	//	http.Error(w, "Address and Tax ID are required for agencies", http.StatusBadRequest)
	//	return
	//}

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
		Type:      "agency",
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
		http.Error(w, "Failed to create agency account", http.StatusInternalServerError)
		fmt.Printf("Failed to create agency account: %v\n", err)
		return
	}

	clientID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get client ID", http.StatusInternalServerError)
		return
	}

	// Insert agency-specific data
	_, err = tx.Exec(`
        INSERT INTO agencies (id, name, address, tax_id, logo_url)
        VALUES (?, ?, ?, ?, ?)`,
		clientID, input.Name, input.Address, input.TaxID, input.LogoURL)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") && strings.Contains(err.Error(), "tax_id") {
			http.Error(w, "Tax ID already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create agency profile", http.StatusInternalServerError)
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

	c.sendSuccessResponse(w, clientID, client.Name, client.Contact, "Agency created successfully")
}

func (c *Controller) RegisterLandlordAgency(w http.ResponseWriter, r *http.Request) {
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
	}
}
