package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"

	. "github.com/EyevinnOSC/client-go/internal/fetch"
)

func NewOSCClient(config *OSCClientConfig) (*OSCClient, error) {
	token := config.PersonalAccessToken
	if token == "" {
		token = os.Getenv("OSC_ACCESS_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("personal access token is required to create an osc client." +
			" Please provide it in the config or set the OSC_ACCESS_TOKEN environment variable")
	}
	env := config.Environment
	if env == "" {
		env = "prod"
	}
	return &OSCClient{
		PersonalAccessToken: token,
		Environment:         env,
		TokenCache:          make(map[string]ServiceAccessToken),
	}, nil
}

func (oscClient *OSCClient) GetPersonalAccessToken() string {
	return oscClient.PersonalAccessToken
}

func (oscClient *OSCClient) GetEnvironment() string {
	return oscClient.Environment
}

func (oscClient *OSCClient) GetCachedToken(serviceId string) (ServiceAccessToken, bool) {
	token, ok := oscClient.TokenCache[serviceId]
	if ok && token.Expiry > time.Now().Unix()+30 {
		return token, true
	}
	return ServiceAccessToken{}, false
}

func (oscClient *OSCClient) GetServiceAccessToken(serviceId string, forceRefresh bool) (ServiceAccessToken, error) {
	if !forceRefresh {
		token, ok := oscClient.GetCachedToken(serviceId)
		if ok {
			return token, nil
		}
	}

	serviceUrl := fmt.Sprintf("https://catalog.svc.%s.osaas.io/mysubscriptions", oscClient.Environment)

	var services []Service

	err := FetchJson(serviceUrl, "GET", nil, &services,
		Auth{Header: "x-pat-jwt", Value: fmt.Sprintf("Bearer %s", oscClient.PersonalAccessToken)})
	if err != nil {
		return ServiceAccessToken{}, err
	}

	var service *Service
	for _, svc := range services {
		if svc.ServiceId == serviceId {
			service = &svc
			break
		}
	}

	if service == nil {
		err = oscClient.ActivateService(serviceId)
		if err != nil {
			return ServiceAccessToken{}, err
		}
	}

	satUrl := fmt.Sprintf("https://token.svc.%s.osaas.io/servicetoken", oscClient.Environment)

	body, _ := json.Marshal(map[string]string{
		"serviceId": serviceId,
	})

	var serviceAccessToken ServiceAccessToken
	err = FetchJson(satUrl, "POST", bytes.NewBuffer(body),
		&serviceAccessToken, Auth{Header: "x-pat-jwt", Value: fmt.Sprintf("Bearer %s", oscClient.PersonalAccessToken)})
	if err != nil {
		return ServiceAccessToken{}, err
	}
	oscClient.TokenCache[serviceId] = serviceAccessToken

	return serviceAccessToken, nil
}

func (oscClient *OSCClient) ActivateService(serviceId string) error {
	serviceUrl := fmt.Sprintf("https://catalog.svc.%s.osaas.io/mysubscriptions", oscClient.Environment)

	body, _ := json.Marshal(Subscriptions{
		Services: []string{serviceId},
	})

	return FetchJson(serviceUrl, "POST", bytes.NewBuffer(body), nil,
		Auth{Header: "x-pat-jwt", Value: fmt.Sprintf("Bearer %s", oscClient.PersonalAccessToken)})
}

/*
	func (oscClient *OSCClient) RefreshServiceAccessToken(serviceId string) (string, error) {
		return oscClient.GetServiceAccessToken(serviceId)
	}
*/
func (oscClient *OSCClient) GetService(serviceId string) (*Service, error) {
	serviceURL := fmt.Sprintf("https://catalog.svc.%s.osaas.io/mysubscriptions", oscClient.GetEnvironment())

	var services []Service
	err := FetchJson(serviceURL, "GET", nil, &services,
		Auth{Header: "x-pat-jwt", Value: fmt.Sprintf("Bearer %s", oscClient.GetPersonalAccessToken())})
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		if service.ServiceId == serviceId {
			return &service, nil
		}
	}
	return nil, fmt.Errorf("service not found in your subscriptions")
}

func (oscClient *OSCClient) CreateInstance(serviceId string,
	body map[string]interface{}) (map[string]interface{}, error) {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return nil, err
	}

	instanceURL := service.ApiUrl
	bodyBytes, _ := json.Marshal(body)
	bodyBuffer := bytes.NewBuffer(bodyBytes)
	token, err := oscClient.GetServiceAccessToken(serviceId, false)
	if err != nil {
		return nil, err
	}
	var instance map[string]interface{}
	err = FetchJson(instanceURL, "POST", bodyBuffer, &instance, Auth{Header: "x-jwt", Value: fmt.Sprintf("Bearer %s", token.Token)})
	if err != nil {
		return nil, err
	}

	if instance["reason"] != nil {
		return nil, fmt.Errorf(instance["reason"].(string))
	}

	return instance, nil
}

func (oscClient *OSCClient) RemoveInstance(serviceId string, name string) error {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return err
	}

	instanceURL := fmt.Sprintf("%s/%s", service.ApiUrl, name)

	slog.Debug("Deleting instance", "instanceUrl", instanceURL)
	token, err := oscClient.GetServiceAccessToken(serviceId, false)
	if err != nil {
		return err
	}
	return FetchJson(instanceURL, "DELETE", nil, nil, Auth{Header: "x-jwt", Value: fmt.Sprintf("Bearer %s", token.Token)})
}

func (oscClient *OSCClient) GetInstance(serviceId string, name string) (map[string]interface{}, error) {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return nil, err
	}

	instanceURL := fmt.Sprintf("%s/%s", service.ApiUrl, name)

	token, err := oscClient.GetServiceAccessToken(serviceId, false)
	if err != nil {
		return nil, err
	}

	var instance map[string]interface{}
	err = FetchJson(instanceURL, "GET", nil, &instance, Auth{Header: "x-jwt", Value: fmt.Sprintf("Bearer %s", token.Token)})
	if err != nil {
		if fetchErr, ok := err.(FetchError); ok {
			if fetchErr.HTTPCode == 404 {
				return nil, nil
			}
		}
		return nil, err
	}
	return instance, nil
}

func (oscClient *OSCClient) ListInstances(serviceId string) ([]map[string]interface{}, error) {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return nil, err
	}

	instanceUrl := service.ApiUrl
	token, err := oscClient.GetServiceAccessToken(serviceId, false)
	if err != nil {
		return nil, err
	}

	var instances []map[string]interface{}
	err = FetchJson(instanceUrl, "GET", nil, &instances, Auth{Header: "x-jwt", Value: fmt.Sprintf("Bearer %s", token.Token)})
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (oscClient *OSCClient) GetPortsForInstance(serviceId string, name string) ([]Port, error) {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return nil, err
	}

	instanceUrl, err := url.Parse(service.ApiUrl)
	if err != nil {
		return nil, err
	}

	portsUrl := fmt.Sprintf("https://%s/ports/%s", instanceUrl.Host, name)
	token, err := oscClient.GetServiceAccessToken(serviceId, false)
	if err != nil {
		return nil, err
	}

	var ports []Port
	err = FetchJson(portsUrl, "GET", nil, &ports, Auth{Header: "x-jwt", Value: fmt.Sprintf("Bearer %s", token.Token)})
	if err != nil {
		return nil, err
	}

	return ports, nil
}

func (oscClient *OSCClient) GetLogsForInstance(serviceId string, name string, timestamps bool) (string, error) {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return "", err
	}

	instanceUrl, err := url.Parse(service.ApiUrl)
	if err != nil {
		return "", err
	}

	logsUrl := fmt.Sprintf("https://%s/logs/%s?timestamps=%t", instanceUrl.Host, name, timestamps)
	token, err := oscClient.GetServiceAccessToken(serviceId, false)
	if err != nil {
		return "", err
	}

	logBytes, err := FetchBytes(logsUrl, "GET", nil, Auth{Header: "x-jwt", Value: fmt.Sprintf("Bearer %s", token.Token)})
	if err != nil {
		return "", err
	}

	return string(logBytes), nil
}

func (oscClient *OSCClient) AddServiceSecret(serviceId string, secretName string, secretData string) error {
	secretUrl := fmt.Sprintf("https://deploy.svc.%s.osaas.io/mysecrets/%s", oscClient.GetEnvironment(), serviceId)

	body := map[string]string{"secretName": secretName, "secretData": secretData}
	bodyBytes, _ := json.Marshal(body)
	bodyBuffer := bytes.NewBuffer(bodyBytes)

	var instance map[string]interface{}
	err := FetchJson(secretUrl, "POST", bodyBuffer, &instance, Auth{Header: "x-pat-jwt", Value: fmt.Sprintf("Bearer %s", oscClient.GetPersonalAccessToken())})
	return err
}
