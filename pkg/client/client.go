package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
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
	}, nil
}

func (oscClient *OSCClient) GetPersonalAccessToken() string {
	return oscClient.PersonalAccessToken
}

func (oscClient *OSCClient) GetEnvironment() string {
	return oscClient.Environment
}

func (oscClient *OSCClient) GetServiceAccessToken(serviceId string) (string, error) {
	serviceUrl := fmt.Sprintf("https://catalog.svc.%s.osaas.io/mysubscriptions", oscClient.Environment)

	var services []Service

	err := createFetch(serviceUrl, "GET", nil, &services,
		Auth{"x-pat-jwt", fmt.Sprintf("Bearer %s", oscClient.PersonalAccessToken)})
	if err != nil {
		return "", err
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
			return "", err
		}
	}

	satUrl := fmt.Sprintf("https://token.svc.%s.osaas.io/servicetoken", oscClient.Environment)

	body, _ := json.Marshal(map[string]string{
		"serviceId": serviceId,
	})

	var serviceAccessToken ServiceAccessToken
	err = createFetch(satUrl, "POST", bytes.NewBuffer(body),
		&serviceAccessToken, Auth{"x-pat-jwt", fmt.Sprintf("Bearer %s", oscClient.PersonalAccessToken)})
	if err != nil {
		return "", err
	}

	return serviceAccessToken.Token, nil
}

func (oscClient *OSCClient) ActivateService(serviceId string) error {
	serviceUrl := fmt.Sprintf("https://catalog.svc.%s.osaas.io/mysubscriptions", oscClient.Environment)

	body, _ := json.Marshal(Subscriptions{
		Services: []string{serviceId},
	})

	return createFetch(serviceUrl, "POST", bytes.NewBuffer(body), nil,
		Auth{"x-pat-jwt", fmt.Sprintf("Bearer %s", oscClient.PersonalAccessToken)})
}

func (oscClient *OSCClient) RefreshServiceAccessToken(serviceId string) (string, error) {
	return oscClient.GetServiceAccessToken(serviceId)
}

func (oscClient *OSCClient) GetService(serviceId string) (*Service, error) {
	serviceURL := fmt.Sprintf("https://catalog.svc.%s.osaas.io/mysubscriptions", oscClient.GetEnvironment())

	var services []Service
	err := createFetch(serviceURL, "GET", nil, &services,
		Auth{"x-pat-jwt", fmt.Sprintf("Bearer %s", oscClient.GetPersonalAccessToken())})
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

func (oscClient *OSCClient) CreateInstance(serviceId string, token string,
	body map[string]interface{}) (map[string]interface{}, error) {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return nil, err
	}

	instanceURL := service.ApiUrl
	bodyBytes, _ := json.Marshal(body)
	bodyBuffer := bytes.NewBuffer(bodyBytes)

	var instance map[string]interface{}
	err = createFetch(instanceURL, "POST", bodyBuffer, &instance, Auth{"x-jwt", fmt.Sprintf("Bearer %s", token)})
	if err != nil {
		return nil, err
	}

	if instance["reason"] != nil {
		return nil, fmt.Errorf(instance["reason"].(string))
	}

	return instance, nil
}

func (oscClient *OSCClient) RemoveInstance(serviceId string, name string, token string) error {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return err
	}

	instanceURL := fmt.Sprintf("%s/%s", service.ApiUrl, name)

	slog.Debug("Deleting instance", "instanceUrl", instanceURL)

	return createFetch(instanceURL, "DELETE", nil, nil, Auth{"x-jwt", fmt.Sprintf("Bearer %s", token)})
}

func (oscClient *OSCClient) GetInstance(serviceId string, name string, token string) (map[string]interface{}, error) {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return nil, err
	}

	instanceURL := fmt.Sprintf("%s/%s", service.ApiUrl, name)

	var instance map[string]interface{}
	err = createFetch(instanceURL, "GET", nil, &instance, Auth{"x-jwt", fmt.Sprintf("Bearer %s", token)})
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

func (oscClient *OSCClient) ListInstances(serviceId, token string) ([]map[string]interface{}, error) {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return nil, err
	}

	instanceUrl := service.ApiUrl

	var instances []map[string]interface{}
	err = createFetch(instanceUrl, "GET", nil, &instances, Auth{"x-jwt", fmt.Sprintf("Bearer %s", token)})
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (oscClient *OSCClient) GetPortsForInstance(serviceId, name, token string) ([]Port, error) {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return nil, err
	}

	instanceUrl, err := url.Parse(service.ApiUrl)
	if err != nil {
		return nil, err
	}

	portsUrl := fmt.Sprintf("https://%s/ports/%s", instanceUrl.Host, name)

	var ports []Port
	err = createFetch(portsUrl, "GET", nil, &ports, Auth{"x-jwt", fmt.Sprintf("Bearer %s", token)})
	if err != nil {
		return nil, err
	}

	return ports, nil
}

func (oscClient *OSCClient) GetLogsForInstance(serviceId, name, token string) ([]string, error) {
	service, err := oscClient.GetService(serviceId)
	if err != nil {
		return nil, err
	}

	instanceUrl, err := url.Parse(service.ApiUrl)
	if err != nil {
		return nil, err
	}

	logsUrl := fmt.Sprintf("https://%s/logs/%s", instanceUrl.Host, name)

	var logs []string
	err = createFetch(logsUrl, "GET", nil, &logs, Auth{"x-jwt", fmt.Sprintf("Bearer %s", token)})
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (oscClient *OSCClient) AddServiceSecret(serviceId, secretName, secretData string) error {
	secretUrl := fmt.Sprintf("https://deploy.svc.%s.osaas.io/mysecrets/%s", oscClient.GetEnvironment(), serviceId)

	body := map[string]string{"secretName": secretName, "secretData": secretData}
	bodyBytes, _ := json.Marshal(body)
	bodyBuffer := bytes.NewBuffer(bodyBytes)

	var instance map[string]interface{}
	err := createFetch(secretUrl, "POST", bodyBuffer, &instance, Auth{"x-pat-jwt", fmt.Sprintf("Bearer %s", oscClient.GetPersonalAccessToken())})
	return err
}
