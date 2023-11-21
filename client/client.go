package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const serverURL = "http://localhost:8080"

// Bid sends a bid to the auction server.
func Bid(amount int) (bool, string, error) {
	url := serverURL + "/bid"

	body, err := json.Marshal(amount)
	if err != nil {
		return false, "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return false, "", err
	}

	return response.Success, response.Message, nil
}

// Result retrieves the result of the auction from the server.
func Result() (int, error) {
	url := serverURL + "/result"

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		HighestBid int `json:"highestBid"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, err
	}

	return result.HighestBid, nil
}

func main() {
	// Example: Bid
	bidAmount := 50
	success, message, err := Bid(bidAmount)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if success {
		fmt.Println("Bid successful:", message)
	} else {
		fmt.Println("Bid failed:", message)
	}

	// Example: Result
	highestBid, err := Result()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Highest Bid:", highestBid)
}
