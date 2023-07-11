# Amazon Product API
An Amazon product API that uses the [Product advertising API](https://webservices.amazon.com/paapi5/documentation/) to provide the latest non-cached data about an ASIN.

This API was used at [CloudSolve](https://cloudsolve.dev/) with several high rate limit Amazon associate accounts to provide clients with fast, high volume Amazon monitoring.

## Features
* Multiple PA API account details can be used in `credentials.json`
* API usage can be tracked and secured with API keys
* The worker queue architecture assigns requests optimally across accounts to avoid rate limits and reduce response times

## PA API Credentials
Amazon associates can get access to the PA API [here](https://webservices.amazon.com/paapi5/documentation/register-for-pa-api.html). Please note that you need a fairly extensive history as an associate to get reasonable rate limit allowances.
Fill in your PA API credentials in the json array in `credentials.json`. 

Each account's rate limit can be calculated [here](https://webservices.amazon.com/paapi5/documentation/troubleshooting/api-rates.html) and inserted in the `requestsPerSecond` property so the job queue best assigns incoming requests. Multiple high rate limit credentials are recommended for a good quality monitor.

## Setup
1. [Install Golang](https://go.dev/dl)
2. Complete `credentials.json` with your PA API details
3. Enter your MongoDB connection string in `mongo.js` to track API usage
4. Navigate to the source directory
5. Execute `go mod download` to install the required packages
6. Specify the desired port in the process `PORT` environment variable. By default, the API will be available at port `8080`
7. Execute `go run main.go` to start the service

Endpoint documentation can be seen at `localhost:PORT/documentation.html`