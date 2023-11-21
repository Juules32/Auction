package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"

	pb "github.com/Juules32/Auction/proto"
	"google.golang.org/grpc"
)

var templateAuctionItemNames = []string{
	"Antique Vase",
	"Vintage Watch",
	"Rare Painting",
	"Collector's Coin",
	"Signed Book",
	"Classic Car",
	"Sports Memorabilia",
	"Designer Handbag",
	"Rare Stamp Collection",
	"Fine Wine",
}

// AuctionServer implements the Auction gRPC service
type AuctionServer struct {
	highestBid int32
	isActive   bool
	itemName   string
}

// Bid implements the Bid RPC method
func (s *AuctionServer) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {
	if !s.isActive {
		return &pb.BidResponse{Success: false, Message: "Auction inactive!"}, nil
	}

	if req.Amount > s.highestBid {
		s.highestBid = req.Amount
		return &pb.BidResponse{Success: true, Message: "Bid successful"}, nil
	}
	return &pb.BidResponse{Success: false, Message: "Bid too low"}, nil
}

// Result implements the Result RPC method
func (s *AuctionServer) Result(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error) {
	return &pb.ResultResponse{HighestBid: int32(s.highestBid)}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	server := grpc.NewServer()
	auctionServer := &AuctionServer{}

	pb.RegisterAuctionServer(server, auctionServer)
	fmt.Println("Auction server is running on :8080")

	go func() {
		err = server.Serve(listener)
		if err != nil {
			fmt.Println("Error serving gRPC:", err)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter command:")
	for scanner.Scan() {
		command := scanner.Text()

		switch strings.ToLower(command) {
		case "start":
			auctionServer.highestBid = int32(rand.Intn(100))
			auctionServer.itemName = randomAuctionItemName()
			auctionServer.isActive = true
			fmt.Printf("Started new auction for %s starting at %d dollars\n", auctionServer.itemName, auctionServer.highestBid)
		case "end":
			auctionServer.isActive = false
			fmt.Printf("Ended auction with winning bid %d\n", auctionServer.highestBid)
		default:
			fmt.Println("Invalid command. Valid commands: 'start', 'end'")
		}
	}
}

func randomAuctionItemName() string {
	return templateAuctionItemNames[rand.Intn(len(templateAuctionItemNames)-1)]
}
