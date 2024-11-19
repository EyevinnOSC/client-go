package osaasclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
)

type Port struct {
	ExternalIP   string `json:"externalIp"`
	ExternalPort int    `json:"externalPort"`
	InternalPort int    `json:"internalPort"`
}

type FetchError struct {
	HTTPCode int
	Message  string
}

func (e FetchError) Error() string {
	return fmt.Sprintf("HTTP Code: %d, Message: %s", e.HTTPCode, e.Message)
}

type UnauthorizedError struct{}

func (e UnauthorizedError) Error() string {
	return "Unauthorized access"
}

func GetService(ctx *Context, serviceId string) (*Service, error) {
	serviceURL := fmt.Sprintf("https://catalog.svc.%s.osaas.io/mysubscriptions", ctx.GetEnvironment())

	var services []Service
	err := createFetch(serviceURL, "GET", nil, &services,
		Auth{"x-pat-jwt", fmt.Sprintf("Bearer %s", ctx.GetPersonalAccessToken())})
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		if service.ServiceId == serviceId {
			return &service, nil
		}
	}
	return nil, errors.New("service not found in your subscriptions")
}

func CreateInstance(ctx *Context, serviceId string, token string,
	body map[string]interface{}) (map[string]interface{}, error) {
	err := ctx.ActivateService(serviceId)
	if err != nil {
		return nil, err
	}

	service, err := GetService(ctx, serviceId)
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

	return instance, nil
}

func RemoveInstance(ctx *Context, serviceId string, name string, token string) error {
	service, err := GetService(ctx, serviceId)
	if err != nil {
		return err
	}

	instanceURL := fmt.Sprintf("%s/%s", service.ApiUrl, name)

	slog.Debug(fmt.Sprintf("Deleting on instanceUrl: %s", instanceURL))

	return createFetch(instanceURL, "DELETE", nil, nil, Auth{"x-jwt", fmt.Sprintf("Bearer %s", token)})
}

func GetInstance(ctx *Context, serviceId string, name string, token string) (map[string]interface{}, error) {
	service, err := GetService(ctx, serviceId)
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

func ListInstances(context *Context, serviceId, token string) ([]map[string]interface{}, error) {
	service, err := GetService(context, serviceId)
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

func GetPortsForInstance(context *Context, serviceId, name, token string) ([]Port, error) {
	service, err := GetService(context, serviceId)
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

func GetLogsForInstance(context *Context, serviceId, name, token string) ([]string, error) {
	service, err := GetService(context, serviceId)
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

func AddServiceSecret(ctx *Context, serviceId, secretName, secretData string) error {
	secretUrl := fmt.Sprintf("https://deploy.svc.%s.osaas.io/mysecrets/%s", ctx.GetEnvironment(), serviceId)

	body := map[string]string{"secretName": secretName, "secretData": secretData}
	bodyBytes, _ := json.Marshal(body)
	bodyBuffer := bytes.NewBuffer(bodyBytes)

	var instance map[string]interface{}
	err := createFetch(secretUrl, "POST", bodyBuffer, &instance, Auth{"x-pat-jwt", fmt.Sprintf("Bearer %s", ctx.GetPersonalAccessToken())})
	return err
}
