package osaasclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type ContextConfig struct {
	PersonalAccessToken string
	Environment         string
}

type ServiceAccessToken struct {
	ServiceId string `json:"serviceId"`
	Token     string `json:"token"`
	Expiry    int64  `json:"expiry"`
}

type Service struct {
	ServiceId   string `json:"serviceId"`
	ApiUrl      string `json:"apiUrl"`
	ServiceType string `json:"serviceType"`
}

type Subscriptions struct {
	TeamId   string   `json:"teamId"`
	Services []string `json:"services"`
}

type Context struct {
	PersonalAccessToken string
	Environment         string
}

func NewContext(config *ContextConfig) (*Context, error) {
	token := config.PersonalAccessToken
	if token == "" {
		token = os.Getenv("OSC_ACCESS_TOKEN")
	}
	if token == "" {
		return nil, errors.New("personal access token is required to create a context." +
			" Please provide it in the config or set the OSC_ACCESS_TOKEN environment variable")
	}
	env := config.Environment
	if env == "" {
		env = "prod"
	}
	return &Context{
		PersonalAccessToken: token,
		Environment:         env,
	}, nil
}

func (ctx *Context) GetPersonalAccessToken() string {
	return ctx.PersonalAccessToken
}

func (ctx *Context) GetEnvironment() string {
	return ctx.Environment
}

func (ctx *Context) GetServiceAccessToken(serviceId string) (string, error) {
	serviceUrl := fmt.Sprintf("https://catalog.svc.%s.osaas.io/mysubscriptions", ctx.Environment)

	var services []Service

	err := createFetch(serviceUrl, "GET", nil, &services,
		Auth{"x-pat-jwt", fmt.Sprintf("Bearer %s", ctx.PersonalAccessToken)})
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
		err = ctx.ActivateService(serviceId)
		if err != nil {
			return "", err
		}
	}

	satUrl := fmt.Sprintf("https://token.svc.%s.osaas.io/servicetoken", ctx.Environment)

	body, _ := json.Marshal(map[string]string{
		"serviceId": serviceId,
	})

	var serviceAccessToken ServiceAccessToken
	err = createFetch(satUrl, "POST", bytes.NewBuffer(body),
		&serviceAccessToken, Auth{"x-pat-jwt", fmt.Sprintf("Bearer %s", ctx.PersonalAccessToken)})
	if err != nil {
		return "", err
	}

	return serviceAccessToken.Token, nil
}

func (ctx *Context) ActivateService(serviceId string) error {
	serviceUrl := fmt.Sprintf("https://catalog.svc.%s.osaas.io/mysubscriptions", ctx.Environment)

	body, _ := json.Marshal(Subscriptions{
		Services: []string{serviceId},
	})

	return createFetch(serviceUrl, "POST", bytes.NewBuffer(body), nil,
		Auth{"x-pat-jwt", fmt.Sprintf("Bearer %s", ctx.PersonalAccessToken)})
}

func (ctx *Context) RefreshServiceAccessToken(serviceId string) (string, error) {
	return ctx.GetServiceAccessToken(serviceId)
}
