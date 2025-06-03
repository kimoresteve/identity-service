package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	middleware "github.com/kimoresteve/identity-service/app/middlewares"
	"github.com/kimoresteve/identity-service/app/models"
	"github.com/kimoresteve/identity-service/app/utils"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

// Status returns the health status of the API server.
// @Summary Server status check
// @Description Health-check endpoint for the API
// @Tags status
// @Produce json
// @Success 200 {string} string "OK"
// @Router / [get]
func (c *Controller) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

// Verify verifies a client using an OTP code.
// @Summary Verify client
// @Description Verifies a client's account using their OTP
// @Tags Client
// @Accept json
// @Produce json
// @Param verification body models.Verify  true "OTP Verification Input"
// @Success 200 {object} models.Response "Client verified successfully"
// @Failure 400 {string} string "Invalid request or bad OTP"
// @Failure 404 {string} string "OTP not found"
// @Router /auth/verify [post]
func (c *Controller) Verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		OTP      string `json:"otp"`
		ClientID int    `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

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

	var dbOtp string
	var expiresAt time.Time

	err = tx.QueryRow(`
		SELECT otp, expires_at 
		FROM otp_codes 
		WHERE client_id = ? AND purpose = 'activation'
		ORDER BY expires_at DESC 
		LIMIT 1
	`, body.ClientID).Scan(&dbOtp, &expiresAt)

	if err != nil {
		http.Error(w, "OTP not found", http.StatusNotFound)
		return
	}
	if dbOtp != body.OTP {
		http.Error(w, "Invalid OTP", http.StatusUnauthorized)
		return
	}
	if time.Now().After(expiresAt) {
		http.Error(w, "OTP expired", http.StatusUnauthorized)
		return
	}
	_, err = tx.Exec(`
		UPDATE clients 
		SET is_verified = 1, updated_at = ? 
		WHERE id = ?
	`, time.Now(), body.ClientID)

	if err != nil {
		http.Error(w, "Failed to verify client", http.StatusInternalServerError)
		return
	}
	_, err = tx.Exec(`DELETE FROM otp_codes WHERE client_id = ?`, body.ClientID)
	if err != nil {
		http.Error(w, "Failed to delete OTP", http.StatusInternalServerError)
		return
	}
	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit verification", http.StatusInternalServerError)
		return
	}
	response := models.Response{
		Success: true,
		Message: "Client verified successfully.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Login logs a landlord into the system.
// @Summary Client login
// @Description Logs in a client by validating contact and password
// @Tags Client
// @Accept json
// @Produce json
// @Param credentials body models.LoginInput true "Login credentials"
// @Success 200 {object} models.Response "Login successful with token"
// @Failure 401 {string} string "Invalid credentials or unverified landlord"
// @Failure 404 {string} string "Landlord not found"
// @Router /auth/login [post]
func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input models.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
	}

	defer r.Body.Close()

	tx, err := c.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	row := tx.QueryRow(`SELECT id, name, contact, type, email, password, created_at, updated_at, is_verified FROM clients WHERE contact = ?`, input.Contact)

	var landlord models.Client
	err = row.Scan(&landlord.ID, &landlord.Name, &landlord.Contact, &landlord.Type, &landlord.Email, &landlord.Password, &landlord.CreatedAt, &landlord.UpdatedAt, &landlord.IsVerified)
	if err != nil {
		http.Error(w, "Landlord not found", http.StatusNotFound)
		return
	}

	if !landlord.IsVerified {
		http.Error(w, "Landlord not verified", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(landlord.Password), []byte(input.Password)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	token, err := middleware.GenerateJWT(uint(landlord.ID))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := models.Response{
		Success: true,
		Message: "Login successful",
		Data: map[string]interface{}{
			"id":      landlord.ID,
			"name":    landlord.Name,
			"contact": landlord.Contact,
			"token":   token,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

// ForgotPassword sends a reset OTP code to the Client.
// @Summary Forgot password
// @Description Sends an OTP to the client to reset password
// @Tags Client
// @Accept json
// @Produce json
// @Param forgotPassword body models.ForgotPasswordInput true "Contact Information"
// @Success 200 {object} models.Response "OTP sent successfully"
// @Failure 400 {string} string "Invalid request payload"
// @Failure 404 {string} string "Client not found"
// @Router /auth/forgot-password [post]
func (c *Controller) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body models.ForgotPasswordInput
	now := time.Now()

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Start transaction
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

	// Find client by contact
	row := tx.QueryRow(`SELECT id, name, contact, email FROM clients WHERE contact = ?`, body.Contact)

	var result struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Contact string `json:"contact"`
		Email   string `json:"email"`
	}

	err = row.Scan(&result.ID, &result.Name, &result.Contact, &result.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Client not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Generate OTP
	otp, err := utils.GenerateOTP()
	if err != nil {
		http.Error(w, "Failed to generate OTP", http.StatusInternalServerError)
		return
	}

	// Insert OTP (still within the same transaction)
	expiresAt := now.Add(15 * time.Minute)
	_, err = tx.Exec(`
        INSERT INTO otp_codes (client_id, otp, expires_at, purpose)
        VALUES (?, ?, ?, ?)`,
		result.ID, otp, expiresAt, "reset")

	if err != nil {
		http.Error(w, "Failed to store OTP", http.StatusInternalServerError)
		fmt.Printf("Error storing OTP: %s", err.Error())
		return
	}

	// Send SMS
	message := fmt.Sprintf("Hello, Your Reset Code is %s. It expires in 15 minutes.", otp)
	if err := utils.SendSmsCode(result.Contact, message); err != nil {
		http.Error(w, "Failed to send OTP", http.StatusInternalServerError)
		fmt.Printf("Error sending SMS: %s", err.Error())
		return
	}

	// Commit transaction only at the end
	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to complete operation", http.StatusInternalServerError)
		fmt.Printf("Error committing transaction: %s", err.Error())
		return
	}

	// Send success response
	response := models.Response{
		Success: true,
		Message: "Reset Password code has been sent",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// ResetPassword resets the Client's password using an OTP.
// @Summary Reset password
// @Description Resets a client's password using a valid OTP code
// @Tags Client
// @Accept json
// @Produce json
// @Param resetPassword body models.ResetPasswordInput true "Reset password request"
// @Success 200 {object} models.Response "Password reset successful"
// @Failure 400 {string} string "Invalid request payload"
// @Failure 401 {string} string "Invalid or expired OTP"
// @Router /auth/reset-password [post]
func (c *Controller) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body models.ResetPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
	}
	defer r.Body.Close()

	tx, err := c.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
	}
	defer tx.Rollback()

	var dbOtp string
	var expiresAt time.Time

	err = tx.QueryRow(`
		SELECT otp, expires_at 
		FROM otp_codes 
		WHERE client_id = ? AND purpose = 'reset'
		ORDER BY expires_at DESC 
		LIMIT 1
	`, body.ID).Scan(&dbOtp, &expiresAt)

	if err != nil {
		http.Error(w, "OTP not found", http.StatusNotFound)
		return
	}
	if dbOtp != body.OTP {
		http.Error(w, "Invalid OTP", http.StatusUnauthorized)
		return
	}
	if time.Now().After(expiresAt) {
		http.Error(w, "OTP expired", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to secure password", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`UPDATE clients SET password = ?, updated = ?,  WHERE id = ? `, hashedPassword, time.Now(), body.ID)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
	}

	_, err = tx.Exec(`DELETE FROM otp_codes WHERE client_id = ?`, body.ID)
	if err != nil {
		http.Error(w, "Failed to delete OTP", http.StatusInternalServerError)
		return
	}
	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit verification", http.StatusInternalServerError)
		return
	}

	response := models.Response{
		Success: true,
		Message: "Password reset successfully",
	}
	json.NewEncoder(w).Encode(response)

}
func (c *Controller) AddUser(w http.ResponseWriter, r *http.Request) {

}

type AgencyInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Contact  string `json:"contact"`
	Password string `json:"password"`
	Address  string `json:"address"`
	TaxID    string `json:"tax_id"`
	LogoURL  string `json:"logo_url,omitempty"`
}

type LandlordInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Contact  string `json:"contact"`
	Password string `json:"password"`
	Address  string `json:"address"`
	AgencyID *int   `json:"agency_id,omitempty"`
}

// Helper function to generate and send OTP
func (c *Controller) sendOTP(tx *sql.Tx, clientID int64, contact string, now time.Time) error {
	otp, err := utils.GenerateOTP()
	if err != nil {
		return fmt.Errorf("Failed to generate OTP")
	}

	expiresAt := now.Add(15 * time.Minute)
	_, err = tx.Exec(`
        INSERT INTO otp_codes (client_id, otp, expires_at)
        VALUES (?, ?, ?)`,
		clientID, otp, expiresAt)

	if err != nil {
		return fmt.Errorf("Failed to store OTP")
	}

	message := fmt.Sprintf("Hello, Your OTP is %s. It expires in 15 minutes.", otp)
	if err := utils.SendSmsCode(contact, message); err != nil {
		return fmt.Errorf("Failed to send OTP")
	}

	return nil
}

// Helper function to send success response
func (c *Controller) sendSuccessResponse(w http.ResponseWriter, clientID int64, name, contact, message string) {
	response := models.Response{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"id":      clientID,
			"name":    name,
			"contact": contact,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// Helper function to generate UUID
func (c *Controller) generateUUID() string {
	// TODO: Use a proper UUID library like github.com/google/uuid
	// return uuid.New().String()

	// Temporary placeholder - replace with proper UUID generation
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}
