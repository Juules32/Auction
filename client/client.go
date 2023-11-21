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
		fmt.Println("Bid successful:", bidResponse.Message)
	} else {
		fmt.Println("Bid failed:", bidResponse.Message)
	}
}

func result(client pb.AuctionClient) {
	resultResponse, err := client.Result(context.Background(), &pb.ResultRequest{})
	if err != nil {
		log.Fatalf("Error getting result: %v", err)
	}

	fmt.Println("Highest Bid:", resultResponse.HighestBid)
}
