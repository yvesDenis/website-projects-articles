package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("non 200 Response found")
)

func TestHandler(t *testing.T) {

	request := events.APIGatewayProxyRequest{
		Body: "{\"user_id\": \"Burger_2\",\"quantity\": \"3\",\"restaurant_id\": \"Restaurant 3\"}",
	}
	t.Run("Unable to get IP", func(t *testing.T) {
		DefaultHTTPGetAddress = "http://127.0.0.1:12345"

		_, err := HandleRequest(request)
		if err == nil {
			t.Fatal("Error failed to trigger with an invalid request")
		}
	})

	t.Run("Non 200 Response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		defer ts.Close()

		DefaultHTTPGetAddress = ts.URL

		_, err := HandleRequest(request)
		if err != nil && err.Error() != ErrNon200Response.Error() {
			t.Fatalf("Error failed to trigger with an invalid HTTP response: %v", err)
		}
	})

	t.Run("Unable decode IP", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		defer ts.Close()

		DefaultHTTPGetAddress = ts.URL

		_, err := HandleRequest(request)
		if err == nil {
			t.Fatal("Error failed to trigger with an invalid HTTP response")
		}
	})

	t.Run("Successful Request", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			fmt.Fprintf(w, "127.0.0.1")
		}))
		defer ts.Close()

		DefaultHTTPGetAddress = ts.URL

		_, err := HandleRequest(request)
		if err != nil {
			t.Fatal("Everything should be ok")
		}
	})
}
