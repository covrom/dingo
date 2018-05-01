package handler

import (
	"encoding/json"
	"strings"

	"github.com/dinever/golf"
	"github.com/covrom/dingo/app/utils"
)

type APISerializeable interface {
	Serialize() []byte
}

type APIStatusJSON struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type APIResponseBodyJSON struct {
	Data   interface{}   `json:"data"`
	Status APIStatusJSON `json:"status"`
}

func NewErrorStatusJSON(msgs ...string) APIStatusJSON {
	msg := strings.Join(msgs, " ")
	return APIStatusJSON{Status: "error", Message: msg}
}

func NewSuccessStatusJSON(msgs ...string) APIStatusJSON {
	msg := strings.Join(msgs, " ")
	return APIStatusJSON{Status: "success", Message: msg}
}

func NewAPISuccessResponse(data interface{}, msgs ...string) APIResponseBodyJSON {
	status := NewSuccessStatusJSON(msgs...)
	return APIResponseBodyJSON{Data: data, Status: status}
}

// APIDocumentationHandler shows which routes match with what functionality,
// similar to https://api.github.com
func APIDocumentationHandler(routes map[string]map[string]interface{}) golf.HandlerFunc {
	return func(ctx *golf.Context) {
		routes["GET"]["api_documentation_url"] = "/api"
		ctx.JSONIndent(map[string]interface{}{"request_method": routes}, "", "  ")
	}
}

// handleErr sends the staus code and error message formatted as JSON.
func handleErr(ctx *golf.Context, statusCode int, err error) {
	ctx.JSONIndent(map[string]interface{}{
		"statusCode": statusCode,
		"error":      err.Error(),
	}, "", "  ")
}

func (status APIStatusJSON) Serialize() []byte {
	serializedStatus, err := json.Marshal(status)
	if err != nil {
		utils.LogOnError(err, "Unable to serialize status.", true)
		return []byte("")
	}
	return serializedStatus
}

func (body APIResponseBodyJSON) Serialize() []byte {
	serializedBody, err := json.Marshal(body)
	if err != nil {
		utils.LogOnError(err, "Unable to serialize response body.", true)
		return []byte("")
	}
	return serializedBody
}
