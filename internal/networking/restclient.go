package networking

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/logger"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
)

type RestClient struct {
	authorization   string
	baseUrl         string
	providerVersion string
}

type NoJSON string

func NewRestClient(baseUrl string, authorization string, providerVersion string) *RestClient {
	return &RestClient{
		authorization:   authorization,
		baseUrl:         baseUrl,
		providerVersion: providerVersion,
	}
}

func GetJSON[T any](c *RestClient, ctx context.Context, pathSegments []string, queryParams url.Values, apiVersion string) (*T, *http.Response, error) {
	return sendRequestJSON[T](c, ctx, http.MethodGet, pathSegments, queryParams, nil, nil, apiVersion)
}

func DeleteJSON[T any](c *RestClient, ctx context.Context, pathSegments []string, queryParams url.Values, apiVersion string) (*T, *http.Response, error) {
	return sendRequestJSON[T](c, ctx, http.MethodDelete, pathSegments, queryParams, nil, nil, apiVersion)
}

func PostJSON[T any](c *RestClient, ctx context.Context, pathSegments []string, queryParams url.Values, body any, apiVersion string) (*T, *http.Response, error) {
	return sendRequestJSON[T](c, ctx, http.MethodPost, pathSegments, queryParams, nil, body, apiVersion)
}

func PatchJSON[T any](c *RestClient, ctx context.Context, pathSegments []string, queryParams url.Values, body any, apiVersion string) (*T, *http.Response, error) {
	return sendRequestJSON[T](c, ctx, http.MethodPatch, pathSegments, queryParams, nil, body, apiVersion)
}

func PatchJSONSpecialContentType[T any](c *RestClient, ctx context.Context, pathSegments []string, queryParams url.Values, body any, apiVersion string) (*T, *http.Response, error) {
	headers := c.buildRequestHeaders()
	headers[headerKeyContentType] = mediaTypeApplicationJsonPatch
	return sendRequestJSON[T](c, ctx, http.MethodPatch, pathSegments, queryParams, &headers, body, apiVersion)
}

func PutJSON[T any](c *RestClient, ctx context.Context, pathSegments []string, queryParams url.Values, body any, apiVersion string) (*T, *http.Response, error) {
	return sendRequestJSON[T](c, ctx, http.MethodPut, pathSegments, queryParams, nil, body, apiVersion)
}

// Private Methods

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

func isJSON[T any]() bool {
	return reflect.TypeOf(new(T)).String() != "*networking.NoJSON"
}

func (c *RestClient) parseJSON(ctx context.Context, response *http.Response, v any) error {
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
		err = c.unwrapError(resp)
	}
	return resp, err
}

func sendRequestJSON[T any](c *RestClient, ctx context.Context, httpMethod string, pathSegments []string, queryParams url.Values, headers *map[string]string, body any, apiVersion string) (*T, *http.Response, error) {
	var requestHeaders map[string]string
	if headers != nil {
		requestHeaders = *headers
	} else {
		requestHeaders = c.buildRequestHeaders()
	}
	resp, err := c.sendRequest(ctx, httpMethod, pathSegments, queryParams, requestHeaders, body, apiVersion)
	if err != nil {
		return nil, nil, err
	}

	var result *T
	if isJSON[T]() {
		err = c.parseJSON(ctx, resp, &result)
	}
	return result, resp, err
}

func (c *RestClient) trimByteOrderMark(body []byte) []byte {
	return bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
}

func (c *RestClient) unwrapError(response *http.Response) (err error) {
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
