package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func getProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	apiKey := r.URL.Query().Get("key")
	usage, err := getUsage(apiKey)
	if err != nil {
		http.Error(w, "Invalid API key.", http.StatusUnauthorized)
		return
	}

	asin := strings.TrimPrefix(r.URL.Path, "/product/")

	if len(asin) != 10 {
		http.Error(w, "Invalid ASIN.", 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Usage", strconv.Itoa(usage+1))

	dataChannel := make(chan ProductData, 1)
	assignDataRequest(asin, dataChannel)

	productData := <-dataChannel
	response, _ := json.Marshal(productData)

	fmt.Fprint(w, string(response))

	go incrementUsage(apiKey)
}

func getUsageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	apiKey := r.URL.Query().Get("key")

	usage, err := getUsage(apiKey)

	if err != nil {
		http.Error(w, "Invalid API key.", 401)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "{\n   \"key\": \""+apiKey+"\",\n   \"requests\": "+strconv.Itoa(usage)+"\n}")
}

func handleRequests() {
	port := ":" + os.Getenv("PORT")

	if port == ":" {
		port = ":8080"
	}

	http.HandleFunc("/product/", getProductHandler)
	http.HandleFunc("/usage", getUsageHandler)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.ListenAndServe(port, nil)
}
