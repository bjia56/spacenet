// A simple test client for the SpaceNet server
package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	// Parse command line flags
	name := flag.String("name", "Anonymous", "Your name for the claim")
	server := flag.String("server", "::1", "IPv6 address of the server")
	port := flag.Int("port", 1337, "Port number of the server")
	source := flag.String("source", "", "Source IP address to claim (optional)")
	flag.Parse()

	// Build sendip command
	cmd := []string{"sendip", "-d", *name, "-p", "ipv6"}
	if *source != "" {
		cmd = append(cmd, "-6s", *source)
	}
	cmd = append(cmd, "-p", "udp", "-ud", fmt.Sprintf("%d", *port), *server)

	// Exec sendip command
	err := exec.Command("sudo", "sh", "-c", strings.Join(cmd, " ")).Run()
	if err != nil {
		log.Fatalf("Failed to execute sendip command: %v", err)
	}

	log.Printf("Claim sent successfully!")
}
