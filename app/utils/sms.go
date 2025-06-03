package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	//"strings"
)

func stagingGetToken() (string, error) {
	url := "https://identity-service.intouchvas.io/auth/api-key"
	msisdn := "+254725035969"
	password := "29/nge/kim"
	credentials := msisdn + ":" + password
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+encodedCredentials)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	//body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse token response: %v", err)
	}

	return result.Token, nil
}

func SendSmsCode(msisdn string, message string) error {
	// Get the authentication token
	token, err := stagingGetToken()
	if err != nil {
		return fmt.Errorf("failed to get staging token: %v", err)
	}
	url := "https://sms-service.intouchvas.io/message/send/transactional"

	payload := map[string]interface{}{
		"msisdn":    msisdn,
		"message":   message,
		"sender_id": "AMARIALTD",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}
