package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

type ErrorFields map[string]string // Error field-value pair type

type ProxyError struct {
	Msg    string      `json:"message"`
	Status int         `json:"status"`
	Data   ErrorFields // For extra error fields e.g. reason, details, etc.
}

// AddErrField adds a new field to the proxy error with given key and value
func (err *ProxyError) AddErrField(key, value string) {
	if err.Data == nil {
		err.Data = make(ErrorFields)
	}
	err.Data[key] = value
}

// RemoveErrField removes existing field matching given key from proxy error
func (err *ProxyError) RemoveErrField(key string) {
	delete(err.Data, key)
}

// MarshalJSON marshals the proxy error into json
func (err *ProxyError) MarshalJSON() ([]byte, error) {
	// Determine json field name for error message
	errType := reflect.TypeOf(*err)
	msgField, ok := errType.FieldByName("Msg")
	msgJsonName := "message"
	if ok {
		msgJsonTag := msgField.Tag.Get("json")
		if msgJsonTag != "" {
			msgJsonName = msgJsonTag
		}
	}
	// Determine json field name for error status code
	statusField, ok := errType.FieldByName("Status")
	statusJsonName := "status"
	if ok {
		statusJsonTag := statusField.Tag.Get("json")
		if statusJsonTag != "" {
			statusJsonName = statusJsonTag
		}
	}
	fieldMap := make(map[string]string)
	fieldMap[msgJsonName] = err.Msg
	fieldMap[statusJsonName] = fmt.Sprintf("%d", err.Status)
	for key, value := range err.Data {
		fieldMap[key] = value
	}
	return json.Marshal(fieldMap)
}

// SerializeJSON converts proxy error into serialized json string. In case of serialization failure, empty string is returned.
func (pxyErr *ProxyError) SerializeJSON() string {
	value, err := json.Marshal(pxyErr)
	if err != nil {
		return ""
	}
	return string(value)
}

// Error implements error interface
func (pxyErr *ProxyError) Error() string {
	return pxyErr.SerializeJSON()
}

// Error creates new proxy error with given message and status code.
func Error(msg string, status int) *ProxyError {
	return &ProxyError{
		Msg:    msg,
		Status: status,
	}
}

// SendErrorResponse writes the serialized error in the server response.
func SendErrorResponse(rw http.ResponseWriter, err error) {
	if proxyErr, ok := err.(*ProxyError); ok {
		rw.Header().Add("content-type", "application/json")
		rw.WriteHeader(proxyErr.Status)
	} else {
		rw.Header().Add("content-type", "text/plain")
		rw.WriteHeader(http.StatusInternalServerError)
	}
	rw.Write([]byte(err.Error()))
}

// Specific Proxy Errors

func ErrorNotFound(reason string) *ProxyError {
	err := Error("err: not found", http.StatusNotFound)
	if reason != "" {
		err.AddErrField("reason", reason)
	}
	return err
}

func ErrorContainerUrlMalformed(rawUrl string) *ProxyError {
	err := Error("err: container url is malformed", http.StatusInternalServerError)
	if rawUrl != "" {
		err.AddErrField("containerRawUrl", rawUrl)
	}
	return err
}
