package networking

import "strconv"

type ImproperError struct {
	Message *string `json:"Message,omitempty"`
}

type WrappedError struct {
	ExceptionId      *string                 `json:"$id,omitempty"`
	InnerError       *WrappedError           `json:"innerException,omitempty"`
	Message          *string                 `json:"message,omitempty"`
	TypeName         *string                 `json:"typeName,omitempty"`
	TypeKey          *string                 `json:"typeKey,omitempty"`
	ErrorCode        *int                    `json:"errorCode,omitempty"`
	EventId          *int                    `json:"eventId,omitempty"`
	CustomProperties *map[string]interface{} `json:"customProperties,omitempty"`
	StatusCode       *int
}

type WrappedImproperError struct {
	Count *int           `json:"count,omitempty"`
	Value *ImproperError `json:"value,omitempty"`
}

func (e WrappedError) Error() string {
	if e.Message == nil {
		if e.StatusCode != nil {
			return "API call returned status code " + strconv.Itoa(*e.StatusCode)
		}
		return ""
	}
	return *e.Message
}
