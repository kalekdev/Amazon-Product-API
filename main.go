package main

import (
	"context"
	"encoding/json"
	paapi5 "github.com/goark/pa-api"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	InfoLogger     *log.Logger
	ErrorLogger    *log.Logger
	apiCredentials []ApiCredential
)

type ApiCredential struct {
	PartnerTag         string `json:"partnerTag"`
	AccessKey          string `json:"accessKey"`
	SecretKey          string `json:"secretKey"`
	Client             paapi5.Client
	HttpClient         http.Client
	Jobs               JobsMap
	OnCooldown         bool
	RequestsSinceError int
	RequestsPerSecond  int `json:"requestsPerSecond"`
}

func init() {
	credentialFile, err := os.Open("credentials.json")
	defer credentialFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	jsonParser := json.NewDecoder(credentialFile)
	jsonParser.Decode(&apiCredentials)

	logFile, err := os.OpenFile("main.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	defer logFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	for i := 0; i < len(apiCredentials); i++ {
		apiCredentials[i].HttpClient = http.Client{Timeout: time.Second * 2}
		apiCredentials[i].Client = paapi5.New(paapi5.WithMarketplace(paapi5.LocaleUnitedStates)).CreateClient(
			apiCredentials[i].PartnerTag,
			apiCredentials[i].AccessKey,
			apiCredentials[i].SecretKey,
			paapi5.WithHttpClient(&apiCredentials[i].HttpClient))

		apiCredentials[i].Jobs = JobsMap{}
		apiCredentials[i].OnCooldown = false

		go runRequestLoop(&apiCredentials[i])

		/* For testing accounts
		dataChannel := make(chan ProductData, 1)
		jobs := []chan ProductData{dataChannel}
		apiCredentials[i].Jobs["B07GRM747Y"] = jobs
		apiCredentials[i].RequestData()

		<- dataChannel*/
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	handleRequests()
}
