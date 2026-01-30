package routes

import (
	"fmt"
	"microsms/constants"
	"microsms/helpers"
	"microsms/models"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var filterWG *sync.WaitGroup

func SetFilterWaitGroup(wg *sync.WaitGroup) {
	filterWG = wg
}

func CreateSMSRequest(c *gin.Context) {
	var smsrequest models.SMSRequest
	if err := c.ShouldBindJSON(&smsrequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := models.CreateSMSRequest(&smsrequest); err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{"error": err.Error()})
		return
	}
	filterWG.Add(1)                                               // increment the waitgroup or else our app won't know of new potential goroutine
	go helpers.CheckSMSMessage(smsrequest.ID, smsrequest.Message) // Execute the CheckSMSMessage in parallel non blocking manner

	c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("SMSRequest Created %s", smsrequest.ID), "smsrequest": smsrequest})
}

// Simple method that checks ID exists and is valid UUID
func getIDCheckValid(c *gin.Context) (string, bool) {
	sms_id := c.Query("id")
	_, err := uuid.Parse(sms_id)
	if sms_id == "" || err != nil {
		fmt.Printf("ID is invalid %s", sms_id)
		//c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("ID is invalid %s", sms_id)})
		return sms_id, false
	}
	return sms_id, true
}

func GetSMSRequest(c *gin.Context) {
	sms_id, goOn := getIDCheckValid(c)
	if goOn == false {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("ID is invalid %s", sms_id)})
		return // exit we already set our c
	}
	smsrequest, err := models.GetSMSRequest(sms_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed finding SMSRequest %s", err)})
		return
	}
	if smsrequest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("SMSRequest ID %s not found", sms_id)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SMSRequest found", "smsrequest": smsrequest})
}

func UpdateSMSRequest(c *gin.Context) {
	var smsrequest *models.SMSRequest
	var smsupdate models.SMSRequest
	var err error
	sms_id, goOn := getIDCheckValid(c)
	if goOn == false {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("ID is invalid %s", sms_id)})
		return // exit
	}

	if err = c.ShouldBindJSON(&smsupdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("FAILED TO PARSE PAYLOAD %s", err)})
		return
	}
	smsrequest, err = models.UpdateSMSRequest(sms_id, smsupdate.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed updating SMSRequest %s", err)})
		return
	}
	if smsrequest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("SMSRequest ID %s not found", sms_id)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SMSRequest found", "smsrequest": smsrequest})
}

func GetReadyToSendSMS(c *gin.Context) {
	smsrequest, err := models.GetEarliestSMSRequest()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to find ready to send SMS %s", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("SMS Request %s ready to send", smsrequest.ID), "smsrequest": smsrequest})
}

func GetPhoneOptIn(c *gin.Context) {
	optInSearch := &models.OptIn{}
	var err error
	if err = c.ShouldBindBodyWithJSON(optInSearch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to parse payload %s", err)})
		return
	}
	if optInSearch.Number, err = constants.GetPhone(optInSearch.Number); err != nil {
		//Combo calls the check and gives us the necessary assignment
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid phone number %s", err)})
		return
	}
	optInSearch, err = models.GetOptIn(optInSearch.Number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Failed to find OptIn record %s", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("OptIn record found for %s", optInSearch.Number), "optin": optInSearch})
}

func GetReadyToAskOptIn(c *gin.Context) {
	optin, err := models.GetEarliestOptIn()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to find ready to ask OptIn %s", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("OptIn %s ready to ask", optin.Number), "optin": optin})
}

func UpdatePhoneOptIn(c *gin.Context) {
	var optinupdate models.OptIn
	var err error

	if err = c.ShouldBindJSON(&optinupdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("FAILED TO PARSE PAYLOAD %s", err)})
		return
	}
	if optinupdate.Number, err = constants.GetPhone(optinupdate.Number); err != nil {
		//Combo calls the check and gives us the necessary assignment
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid phone number %s", err)})
		return
	}
	// Check our optin against the message we got
	optIn, err := models.UpdateOptInAndRequestsIfAuthD(optinupdate.Number, optinupdate.Codeword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed updating OptIn %s", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OptIn updated", "optin": optIn})
}
