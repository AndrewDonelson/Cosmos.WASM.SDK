package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	DEFAULT_API_HOST = "http://localhost:1317"
)

// consumeAPI consumes and API endpoint and returns the response as a map
// Scope: Private
// Parameters:
//
//	url: The URL of the API
//	endpoint: The endpoint to consume
//
// Returns:
//
//	map[string]interface{}: The response as a map
//	error: Any error that occurred
func consumeAPI(url string, endpoint string) (map[string]interface{}, error) {
	fmt.Println("Request:", url, endpoint)

	epURL := url + endpoint

	// Make an HTTP GET request to the API endpoint
	response, err := http.Get(epURL)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP GET request: %v", err)
	}
	defer response.Body.Close()

	// Check that the HTTP status is 200 OK
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %d", response.StatusCode)
	}

	// Check that the content type is JSON
	contentType := response.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}

	// Read the response body and unmarshal it into a map
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	fmt.Println("Response:", PrettyPrint(data))
	return data, nil
}
