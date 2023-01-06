package graph

import (
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
)

const (
	pathApis = "_apis"
)

type Client struct {
	vsspsClient *networking.RestClient
}

func NewClient(vsspsClient *networking.RestClient) *Client {
	return &Client{
		vsspsClient: vsspsClient,
	}
}
