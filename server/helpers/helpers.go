package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

/**
A simple package to help querying the SMS filter API. The filter API runs qwen and gives a simple
boolean to indicate if a message is "blocked"
**/

var filterAPIURL string = "http://192.168.8.100:8000/api/v0/filter/sms"

// SMSFilterRequest represents the request payload structure
type SMSFilterRequest struct {
	SMS string `json:"sms"`
}

// SMSFilterResponse represents the response structure from the API
type SMSResponse struct {
	Blocked            bool     `json:"blocked"`
	Reason             string   `json:"reason"`
	IncludedCategories []string `json:"included_categories"`
	ExcludedCategories []string `json:"excluded_categories"`
}

// Fires Filter API request and returns the bool value
func CheckSMSMessage(message string) (bool, error) {
	var smsResponse SMSResponse
	var body []byte
	var resp *http.Response
	smsFilter := SMSFilterRequest{SMS: message}
	payload, err := json.Marshal(smsFilter)
	if err != nil {
		goto ERROR
	}
	resp, err = http.Post(filterAPIURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		goto ERROR
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		goto ERROR
	}

	err = json.Unmarshal(body, &smsResponse)

	return smsResponse.Blocked, nil
ERROR:
	return false, err
}
