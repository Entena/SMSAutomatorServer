package main

import (
	"fmt"
	"microsms/config"
	"microsms/helpers"
	"microsms/models"
	"microsms/routes"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var server *gin.Engine
var startTime time.Time
var filterWG sync.WaitGroup
var filterResultChan chan helpers.FilterResult

func main() {
	var err error
	fmt.Printf("Loading config....\n")
	cfg := config.Load()
	fmt.Printf("Config loaded! Contents\n")
	cfg.Print()

	db, err := models.InitDB(cfg.Database.Path)
	fmt.Println("DB Initialized correctly:", db)
	if err != nil {
		panic("FAILED TO START DB")
	}

	// Initialize channel for filter results
	filterResultChan = make(chan helpers.FilterResult, 10)

	// Set global variables in helpers package with update function
	// Max 5 concurrent filter checks at a time
	helpers.SetFilterGlobals(&filterWG, filterResultChan, cfg.Filter.MaxConcurrent, cfg.Filter.APIURL)

	// Pass waitgroup to routes for goroutine spawning
	routes.SetFilterWaitGroup(&filterWG)

	// Start goroutine to handle filter results
	go helpers.HandleFilterResults()

	server = gin.Default()
	// converts into a single slash (/) when trying to match a route.
	server.RemoveExtraSlash = true

	// Optional: This handles cases where a user visits /health/ and you defined /health
	server.RedirectTrailingSlash = true
	startTime = time.Now()
	setupRoutes()

	server.Run(":8080") // start server on port 8080

	// Graceful shutdown: wait for all filter checks to complete
	fmt.Println("Server stopped, waiting for pending filter checks to complete...")
	close(filterResultChan) // Signal no more results coming
	filterWG.Wait()         // Wait for all goroutines to finish
	fmt.Println("All filter checks completed, exiting cleanly")
}

func GetHealth(c *gin.Context) {
	uptime := time.Since(startTime)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("API Healthy for %s", uptime)})
}

func setupRoutes() {
	apiGroup := server.Group("/api/v0")
	{
		apiGroup.POST("/create", routes.CreateSMSRequest)
		apiGroup.GET("/health", GetHealth)
		apiGroup.GET("/smsrequest", routes.GetSMSRequest)
		apiGroup.GET("/ready", routes.GetReadyToSendSMS)
		apiGroup.PATCH("/smsrequest", routes.UpdateSMSRequest)
	}

}
