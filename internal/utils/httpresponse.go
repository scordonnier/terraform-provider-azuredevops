package utils

import (
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"net/http"
	"strings"
)

func ResponseWasNotFound(err error) bool {
	statusNotFound := ResponseWasStatusCode(err, http.StatusNotFound)
	if statusNotFound {
		return statusNotFound
	}

	// Some APIs return 400 BadRequest with the VS800075 error message if
	// DevOps Project doesn't exist. If parent project doesn't exist, all
	// child resources are considered "doesn't exist".
	statusBadRequest := ResponseWasStatusCode(err, http.StatusBadRequest)
	if statusBadRequest {
		return ResponseContainsStatusMessage(err, "VS800075")
	}
	return false
}

func ResponseWasStatusCode(err error, statusCode int) bool {
	if err == nil {
		return false
	}
	if wrapperErr, ok := err.(networking.WrappedError); ok {
		if wrapperErr.StatusCode != nil && *wrapperErr.StatusCode == statusCode {
			return true
		}
	}
	return false
}

func ResponseContainsStatusMessage(err error, statusMessage string) bool {
	if err == nil {
		return false
	}
	if wrapperErr, ok := err.(networking.WrappedError); ok {
		if wrapperErr.Message == nil {
			return false
		}
		return strings.Contains(*wrapperErr.Message, statusMessage)
	}
	return false
}
