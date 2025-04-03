package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/EyevinnOSC/client-go/pkg/client"
)

func usage() {
	fmt.Println("Usage: oscgo <command> [flags]")
}

func serviceAccessToken(oscClient *client.OSCClient, args []string) {
	satCmd := flag.NewFlagSet("sat", flag.ExitOnError)
	satServiceId := satCmd.String("s", "", "The service ID to get the service access token for")
	err := satCmd.Parse(args)
	if err != nil {
		fmt.Println("Error parsing service access token flags:", err)
		satCmd.Usage()
		os.Exit(1)
	}
	if *satServiceId == "" {
		satCmd.Usage()
		os.Exit(1)
	}
	sat, err := oscClient.GetServiceAccessToken(*satServiceId)
	if err != nil {
		fmt.Println("Error getting service access token:", err)
		os.Exit(1)
	}
	fmt.Println(sat)
}

func main() {

	config := &client.OSCClientConfig{
		PersonalAccessToken: os.Getenv("OSC_PAT"),
		Environment:         "prod",
	}
	oscClient, err := client.NewOSCClient(config)
	if err != nil {
		fmt.Println("Error creating OSCClient:", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {

	case "sat":
		serviceAccessToken(oscClient, os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}
