package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
)

const serverAddr = "localhost:8080"

func main() {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	client := NewAuctionClient(conn)

	// Example: Bid
	bidAmount := int32(50)
	bidResponse, err := client.Bid(context.Background(), &BidRequest{Amount: bidAmount})
	if err != nil {
		log.Fatalf("Error bidding: %v", err)
	}

	if bidResponse.Success {
		fmt.Println("Bid successful:", bidResponse.Message)
	} else {
		fmt.Println("Bid failed:", bidResponse.Message)
	}

	// Example: Result
	resultResponse, err := client.Result(context.Background(), &ResultRequest{})
	if err != nil {
		log.Fatalf("Error getting result: %v", err)
	}

	fmt.Println("Highest Bid:", resultResponse.HighestBid)
}
