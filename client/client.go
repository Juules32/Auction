package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	pb "github.com/Juules32/Auction/proto"
	"google.golang.org/grpc"
)

const serverAddr = "localhost:8080"

func main() {
	writeToLogAndTerminal("Starting new client...")

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewAuctionClient(conn)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter command:")
	for scanner.Scan() {

		input := scanner.Text()
		words := strings.Split(input, " ")

		if words[0] == "" {
			continue
		}

		switch strings.ToLower(words[0]) {
		case "bid":
			amount, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Println("Invalid bidding amount!")
				continue
			}
			bid(client, int32(amount))
		case "result":
			writeToLogAndTerminal("Client queries auction result")
			result(client)
		default:
			fmt.Println("Invalid command. Valid commands: 'bid', 'result'")
		}
	}
}

func bid(client pb.AuctionClient, amount int32) {
	bidResponse, err := client.Bid(context.Background(), &pb.BidRequest{Amount: amount})
	if err != nil {
		log.Fatalf("Error bidding: %v", err)
	}

	if bidResponse.Success {
		writeToLogAndTerminal("Client bid successfully: " + bidResponse.Message)
	} else {
		writeToLogAndTerminal("Client bid failed: " + bidResponse.Message)
	}
}

func result(client pb.AuctionClient) {
	resultResponse, err := client.Result(context.Background(), &pb.ResultRequest{})
	if err != nil {
		log.Fatalf("Error getting result: %v", err)
	}

	if resultResponse.IsActive {
		writeToLogAndTerminal("Highest Bid: " + strconv.Itoa(int(resultResponse.HighestBid)))

	} else {
		writeToLogAndTerminal("There is no active auction")
	}

}

func writeToLogAndTerminal(message string) {
	fmt.Println(message)

	f, err := os.OpenFile("log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)

	log.Println(message)

	defer f.Close()
}
