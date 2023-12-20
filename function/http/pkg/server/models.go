package server

import (
	"encoding/json"
	"fmt"
)

// eventRequestContext represents a request context
type eventRequestContext struct {
	ServiceID string `json:"serviceId"`
	RequestID string `json:"requestId"`
	Method    string `json:"httpMethod"`
	Path      string `json:"path"`
	SourceIP  string `json:"sourceIp"`
	Stage     string `json:"stage"`
	Identity  struct {
		SecretID *string `json:"secretId"`
	} `json:"identity"`
}

// EventRequest represents an API gateway request
type EventRequest struct {
	Headers     map[string]string   `json:"headers"`
	Method      string              `json:"httpMethod"`
	Path        string              `json:"path"`
	QueryString eventQueryString    `json:"queryString"`
	Body        string              `json:"body"`
	Context     eventRequestContext `json:"requestContext"`

	// the following fields are ignored
	// HeaderParameters      interface{} `json:"headerParameters"`
	// PathParameters        interface{} `json:"pathParameters"`
	// QueryStringParameters interface{} `json:"queryStringParameters"`
}

// EventResponse represents an API gateway response
type EventResponse struct {
	IsBase64Encoded bool              `json:"isBase64Encoded"`
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
}

// eventQueryString represents query string of an API gateway request
type eventQueryString map[string][]string

// UnmarshalJSON implements the json.Unmarshaller interface,
// it handles the query string properly
func (qs *eventQueryString) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	r := make(eventQueryString)
	for k, v := range m {
		switch v.(type) {
		case bool:
			r[k] = []string{}
		case string:
			r[k] = []string{v.(string)}
		case []string:
			r[k] = v.([]string)
		case []interface{}:
			vs := v.([]interface{})
			for _, sv := range vs {
				s, ok := sv.(string)
				if !ok {
					return fmt.Errorf("unexpected query string value: %+v, type: %T", v, v)
				}
				r[k] = append(r[k], s)
			}
		default:
			return fmt.Errorf("unexpected query string value: %+v, type: %T", v, v)
		}
	}
	*qs = r
	return nil
}
