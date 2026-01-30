package constants

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
)

type OptInStatus string

const (
	OptInStatus_TRUE  OptInStatus = "true"
	OptInStatus_FALSE OptInStatus = "false"
	OptInStatus_ASK   OptInStatus = "ask"
	OptInStatus_ASKED OptInStatus = "asked"
)

type RequestAuthStatus string

const (
	RequestAuthStatus_READY     RequestAuthStatus = "ready"
	RequestAuthStatus_TO_VERIFY RequestAuthStatus = "to_verify"
	RequestAuthStatus_DENIED    RequestAuthStatus = "denied"
)

type RequestStatus string

const (
	RequestStatus_VERIFY_CHECK  RequestStatus = "verify_check"
	RequestStatus_READY_TO_SEND RequestStatus = "ready_to_send"
	RequestStatus_TAKEN         RequestStatus = "taken"
	RequestStatus_SENT          RequestStatus = "sent"
	RequestStatus_ERROR         RequestStatus = "error"
	RequestStatus_BLOCKED       RequestStatus = "blocked"
)

func IsValidOptInStatus(status string) bool {
	if status != string(OptInStatus_TRUE) && status != string(OptInStatus_FALSE) && status != string(OptInStatus_ASK) {
		return false
	}
	return true
}

func IsValidRequestAuthStatus(status string) bool {
	if status != string(RequestAuthStatus_READY) && status != string(RequestAuthStatus_TO_VERIFY) && status != string(RequestAuthStatus_DENIED) {
		return false
	}
	return true
}

func IsValidRequestStatus(status string) bool {
	if status != string(RequestStatus_VERIFY_CHECK) && status != string(RequestStatus_READY_TO_SEND) && status != string(RequestStatus_TAKEN) && status != string(RequestStatus_SENT) && status != string(RequestStatus_ERROR) && status != string(RequestStatus_BLOCKED) {
		return false
	}
	return true
}

func GetPhone(number string) (string, error) {
	if !IsValidPhone(number) {
		return "", fmt.Errorf("Error invalid phone number %s", number)
	}
	var phoneRegex = `^(?:\(\d{3}\)|\d{3})[-. ]?\d{3}[-. ]?\d{4}$`
	re := regexp.MustCompile(phoneRegex)
	match := re.FindStringSubmatch(number)

	// Extract components
	areaCode := match[0][:3] // Get area code
	centralOfficeCode := match[1]
	subscriberNumber := match[2]

	// Normalize to +1 (XXX) XXX-XXXX
	normalized := fmt.Sprintf("(%s)-%s-%s", areaCode, centralOfficeCode, subscriberNumber)

	return normalized, nil
}

// Helper for phone numbers
func IsValidPhone(number string) bool {
	// Define a regex pattern for phone numbers with an area code
	var phoneRegex = `^(?:\(\d{3}\)|\d{3})[-. ]?\d{3}[-. ]?\d{4}$`
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(number)
}

func GenerateCodePhrase() string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, 10)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "ERRORCODE"
		}
		b[i] = chars[n.Int64()]
	}

	return string(b)
}
