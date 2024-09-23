package webhooks_server

import (
	"net/http"
	"strings"
)

type response struct {
	statusCode int    `json:"-"`
	Code       string `json:"code"`
	Details    string `json:"details,omitempty"`
}

const (
	codeWelcome             = "welcome"
	codeSuccess             = "success"
	codeBadAccessKey        = "bad_access_key"
	codeEmptyData           = "empty_data"
	codeMalformedData       = "malformed_data"
	codeUnknownRequestType  = "unknown_request_type"
	codeWebhookDisabled     = "webhook_disabled"
	codeQueueFull           = "queue_full"
	codeInternalServerError = "internal_server_error"
)

var (
	WelcomeResponse = response{
		statusCode: http.StatusOK,
		Code:       codeWelcome,
		Details:    "Welcome to discord-bot for ZeroOnyx, for usage instructions see https://github.com/ZeroHubProjects/discord-bot/blob/master/README.md",
	}
	SuccessResponse = response{
		statusCode: http.StatusOK,
		Code:       codeSuccess,
	}
	ForbiddenResponse = response{
		statusCode: http.StatusForbidden,
		Code:       codeBadAccessKey,
		Details:    "An access `key` is required, but is invalid or missing",
	}
	InternalErrorResponse = response{
		statusCode: http.StatusInternalServerError,
		Code:       codeInternalServerError,
		Details:    "An internal server error has occured. Please contact the maintainers or submit an issue at https://github.com/ZeroHubProjects/discord-bot/issues",
	}
)

func getResponse(statusCode int, code string, details ...string) response {
	return response{
		statusCode: statusCode,
		Code:       code,
		Details:    strings.Join(details, "; "),
	}
}
