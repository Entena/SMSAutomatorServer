package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"microsms/models"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

/**
A simple package to help querying the SMS filter API. The filter API runs qwen and gives a simple
boolean to indicate if a message is "blocked"
**/

var filterAPIURL string

var filterWG *sync.WaitGroup
var filterResultChan chan FilterResult
var semaphore chan struct{}

type FilterResult struct {
	SMSID   uuid.UUID
	Blocked bool
	Err     error
}

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

// SetFilterGlobals sets the waitgroup, channel, and API URL from main
func SetFilterGlobals(wg *sync.WaitGroup, ch chan FilterResult, maxConcurrent int, apiURL string) {
	filterWG = wg
	filterResultChan = ch
	semaphore = make(chan struct{}, maxConcurrent)
	filterAPIURL = apiURL
}

// HandleFilterResults processes the results from the filter API channel
func HandleFilterResults() {
	for result := range filterResultChan {
		fmt.Printf("Handling filter result for SMS ID: %s\n", result.SMSID)

		if result.Err != nil {
			fmt.Printf("Error filtering SMS %s: %s\n", result.SMSID, result.Err)
			_, err := models.UpdateSMSRequest(result.SMSID.String(), models.ERROR)
			if err != nil {
				fmt.Printf("Failed to update SMS %s to ERROR status: %s\n", result.SMSID, err)
			}
			continue
		}

		if result.Blocked {
			fmt.Printf("SMS %s was blocked by filter\n", result.SMSID)
			_, err := models.UpdateSMSRequest(result.SMSID.String(), models.BLOCKED)
			if err != nil {
				fmt.Printf("Failed to update SMS %s to BLOCKED status: %s\n", result.SMSID, err)
			}
		} else {
			fmt.Printf("SMS %s passed filter, marking as READY_TO_SEND\n", result.SMSID)
			_, err := models.UpdateSMSRequest(result.SMSID.String(), models.READY_TO_SEND)
			if err != nil {
				fmt.Printf("Failed to update SMS %s to READY_TO_SEND status: %s\n", result.SMSID, err)
			}
		}
	}
}

// CheckSMSMessage checks the message and sends result to channel (runs in goroutine)
func CheckSMSMessage(smsID uuid.UUID, message string) {
	defer filterWG.Done() // WG will decrement on function finish

	// Acquire semaphore slot (blocks if max concurrent reached)
	semaphore <- struct{}{}
	defer func() { <-semaphore }() // Release slot when done

	blocked, err := checkSMSMessage(message)
	filterResultChan <- FilterResult{
		SMSID:   smsID,
		Blocked: blocked,
		Err:     err,
	}
}

// Fires Filter API request and returns the bool value
func checkSMSMessage(message string) (bool, error) {
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
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Filter API returned non-200 status: %d", resp.StatusCode)
		goto ERROR
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		goto ERROR
	}
	fmt.Printf("Safety API returned %s", body)
	err = json.Unmarshal(body, &smsResponse)

	return smsResponse.Blocked, nil
ERROR:
	return false, err
}
