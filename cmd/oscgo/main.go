package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/EyevinnOSC/client-go/pkg/client"

	. "github.com/EyevinnOSC/client-go/internal/fetch"
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
	sat, err := oscClient.GetServiceAccessToken(*satServiceId, false)
	if err != nil {
		fmt.Println("Error getting service access token:", err)
		os.Exit(1)
	}
	fmt.Println(sat)
}

func listInstances(oscClient *client.OSCClient, args []string) {
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listServiceId := listCmd.String("s", "", "The service ID to list instances for")
	err := listCmd.Parse(args)
	if err != nil {
		fmt.Println("Error parsing list instances flags:", err)
		listCmd.Usage()
		os.Exit(1)
	}
	instances, err := oscClient.ListInstances(*listServiceId)
	if err != nil {
		fmt.Println("Error listing instances:", err)
		os.Exit(1)
	}
	json, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling instances:", err)
		os.Exit(1)
	}
	fmt.Println(string(json))
}

func callInstanceApi(oscClient *client.OSCClient, args []string) {
	callCmd := flag.NewFlagSet("call", flag.ExitOnError)
	callServiceId := callCmd.String("s", "", "The service ID to call the instance API for")
	callInstanceName := callCmd.String("n", "", "The name of the instance to call the API for")
	callMethod := callCmd.String("m", "GET", "The method to call on the instance API")
	callBody := callCmd.String("b", "", "The body to send to the instance API")
	err := callCmd.Parse(args)
	if err != nil {
		fmt.Println("Error parsing call instance API flags:", err)
		callCmd.Usage()
		os.Exit(1)
	}
	if *callServiceId == "" || *callInstanceName == "" {
		fail("Service ID and instance name are required")
	}
	args = args[len(args)-callCmd.NArg():]
	if len(args) != 1 {
		fail("Unexpected arguments: endpoint path must be provided")
	}
	callPath := args[0]
	if *callBody != "" && *callMethod == "GET" {
		*callMethod = "POST"
	}
	instance, err := oscClient.GetInstance(*callServiceId, *callInstanceName)
	if err != nil {
		fail("Error getting instance: %s", err.Error())
	}
	apiUrlStr := instance["url"]
	if apiUrlStr == "" {
		fail("URL empty")
	}
	apiUrl, err := url.Parse(apiUrlStr.(string))
	if err != nil {
		fail("Error parsing URL: %s", err.Error())
	}
	var requestBody *bytes.Buffer
	if *callBody != "" {
		requestBody = bytes.NewBuffer([]byte(*callBody))
	}
	endpointUrl := apiUrl.JoinPath(callPath)
	token, err := oscClient.GetServiceAccessToken(*callServiceId, false)
	if err != nil {
		fail("Error getting service access token: %s", err.Error())
	}
	response, err := Fetch(endpointUrl.String(), *callMethod, requestBody, Auth{Header: "x-jwt", Value: fmt.Sprintf("Bearer %s", token.Token)})
	if err != nil {
		fail("Error calling instance API: %s", err.Error())
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		fail("Error reading response body: %s", err.Error())
	}
	fmt.Printf("Response status: %d\n", response.StatusCode)
	fmt.Println(string(responseBody))
}

func logs(oscClient *client.OSCClient, args []string) {
	logsCmd := flag.NewFlagSet("logs", flag.ExitOnError)
	logsServiceId := logsCmd.String("s", "", "The service ID to get logs for")
	logsInstanceName := logsCmd.String("n", "", "The name of the instance to get logs for")
	logsTimestamps := logsCmd.Bool("t", false, "Whether to include timestamps in the logs")
	err := logsCmd.Parse(args)
	if err != nil {
		fmt.Println("Error parsing logs flags:", err)
		logsCmd.Usage()
		os.Exit(1)
	}
	logs, err := oscClient.GetLogsForInstance(*logsServiceId, *logsInstanceName, *logsTimestamps)
	if err != nil {
		fmt.Println("Error getting logs:", err)
		os.Exit(1)
	}

	fmt.Println(logs)
}

func fail(format string, a ...any) {
	fmt.Printf(format, a...)
	fmt.Println()
	os.Exit(1)
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
	case "list":
		listInstances(oscClient, os.Args[2:])
	case "logs":
		logs(oscClient, os.Args[2:])
	case "call":
		callInstanceApi(oscClient, os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}
