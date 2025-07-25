// A simple test client for the SpaceNet server
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

func main() {
	// Parse command line flags
	name := flag.String("name", "Anonymous", "Your name for the claim")
	server := flag.String("server", "::1", "IPv6 address of the server")
	port := flag.Int("port", 1337, "Port number of the server")
	flag.Parse()

	// Resolve server address
	serverAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(*server, fmt.Sprintf("%d", *port)))
	if err != nil {
		log.Fatalf("Failed to resolve server address: %v", err)
	}

	// Create a UDP connection
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send the claim
	log.Printf("Sending claim from %s to %s", conn.LocalAddr(), serverAddr)
	_, err = conn.Write([]byte(*name))
	if err != nil {
		log.Fatalf("Failed to send claim: %v", err)
	}

	log.Printf("Claim sent successfully!")
}
