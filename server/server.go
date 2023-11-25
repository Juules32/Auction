package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/Juules32/Auction/proto"
	"google.golang.org/grpc"
)

// Template auction items for flavor
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
	"Ancient Manuscript",
	"Space Exploration Artifact",
	"Modern Sculpture",
	"Historical Document",
	"Exotic Gemstone",
	"Musical Instrument",
	"Scientific Invention Prototype",
	"Vintage Camera",
	"Luxury Yacht",
	"Fossilized Dinosaur Bone",
	"Exclusive Fashion Accessory",
	"Limited Edition Chess Set",
	"Artificial Intelligence Robot",
	"Astronomical Telescope",
	"Rare Whisky Collection",
	"Japanese Samurai Sword",
	"Gold Bullion Bars",
	"High-End Audio System",
	"Rare Gemstone Jewelry",
}

// AuctionServer implements the Auction gRPC service
type AuctionServer struct {
	HighestBid int32  `json:"HighestBid"`
	MinimumBid int32  `json:"MinimumBid"`
	IsActive   bool   `json:"IsActive"`
	ItemName   string `json:"ItemName"`
}

// Struct used to save and update information about the auction
var auctionServer *AuctionServer
var serverListener net.Listener
var mut sync.Mutex

// Bid implements the Bid RPC method
func (s *AuctionServer) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {
	mut.Lock()
	defer mut.Unlock()

	// There must be an active auction
	if !s.IsActive {
		return &pb.BidResponse{Success: false, Message: "Auction inactive!"}, nil
	}

	// The amount must be higher than the highest bid
	// or higher or equal to the minimum bid
	if req.Amount <= s.HighestBid || req.Amount < s.MinimumBid {
		return &pb.BidResponse{Success: false, Message: "Bid too low"}, nil
	}

	s.HighestBid = req.Amount
	handleBackupReplicas(serverListener)
	return &pb.BidResponse{Success: true, Message: "Bid successful"}, nil
}

// Result implements the Result RPC method
func (s *AuctionServer) Result(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error) {
	mut.Lock()
	defer mut.Unlock()

	return &pb.ResultResponse{IsActive: s.IsActive, HighestBid: int32(s.HighestBid)}, nil
}

func main() {
	writeToLogAndTerminal("Starting new server...")

	// Starts grpc server
	server := grpc.NewServer()

	// Initializes auction with default values
	auctionServer = &AuctionServer{}

	// Backup replicas go through this for loop until one becomes leader
	for {
		if becomesLeader() {
			break
		}
		receiveAuctionDataFromPrimaryReplica()

		time.Sleep(time.Second * 2)
	}

	// Initializes the server listener on port 5050
	var err error
	serverListener, err = net.Listen("tcp", "localhost:5050")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer serverListener.Close()

	// Handles grpc requests from clients
	go serveClients(server)

	// Continually sends data to backup replicas
	go func() {
		for {
			handleBackupReplicas(serverListener)
		}
	}()

	// Handles text input from the terminal to perform various tasks
	takeInputs(server)

}

func becomesLeader() bool {
	// Tries to become primary replica by acquiring the 5050 port used for server communication
	serverListener, err := net.Listen("tcp", "localhost:5050")
	if err != nil {
		return false
	}
	serverListener.Close()
	writeToLogAndTerminal("Primary replica has been found")
	return true
}

func receiveAuctionDataFromPrimaryReplica() {
	// Tries to dial up the primary replica
	conn, err := net.Dial("tcp", "localhost:5050")
	if err != nil {
		fmt.Println(err)
	}

	// Receives JSON data from the primary replica
	responseBuffer := make([]byte, 1024)
	n, err := conn.Read(responseBuffer)
	if err != nil {
		fmt.Println("Error reading JSON data:", err)
		return
	}

	// Decodes JSON data into AuctionServer struct
	var receivedAuctionServer AuctionServer
	err = json.Unmarshal(responseBuffer[:n], &receivedAuctionServer)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	// Sets the received struct as the struct
	auctionServer = &receivedAuctionServer
	writeToLogAndTerminal("Backup replica receives auction data from primary replica: " + auctionDataString())

	conn.Close()
}

func serveClients(server *grpc.Server) {
	clientListener, err := net.Listen("tcp", "localhost:8080")

	pb.RegisterAuctionServer(server, auctionServer)

	// Continually serves client requests
	go func() {
		err = server.Serve(clientListener)
		if err != nil {
			fmt.Println("Error serving listener:", err)
		}
	}()

	writeToLogAndTerminal("Server is running on localhost:8080")
}

func handleBackupReplicas(serverListener net.Listener) {

	// Accepts dial
	conn, err := serverListener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection:", err)
		return
	}
	defer conn.Close()

	// Encodes the struct to JSON
	jsonData, err := json.Marshal(auctionServer)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// Sends encoded data back
	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}

}

func takeInputs(server *grpc.Server) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter command:")
	for scanner.Scan() {
		command := scanner.Text()

		switch strings.ToLower(command) {
		case "start":
			auctionServer.HighestBid = 0
			auctionServer.MinimumBid = int32(rand.Intn(100))
			auctionServer.ItemName = templateAuctionItemNames[rand.Intn(len(templateAuctionItemNames)-1)]
			auctionServer.IsActive = true
			writeToLogAndTerminal("Server started new auction for " + auctionServer.ItemName + " starting at " + strconv.Itoa(int(auctionServer.MinimumBid)) + " dollars")
		case "end":
			auctionServer.IsActive = false
			writeToLogAndTerminal("Server ended auction with winning bid " + strconv.Itoa(int(auctionServer.HighestBid)))
		case "crash":
			writeToLogAndTerminal("Stopping gRPC server...")
			server.GracefulStop()
			return
		case "print":
			writeToLogAndTerminal(auctionDataString())
		default:
			fmt.Println("Invalid command. Valid commands: 'start', 'end', 'crash', 'print'")
		}
	}
}

func auctionDataString() string {
	return strconv.Itoa(int(auctionServer.HighestBid)) + " " + strconv.Itoa(int(auctionServer.MinimumBid)) + " " + strconv.FormatBool(auctionServer.IsActive) + " " + auctionServer.ItemName
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
