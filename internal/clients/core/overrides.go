package core

import (
	"encoding/json"
	"strings"
	"time"
)

func (t *Time) UnmarshalJSON(b []byte) error {
	t2 := time.Time{}
	err := json.Unmarshal(b, &t2)

	if err != nil {
		parseError, ok := err.(*time.ParseError)
		if ok {
			if parseError.Value == "\"0001-01-01T00:00:00\"" {
				// ignore errors for 0001-01-01T00:00:00 dates. The Azure DevOps service
				// returns default dates in a format that is invalid for a time.Time. The
				// correct value would have a 'z' at the end to represent utc. We are going
				// to ignore this error, and set the value to the default time.Time value.
				// https://github.com/microsoft/azure-devops-go-api/issues/17
				err = nil
			} else {
				// workaround for bug https://github.com/microsoft/azure-devops-go-api/issues/59
				// policy.CreatePolicyConfiguration returns an invalid date format of form
				// "2006-01-02T15:04:05.999999999"
				var innerError error
				t2, innerError = time.Parse("2006-01-02T15:04:05.999999999", strings.Trim(parseError.Value, "\""))
				if innerError == nil {
					err = nil
				}
			}
		}
	}

	t.Time = t2
	return err
}
