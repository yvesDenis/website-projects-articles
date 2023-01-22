package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleRequest(t *testing.T) {

	t.Run("Check if the order sent status is ok then the error message is empty", func(t *testing.T) {
		RandInt = func(n int) int {
			return 0
		}
		var event InEvent
		response := HandleRequest(event)

		fmt.Printf("response : %v", response)

		assert.Empty(t, response.ErrorMessage)
		assert.Equal(t, response.Status, "ok")
	})

	t.Run("Case 1 : If the order sent status is error then the error message is 'could not contact restaurant'", func(t *testing.T) {
		RandInt = func(n int) int {
			if n <= 2 {
				return 1
			} else {
				return 0
			}
		}
		var event InEvent
		response := HandleRequest(event)

		fmt.Printf("response : %v", response)

		assert.Equal(t, response.ErrorMessage, "could not contact restaurant")
		assert.Equal(t, response.Status, "error")
	})

	t.Run("Case 2 : If the order sent status is error then the error message is 'could not understand order'", func(t *testing.T) {
		RandInt = func(n int) int {
			return 1
		}
		var event InEvent
		response := HandleRequest(event)

		fmt.Printf("response : %v", response)

		assert.Equal(t, response.ErrorMessage, "could not understand order")
		assert.Equal(t, response.Status, "error")
	})

	t.Run("Case 3 : If the order sent status is error then the error message is 'unknown error'", func(t *testing.T) {
		RandInt = func(n int) int {
			if n <= 2 {
				return 1
			} else {
				return 2
			}
		}
		var event InEvent
		response := HandleRequest(event)

		fmt.Printf("response : %v", response)

		assert.Equal(t, response.ErrorMessage, "unknown error")
		assert.Equal(t, response.Status, "error")
	})

	t.Run("HandleRequest works as expected with the real flow", func(t *testing.T) {
		var event InEvent
		response := HandleRequest(event)

		fmt.Printf("response : %v", response)
		assert.NotEmpty(t, response)
	})
}
