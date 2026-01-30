package models

import (
	"fmt"
	"microsms/constants"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OptIn struct {
	ID       uuid.UUID             `json:"id" gorm:"primary_key"`
	Number   string                `json:"number" gorm:"uniqueIndex"`
	Codeword string                `json:"codeword"`                // the codeword we sent in our opt in msg
	Status   constants.OptInStatus `json:"contact" gorm:"not null"` // true means they opted in
	Created  int64                 `json:"created" gorm:"autoCreateTime"`
	Updated  int64                 `json:"updated" gorm:"autoUpdateTime"`
}

// Find or create an optin record for a phone number
func FindOrCreateOptIn(number string) (*OptIn, error) {
	var foundOptIn OptIn
	result := DB.First(&foundOptIn, &OptIn{Number: number})
	switch result.RowsAffected { // Create records if they don't exist
	case 0:
		// No optin found, so set status to ASK
		newOptIn := OptIn{
			Number:   number,
			Status:   constants.OptInStatus_ASK,
			Codeword: constants.GenerateCodePhrase(), // Generate a unique code for our pass
		}
		err := DB.Create(&newOptIn).Error
		if err != nil {
			return nil, err
		}
		return &newOptIn, nil
	}
	return &foundOptIn, nil
}

// Prevent duplicate optin records
func (optin *OptIn) BeforeCreate(tx *gorm.DB) error {
	optin.ID = uuid.New()
	// Make sure we aren't creating a naughty boi request
	var foundOptIn OptIn
	tx.First(&foundOptIn, &OptIn{Number: optin.Number})
	switch tx.RowsAffected { // Create records if they don't exist
	case 0:
		// No optin found, all good
		break
	default:
		return fmt.Errorf("Duplicate optin found for number %s", optin.Number)
	}
	return nil
}

// Check if the response contains the codeword
func (optin *OptIn) ContainsCodeword(response string) bool {
	response = strings.ToLower(response)
	codeword := strings.ToLower(optin.Codeword)
	return strings.Contains(response, codeword)
}

// Get the opt in record for a phone
func GetOptIn(phone string) (*OptIn, error) {
	var optin OptIn

	err := DB.First(&optin, &OptIn{Number: phone}).Error
	if err != nil {
		return nil, err
	}
	return &optin, nil
}

func UpdateOptInFromAskTo(number string, newStatus constants.OptInStatus) (*OptIn, error) {
	var optin *OptIn
	var err error
	if !constants.IsValidOptInStatus(string(newStatus)) || newStatus == constants.OptInStatus_ASK {
		return nil, fmt.Errorf("%s status is not a valid opt in status", newStatus)
	}
	if optin, err = GetOptIn(number); err != nil {
		return nil, fmt.Errorf("Error fetching option %s", err)
	}
	optin.Status = newStatus
	err = DB.Save(optin).Error
	if err != nil {
		return nil, fmt.Errorf("Error saving to db %s", err)
	}
	return optin, nil
}

// So let's just make this dumb, meaning you can hit it multiple times and it will swap status
// based on the message you gave it. The trick is we will rely on the codeword, it's silly
// and not the most secure but it's a fine way to give toggable optin/optout. Note that this
// design implies that the phone will be responding back to the API with all of it's texts
// so good thing we chose Go
func UpdateOptInAndRequestsIfAuthD(phone string, response string) (*OptIn, error) {
	optin, err := GetOptIn(phone)
	if err != nil {
		return nil, err
	}
	// Check auth
	if optin.ContainsCodeword(response) {
		var smsrequests []SMSRequest
		switch optin.Status {
		case constants.OptInStatus_TRUE:
			// They are already opted in, so opt them out
			optin.Status = constants.OptInStatus_FALSE
		case constants.OptInStatus_FALSE, constants.OptInStatus_ASK:
			// They are opting in or toggling
			optin.Status = constants.OptInStatus_TRUE
		}
		// After we toggle our case trigger our SMSRequest update logic
		switch optin.Status {
		case constants.OptInStatus_TRUE: // we are now true so update all verify check status where we are one of the numbers
			DB.Where(&SMSRequest{Status: constants.RequestStatus_VERIFY_CHECK, FromNumber: optin.Number}).Or(&SMSRequest{Status: constants.RequestStatus_VERIFY_CHECK, ToNumber: optin.Number}).Find(&smsrequests)
			for i := 0; i < len(smsrequests); i++ {
				if smsrequests[i].FromOptIn.Status == constants.OptInStatus_TRUE && smsrequests[i].ToOptIn.Status == constants.OptInStatus_TRUE {
					smsrequests[i].Status = constants.RequestStatus_READY_TO_SEND // if they are both opted in then mark this as ready
				}
			}
			DB.Save(&smsrequests) // Bulk save records
		case constants.OptInStatus_FALSE: // Get all the ready to send/verify checks so we can now update them to be denied
			DB.Where(&SMSRequest{Status: constants.RequestStatus_READY_TO_SEND, FromNumber: optin.Number}).Or(&SMSRequest{Status: constants.RequestStatus_VERIFY_CHECK, FromNumber: optin.Number}).Or(&SMSRequest{Status: constants.RequestStatus_READY_TO_SEND, ToNumber: optin.Number}).Or(&SMSRequest{Status: constants.RequestStatus_VERIFY_CHECK, ToNumber: optin.Number}).Updates(SMSRequest{Status: constants.RequestStatus_BLOCKED})
		}

	}
	err = DB.Save(optin).Error
	if err != nil {
		return nil, err
	}
	return optin, nil
}

// Get the earliest opt in we haven't asked yet
func GetEarliestOptIn() (*OptIn, error) {
	var earliest OptIn
	result := DB.Model(&OptIn{}).Where(&OptIn{Status: constants.OptInStatus_ASK}).Order("created asc").Limit(1).First(&earliest)
	if result.Error != nil {
		return nil, result.Error
	}
	return &earliest, nil
}
