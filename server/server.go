package main

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

// AuctionServer implements the Auction gRPC service
type AuctionServer struct {
	highestBid int
}

// Bid implements the Bid RPC method
func (s *AuctionServer) Bid(ctx context.Context, req *BidRequest) (*BidResponse, error) {
	if req.Amount > s.highestBid {
		s.highestBid = req.Amount
		return &BidResponse{Success: true, Message: "Bid successful"}, nil
	}
	return &BidResponse{Success: false, Message: "Bid too low"}, nil
}

// Result implements the Result RPC method
func (s *AuctionServer) Result(ctx context.Context, req *ResultRequest) (*ResultResponse, error) {
	return &ResultResponse{HighestBid: int32(s.highestBid)}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	server := grpc.NewServer()
	RegisterAuctionServer(server, &AuctionServer{})

	fmt.Println("Auction server is running on :8080")

	err = server.Serve(listener)
	if err != nil {
		fmt.Println("Error serving gRPC:", err)
	}
}
