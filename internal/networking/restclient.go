package networking

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/logger"
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
	headers := c.buildRequestHeaders()
	return c.sendRequest(ctx, http.MethodGet, pathSegments, queryParams, headers, nil, apiVersion)
}

func (c *RestClient) DeleteJSON(ctx context.Context, pathSegments []string, queryParams url.Values, apiVersion string) (*http.Response, error) {
	headers := c.buildRequestHeaders()
	return c.sendRequest(ctx, http.MethodDelete, pathSegments, queryParams, headers, nil, apiVersion)
}

func (c *RestClient) ParseJSON(ctx context.Context, response *http.Response, v any) error {
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
	logger.Debug(ctx, string(body))
	return json.Unmarshal(body, &v)
}

func (c *RestClient) PostJSON(ctx context.Context, pathSegments []string, queryParams url.Values, body any, apiVersion string) (*http.Response, error) {
	headers := c.buildRequestHeaders()
	return c.sendRequest(ctx, http.MethodPost, pathSegments, queryParams, headers, body, apiVersion)
}

func (c *RestClient) PatchJSON(ctx context.Context, pathSegments []string, queryParams url.Values, body any, apiVersion string) (*http.Response, error) {
	headers := c.buildRequestHeaders()
	return c.sendRequest(ctx, http.MethodPatch, pathSegments, queryParams, headers, body, apiVersion)
}

func (c *RestClient) PatchJSONSpecialContentType(ctx context.Context, pathSegments []string, queryParams url.Values, body any, apiVersion string) (*http.Response, error) {
	headers := c.buildRequestHeaders()
	headers[headerKeyContentType] = mediaTypeApplicationJsonPatch
	return c.sendRequest(ctx, http.MethodPatch, pathSegments, queryParams, headers, body, apiVersion)
}

func (c *RestClient) PutJSON(ctx context.Context, pathSegments []string, queryParams url.Values, body any, apiVersion string) (*http.Response, error) {
	headers := c.buildRequestHeaders()
	return c.sendRequest(ctx, http.MethodPut, pathSegments, queryParams, headers, body, apiVersion)
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

func (c *RestClient) buildRequestHeaders() map[string]string {
	return map[string]string{
		headerKeyAccept:        mediaTypeApplicationJson,
		headerKeyAuthorization: c.authorization,
		headerKeyContentType:   mediaTypeApplicationJson,
		headerKeyUserAgent:     "go/" + runtime.Version() + " (" + runtime.GOOS + " " + runtime.GOARCH + ") terraform-provider-azuredevops/" + c.providerVersion,
	}
}

func (c *RestClient) generateUrl(pathSegments []string, queryParams url.Values, apiVersion string) string {
	var builder strings.Builder
	builder.WriteString(strings.TrimSuffix(c.baseUrl, "/"))
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

func (c *RestClient) sendRequest(ctx context.Context, httpMethod string, pathSegments []string, queryParams url.Values, headers map[string]string, body any, apiVersion string) (*http.Response, error) {
	endpointUrl := c.generateUrl(pathSegments, queryParams, apiVersion)
	logger.Info(ctx, httpMethod+" "+endpointUrl)
	var jsonReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		logger.Debug(ctx, string(jsonBody))
		jsonReader = bytes.NewReader(jsonBody)
	}
	req, err := http.NewRequest(httpMethod, endpointUrl, jsonReader)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
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
