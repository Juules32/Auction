package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Auction represents the state of the auction.
type Auction struct {
	sync.Mutex
	highestBid int
}

// Response represents the response format for the Bid API.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var auction = &Auction{}

// BidHandler handles the Bid API.
func BidHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var bidAmount int
	err := json.NewDecoder(r.Body).Decode(&bidAmount)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	auction.Lock()
	defer auction.Unlock()

	if bidAmount > auction.highestBid {
		auction.highestBid = bidAmount
		response := Response{Success: true, Message: "Bid successful"}
		json.NewEncoder(w).Encode(response)
	} else {
		response := Response{Success: false, Message: "Bid too low"}
		json.NewEncoder(w).Encode(response)
	}
}

// ResultHandler handles the Result API.
func ResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	auction.Lock()
	defer auction.Unlock()

	result := struct {
		HighestBid int `json:"highestBid"`
	}{auction.highestBid}

	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/bid", BidHandler)
	http.HandleFunc("/result", ResultHandler)

	port := 8080
	fmt.Printf("Auction server is running on :%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
