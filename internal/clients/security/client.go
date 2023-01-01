package security

import (
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
)

const (
	pathApis = "_apis"
)

type Client struct {
	restClient *networking.RestClient
}

func NewClient(restClient *networking.RestClient) *Client {
	return &Client{
		restClient: restClient,
	}
}
