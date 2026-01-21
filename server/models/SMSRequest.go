package models

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB // Global DB pointer

// represent enum as strings cause I hate mapping this
type RequestStatus string

const (
	PAYMENT_OWED  RequestStatus = "payment_owed"
	READY_TO_SEND RequestStatus = "ready_to_send"
	TAKEN         RequestStatus = "taken"
	SENT          RequestStatus = "sent"
	ERROR         RequestStatus = "error"
	BLOCKED       RequestStatus = "blocked"
)

// Enums don't necessarily restrict input
func IsValidStatus(status string) bool {
	if status != "payment_owed" && status != "ready_to_send" && status != "taken" && status != "sent" && status != "error" && status != "blocked" {
		return false
	}
	return true
}

// SMSRequest definition
type SMSRequest struct {
	ID      uuid.UUID     `json:"id" gorm:"primary_key"`
	Number  string        `json:"number" gorm:"not null"`
	Status  RequestStatus `json:"status"`
	Message string        `json:"message"`
	Created int64         `json:"created" gorm:"autoCreateTime"`
}

// To String my struct
func (smsrequest SMSRequest) String() string {
	return fmt.Sprintf("SMSRequest{ ID: %s, Status: %s, Message: %s}", smsrequest.ID, smsrequest.Status, smsrequest.Message)
}

// Good old precreate hook to populate the id
func (smsrequest *SMSRequest) BeforeCreate(tx *gorm.DB) error {
	smsrequest.ID = uuid.New()
	smsrequest.Status = PAYMENT_OWED
	return nil
}

// Helper for phone numbers
func IsValidPhone(number string) bool {
	// Define a regex pattern for phone numbers with an area code
	var phoneRegex = `^(?:\(\d{3}\)|\d{3})[-. ]?\d{3}[-. ]?\d{4}$`
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(number)
}

// Method to create new SMSRequest
func CreateSMSRequest(smsrequest *SMSRequest) error {
	if smsrequest.Message == "" {
		return errors.New(fmt.Sprintf("Error invalid message %s", smsrequest.Message))
	}
	if !IsValidPhone((smsrequest.Number)) {
		return errors.New(fmt.Sprintf("Error invalid phone number %s", smsrequest.Number))
	}
	result := DB.Create(smsrequest)
	if result.Error != nil {
		fmt.Println("Error creating SMS Request:", result.Error)
		return result.Error
	}
	fmt.Println("Create new SMS Request: ", smsrequest)
	return nil
}

// Update SMSRequest with new status
func UpdateSMSRequest(id string, newStatus RequestStatus) (*SMSRequest, error) {
	if !IsValidStatus(string(newStatus)) {
		return nil, errors.New(fmt.Sprintf("Error invalid status %s", newStatus))
	}
	smsrequest, err := GetSMSRequest(id)
	if err != nil {
		fmt.Printf("ERROR UPDATING SMSREQUEST %s, %s", id, err)
		return nil, err
	}
	if smsrequest == nil {
		fmt.Printf("ERROR COULD NOT FIND SMSREQUEST TO UPDATE %s", id)
		return nil, errors.New("COULD NOT FIND RECORD")
	}
	smsrequest.Status = newStatus
	err = DB.Save(smsrequest).Error
	if err != nil {
		fmt.Printf("ERROR SAVING UPDATE TO DB %s", err)
		return nil, err
	}
	return smsrequest, nil
}

// Get the single SMSRequest or return nil
func GetSMSRequest(id string) (*SMSRequest, error) {
	fmt.Printf("GET SMSREQUEST BY ID %s\n", id)
	smsrequest := SMSRequest{}
	uid := uuid.MustParse(id)
	result := DB.First(&smsrequest, uid)
	if result.Error != nil {
		fmt.Printf("ERROR FINDING SMSREQUEST %s\n", result.Error)
		return nil, result.Error
	}
	return &smsrequest, nil
}

// Get the earliest ready_to_send record
func GetEarliestSMSRequest() (*SMSRequest, error) {
	fmt.Printf("GET EARLIEST SMSREQUEST")
	var earliest SMSRequest
	result := DB.Model(&SMSRequest{}).Where(&SMSRequest{Status: READY_TO_SEND}).Order("created ASC").First(&earliest)
	if result.Error != nil {
		fmt.Printf("Error finding ready to send SMS")
	}
	return &earliest, nil
}

// Create DB connection
// TODO use psql in the future and set up support for it
func InitDB(dbPath string) (*gorm.DB, error) {
	fmt.Printf("Initing DB with dbPath %s\n", dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	DB = db
	DB.AutoMigrate(&SMSRequest{}) // create the reqeuest table
	fmt.Println("DB Inited")
	return DB, nil
}
