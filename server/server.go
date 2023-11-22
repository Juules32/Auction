package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

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
	minimumBid int32
	isActive   bool
	itemName   string
}

func (s *AuctionServer) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {

	s.highestBid = req.HighestBid
	s.isActive = req.IsActive
	s.itemName = req.ItemName
	s.minimumBid = req.MinimumBid
	return &pb.UpdateResponse{}, nil
}

// Bid implements the Bid RPC method
func (s *AuctionServer) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {

	if !s.isActive {
		return &pb.BidResponse{Success: false, Message: "Auction inactive!"}, nil
	}

	if req.Amount > s.highestBid && req.Amount >= s.minimumBid {
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
	auctionServer := &AuctionServer{}
	stop := false
	for !stop {
		serverListener, err := net.Listen("tcp", "localhost:5050")
		if err == nil {
			serverListener.Close()
			break
		}
		fmt.Println(serverListener)

		time.Sleep(time.Second * 2)
	}

	serverListener, err := net.Listen("tcp", "localhost:5050")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer serverListener.Close()
	fmt.Println(serverListener)

	listener, err := net.Listen("tcp", "localhost:8080")

	server := grpc.NewServer()

	pb.RegisterAuctionServer(server, auctionServer)
	go func() {
		err = server.Serve(listener)
		if err != nil {
			fmt.Println("sss", err)
		}
		err = server.Serve(serverListener)
		if err != nil {
			fmt.Println("ddd", err)
		}
	}()

	go func() {
		conn, err := grpc.Dial("localhost:5050", grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			stop = true
		}
		client := pb.NewReplicationClient(conn)

		updateResponse, err := client.Update(context.Background(), &pb.UpdateRequest{HighestBid: auctionServer.highestBid, MinimumBid: auctionServer.minimumBid, IsActive: auctionServer.isActive, ItemName: auctionServer.itemName})
		fmt.Print(updateResponse)
		conn.Close()
		time.Sleep(time.Second * 2)

	}()

	fmt.Println("Auction server is running on localhost:8080")

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter command:")
	for scanner.Scan() {
		command := scanner.Text()

		switch strings.ToLower(command) {
		case "start":
			auctionServer.highestBid = 0
			auctionServer.minimumBid = int32(rand.Intn(100))
			auctionServer.itemName = randomAuctionItemName()
			auctionServer.isActive = true
			fmt.Printf("Started new auction for %s starting at %d dollars\n", auctionServer.itemName, auctionServer.minimumBid)
		case "end":
			auctionServer.isActive = false
			fmt.Printf("Ended auction with winning bid %d\n", auctionServer.highestBid)
		case "crash":
			fmt.Println("Stopping gRPC server...")
			server.GracefulStop()
			return
		case "print":
			fmt.Print(auctionServer)
		case "bootup":
			auctionServer = &AuctionServer{}

			listener, err = net.Listen("tcp", "localhost:8080")
			if err != nil {
				fmt.Println("Error starting server:", err)
				return
			}

			server = grpc.NewServer()

			pb.RegisterAuctionServer(server, auctionServer)
			go func() {
				err := server.Serve(listener)
				if err != nil {
					fmt.Println("Error serving gRPC:", err)
				}
			}()
			fmt.Println("Auction server is running on localhost:8080")

		default:
			fmt.Println("Invalid command. Valid commands: 'start', 'end'")
		}
	}
}

func randomAuctionItemName() string {
	return templateAuctionItemNames[rand.Intn(len(templateAuctionItemNames)-1)]
}
