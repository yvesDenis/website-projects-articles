package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const responseMessageExpected = "Hello, i'm a base app!\n"

func TestHelloFunction(t *testing.T) {

	t.Run("Return a hello message ", func(t *testing.T) {

		request, _ := http.NewRequest(http.MethodGet, "hello", nil)
		response := httptest.NewRecorder()

		Hello(response, request)

		responseMessageReceived := response.Body.String()

		if responseMessageExpected != responseMessageReceived {
			t.Errorf("received %q but wanted %q", responseMessageReceived, responseMessageExpected)
		}
	})
}
