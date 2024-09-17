package handlers

import "net/http"

type response struct {
	statusCode int    `json:"-"`
	Code       string `json:"code"`
	Details    string `json:"details,omitempty"`
}

const (
	codeWelcome            = "welcome"
	codeSuccess            = "success"
	codeEmptyData          = "empty_data"
	codeMalformedData      = "malformed_data"
	codeUnknownRequestType = "unknown_request_type"
)

var (
	WelcomeResponse = response{
		statusCode: http.StatusOK,
		Code:       "welcome",
		Details:    "Welcome to discord-bot for ZeroOnyx, for usage instructions see https://github.com/ZeroHubProjects/discord-bot/blob/master/README.md",
	}
	SuccessResponse = response{
		statusCode: http.StatusOK,
		Code:       "success",
	}
	NoContentResponse = response{
		statusCode: http.StatusNoContent,
	}
	ForbiddenResponse = response{
		statusCode: http.StatusForbidden,
		Code:       "bad_access_key",
		Details:    "An access `key` is required, but is invalid or missing",
	}
	InternalErrorResponse = response{
		statusCode: http.StatusInternalServerError,
		Code:       "internal_server_error",
		Details:    "An internal server error has occured. Please contact the maintainers or submit an issue at https://github.com/ZeroHubProjects/discord-bot/issues",
	}
)

func GetBadRequestResponse(code, details string) response {
	return response{
		statusCode: http.StatusBadRequest,
		Code:       code,
		Details:    details,
	}
}
