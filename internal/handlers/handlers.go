package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func WebhookRequestHandler(w http.ResponseWriter, r *http.Request) {
	response := handleRequest(r)

	if response.statusCode == http.StatusNoContent {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.statusCode)
	json.NewEncoder(w).Encode(response)

}

func handleRequest(r *http.Request) response {
	query := r.URL.Query()
	fmt.Printf("conn: %+v\n", r)
	debugString := "accepting a connection: "
	debugStringElements := []string{}
	for k := range query {
		if k == "key" {
			continue
		}
		debugStringElements = append(debugStringElements, fmt.Sprintf("%s: %+v", k, query.Get(k)))
	}
	debugString += strings.Join(debugStringElements, ", ")
	fmt.Println(debugString)

	requestType := query.Get("type")
	accessKey := query.Get("key")
	// TODO(rufus): implement proper access management
	authorized := accessKey == "test-key"

	if !authorized && (accessKey != "") {
		fmt.Printf("invalid key authorization attempted: got %v\n", accessKey)
	}

	data, err := url.QueryUnescape(query.Get("data"))
	if err != nil {
		err := fmt.Errorf("failed to url decode data: %w", err)
		fmt.Println(err)
		return GetBadRequestResponse(codeMalformedData, err.Error())
	}

	var resp response
	switch requestType {
	case "ooc":
		resp, err = handleOOC(data, authorized)
	case "":
		return WelcomeResponse
	default:
		return GetBadRequestResponse(codeUnknownRequestType, fmt.Sprintf("Webhook requests of type `%s` are not supported", requestType))
	}

	if err != nil {
		fmt.Println(err)
	}
	return resp

}

func handleOOC(data string, authorized bool) (response, error) {
	if !authorized {
		return ForbiddenResponse, nil
	}
	if data == "" {
		return GetBadRequestResponse(codeEmptyData, "A request `data` was expected, but is missing"), nil
	}
	var requestData oocRequestData
	err := json.Unmarshal([]byte(data), &requestData)
	if err != nil {
		err := fmt.Errorf("failed to unmarshal data: %w", err)
		return GetBadRequestResponse(codeMalformedData, err.Error()), err
	}
	if requestData.Ckey == "" || requestData.Message == "" {
		return GetBadRequestResponse(codeMalformedData, "Both `ckey` and `message` are required in the `data`"), nil
	}
	fmt.Printf("%+v\n", requestData)
	return SuccessResponse, nil
}
