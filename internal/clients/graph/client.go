package graph

import (
	"context"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"net/url"
)

const (
	pathApis        = "_apis"
	pathDescriptors = "descriptors"
	pathGraph       = "graph"
	pathGroups      = "groups"
)

type Client struct {
	vsspsClient *networking.RestClient
}

func NewClient(vsspsClient *networking.RestClient) *Client {
	return &Client{
		vsspsClient: vsspsClient,
	}
}

func (c *Client) GetDescriptor(ctx context.Context, storageKey string) (*string, error) {
	pathSegments := []string{pathApis, pathGraph, pathDescriptors, storageKey}
	resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var descriptor *GraphDescriptorResult
	err = c.vsspsClient.ParseJSON(ctx, resp, &descriptor)
	return descriptor.Value, err
}

func (c *Client) GetGroups(ctx context.Context, descriptor string, continuationToken string) (*[]GraphGroup, error) {
	pathSegments := []string{pathApis, pathGraph, pathGroups}
	queryParams := url.Values{"scopeDescriptor": []string{descriptor}}
	if continuationToken != "" {
		queryParams.Add("continuationToken", continuationToken)
	}

	var groups []GraphGroup
	hasMore := true
	for hasMore {
		resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, queryParams, networking.ApiVersion70Preview1)
		if err != nil {
			return nil, err
		}

		var collection *GraphGroupCollection
		err = c.vsspsClient.ParseJSON(ctx, resp, &collection)
		if err != nil {
			return nil, err
		}

		for _, group := range *collection.Value {
			groups = append(groups, group)
		}

		continuationToken = resp.Header.Get(networking.HeaderKeyContinuationToken)
		hasMore = continuationToken != ""
	}

	return &groups, nil
}
