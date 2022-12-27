package networking

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
)

type RestClient struct {
	authorization   string
	baseUrl         string
	providerVersion string
}

func NewRestClient(baseUrl string, authorization string, providerVersion string) *RestClient {
	return &RestClient{
		authorization:   authorization,
		baseUrl:         baseUrl,
		providerVersion: providerVersion,
	}
}

func (c *RestClient) GetJSON(ctx context.Context, pathSegments []string, queryParams url.Values, apiVersion string) (*http.Response, error) {
	return c.sendRequest(ctx, http.MethodGet, pathSegments, queryParams, apiVersion)
}

func (c *RestClient) ParseJSON(response *http.Response, v any) error {
	if response == nil || response.Body == nil {
		return nil
	}

	var err error
	defer func() {
		if closeError := response.Body.Close(); closeError != nil {
			err = closeError
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	body = c.trimByteOrderMark(body)
	return json.Unmarshal(body, &v)
}

func (c *RestClient) UnwrapError(response *http.Response) (err error) {
	if response.ContentLength == 0 {
		message := "Request returned status: " + response.Status
		return &WrappedError{
			Message:    &message,
			StatusCode: &response.StatusCode,
		}
	}

	defer func() {
		if closeError := response.Body.Close(); closeError != nil {
			err = closeError
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	body = c.trimByteOrderMark(body)

	contentType, ok := response.Header[headerKeyContentType]
	if ok && len(contentType) > 0 && strings.Index(contentType[0], mediaTypeTextPlain) >= 0 {
		message := string(body)
		statusCode := response.StatusCode
		return WrappedError{Message: &message, StatusCode: &statusCode}
	}

	var wrappedError WrappedError
	err = json.Unmarshal(body, &wrappedError)
	wrappedError.StatusCode = &response.StatusCode
	if err != nil {
		return err
	}

	if wrappedError.Message == nil {
		var wrappedImproperError WrappedImproperError
		err = json.Unmarshal(body, &wrappedImproperError)
		if err == nil && wrappedImproperError.Value != nil && wrappedImproperError.Value.Message != nil {
			return &WrappedError{
				Message:    wrappedImproperError.Value.Message,
				StatusCode: &response.StatusCode,
			}
		}
	}

	return wrappedError
}

func (c *RestClient) createRequest(ctx context.Context, httpMethod string, pathSegments []string, queryParams url.Values, apiVersion string) (*http.Request, error) {
	endpointUrl := c.generateUrl(pathSegments, queryParams, apiVersion)
	tflog.Info(ctx, "Request Url = "+endpointUrl)
	req, err := http.NewRequest(httpMethod, endpointUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(headerKeyAccept, mediaTypeApplicationJson)
	req.Header.Add(headerKeyAuthorization, c.authorization)
	req.Header.Add(headerKeyUserAgent, "go/"+runtime.Version()+" ("+runtime.GOOS+" "+runtime.GOARCH+") terraform-provider-azuredevops/"+c.providerVersion)

	return req, nil
}

func (c *RestClient) generateUrl(pathSegments []string, queryParams url.Values, apiVersion string) string {
	var builder strings.Builder
	builder.WriteString(c.baseUrl)
	builder.WriteString("_apis")
	for _, segment := range pathSegments {
		builder.WriteString("/")
		builder.WriteString(url.PathEscape(segment))
	}
	if queryParams == nil {
		queryParams = make(url.Values)
	}
	queryParams.Add("api-version", apiVersion)
	builder.WriteString("?")
	builder.WriteString(queryParams.Encode())
	return builder.String()
}

func (c *RestClient) sendRequest(ctx context.Context, httpMethod string, pathSegments []string, queryParams url.Values, apiVersion string) (*http.Response, error) {
	req, err := c.createRequest(ctx, httpMethod, pathSegments, queryParams, apiVersion)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if resp != nil && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		err = c.UnwrapError(resp)
	}
	return resp, err
}

func (c *RestClient) trimByteOrderMark(body []byte) []byte {
	return bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
}
