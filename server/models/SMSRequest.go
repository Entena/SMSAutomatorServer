package models

import (
	"errors"
	"fmt"
	"microsms/constants"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SMSRequest definition
type SMSRequest struct {
	ID          uuid.UUID               `json:"id" gorm:"primary_key"`
	ToNumber    string                  `json:"to_number" gorm:"not null"`
	ToOptInID   uuid.UUID               `gorm:"index:toOpt_index;not null"`
	FromOptInID uuid.UUID               `gorm:"index:fromOpt_index;not null"`
	FromNumber  string                  `json:"from_number" gorm:"not null"`
	Status      constants.RequestStatus `json:"status"`
	Message     string                  `json:"message"`
	Created     int64                   `json:"created" gorm:"autoCreateTime"`

	// Define the association to OptIn
	ToOptIn   OptIn `gorm:"references:ID"`
	FromOptIn OptIn `gorm:"references:ID"`
}

// This should never happen, but hey if it does we can at least log something
func (smsrequest *SMSRequest) ToNumberF() string {
	f_phone, err := constants.GetPhone(smsrequest.ToNumber)
	if err != nil {
		fmt.Printf("ERROR COULD NOT RETURN A FORMATTED TO PHONE NUMBER!!! THIS REQUIRES MANUAL REMEDIATION")
		return smsrequest.ToNumber
	}
	return f_phone
}

func (smsrequest *SMSRequest) FromNumberF() string {
	f_phone, err := constants.GetPhone(smsrequest.ToNumber)
	if err != nil {
		fmt.Printf("ERROR COULD NOT RETURN A FORMATTED TO PHONE NUMBER!!! THIS REQUIRES MANUAL REMEDIATION")
		return smsrequest.ToNumber
	}
	return f_phone
}

// To String my struct
func (smsrequest SMSRequest) String() string {
	return fmt.Sprintf("SMSRequest{ ID: %s, Status: %s, Message: %s}", smsrequest.ID, smsrequest.Status, smsrequest.Message)
}

// Good old precreate hook to populate the id
func (smsrequest *SMSRequest) BeforeCreate(tx *gorm.DB) error {
	var err error
	var fNumberF, tNumberF string
	var fromOptIn, toOptIn *OptIn
	// Fetch our request auth to check status
	// validate our phone numbers
	if fNumberF, err = constants.GetPhone(smsrequest.FromNumber); err != nil {
		return fmt.Errorf("Error invalid from phone number %s", smsrequest.FromNumber)
	}
	if tNumberF, err = constants.GetPhone(smsrequest.ToNumber); err != nil {
		return fmt.Errorf("Error invalid to phone number %s", smsrequest.ToNumber)
	}
	smsrequest.ID = uuid.New()
	if fromOptIn, err = FindOrCreateOptIn(fNumberF); err != nil {
		return fmt.Errorf("Error with from opt in %s", err)
	}
	if toOptIn, err = FindOrCreateOptIn(tNumberF); err != nil {
		return fmt.Errorf("Error with to opt in %s", err)
	}
	smsrequest.ToOptInID = toOptIn.ID
	smsrequest.FromOptInID = fromOptIn.ID
	// Check all of our opt in statuses to determine initial request status
	if fromOptIn.Status == constants.OptInStatus_FALSE || toOptIn.Status == constants.OptInStatus_FALSE {
		smsrequest.Status = constants.RequestStatus_BLOCKED
	} else if fromOptIn.Status == constants.OptInStatus_ASK || toOptIn.Status == constants.OptInStatus_ASK {
		smsrequest.Status = constants.RequestStatus_VERIFY_CHECK
	} else if fromOptIn.Status == constants.OptInStatus_TRUE && toOptIn.Status == constants.OptInStatus_TRUE {
		smsrequest.Status = constants.RequestStatus_READY_TO_SEND
	}

	return nil
}

// Method to create new SMSRequest
func CreateSMSRequest(smsrequest *SMSRequest) error {
	if smsrequest.Message == "" {
		return fmt.Errorf("Error invalid message %s", smsrequest.Message)
	}
	if !constants.IsValidPhone((smsrequest.ToNumber)) {
		return fmt.Errorf("Error invalid to phone number %s", smsrequest.ToNumber)
	}
	if !constants.IsValidPhone(smsrequest.FromNumber) { // Get the raw numbers on purpose
		return fmt.Errorf("Error invalid from phone number %s", smsrequest.FromNumber)
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
func UpdateSMSRequest(id string, newStatus constants.RequestStatus) (*SMSRequest, error) {
	if !constants.IsValidRequestStatus(string(newStatus)) {
		return nil, fmt.Errorf("Error invalid status %s", newStatus)
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
	result := DB.Model(&SMSRequest{}).Where(&SMSRequest{Status: constants.RequestStatus_READY_TO_SEND}).Order("created ASC").First(&earliest)
	if result.Error != nil {
		fmt.Printf("Error finding ready to send SMS")
	}
	return &earliest, nil
}

// Based on the optin status
func UpdateSMSRequestStatusForNumber(optin *OptIn) {

}
