package client

import (
	"fmt"
)

type OSCClientConfig struct {
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

type OSCClient struct {
	PersonalAccessToken string
	Environment         string
}

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
