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

type ServiceInstanceOption struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Enum        []string `json:"enums"`
	Mandatory   bool     `json:"mandatory"`
}

type SubscriptionFee struct {
	CurrencyCode string  `json:"currencyCode"`
	Value        float64 `json:"value"`
}

type ServiceMetadata struct {
	OwnerTeamId       string          `json:"ownerTeamId"`
	CompanyName       string          `json:"companyName"`
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	DocumentationUrl  string          `json:"documentationUrl"`
	RepoUrl           string          `json:"repoUrl"`
	ImagePreviewUrl   string          `json:"imagePreviewUrl"`
	InstanceNoun      string          `json:"instanceNoun"`
	IndicativePricing string          `json:"indicativePricing"`
	SubscriptionFee   SubscriptionFee `json:"subscriptionFee"`
	ExcludeFromFree   bool            `json:"excludeFromFree"`
	Category          string          `json:"category"`
	Keywords          []string        `json:"keywords"`
}

type Service struct {
	ServiceId              string                  `json:"serviceId"`
	Metadata               ServiceMetadata         `json:"serviceMetadata"`
	ApiUrl                 string                  `json:"apiUrl"`
	ServiceInstanceOptions []ServiceInstanceOption `json:"serviceInstanceOptions"`
	OpenSourceLicense      string                  `json:"openSourceLicense"`
	OpenSourceLicenseText  string                  `json:"openSourceLicenseText"`
	ServiceType            string                  `json:"serviceType"`
	Status                 string                  `json:"status"`
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
