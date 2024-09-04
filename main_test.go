package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Integration tests to simulate high traffic scenarios and verify that the rate
// limiter behaves as expected.
func TestRateLimiterIntegration(t *testing.T) {

	handler := http.Handler(PerClientRateLimiter(UserHandler, IncommingLimit{5, 5}))
	for i := 1; i <= 6; i++ {
		req, err := http.NewRequest("GET", "/user/1/data", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if i <= 5 && rr.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i, rr.Code)
		}

		// For the 6th, we expect a 429 Too Many Requests status
		if i > 5 && rr.Code != http.StatusTooManyRequests {
			t.Errorf("Request %d: expected status 429, got %d", i, rr.Code)
		}
	}
}

// validate rate-limiting middleware
func TestPerClientRateLimiter(t *testing.T) {
	nextHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	incomming := IncommingLimit{limit: 5, brust: 5}
	handler := PerClientRateLimiter(nextHandler, incomming)

	req1 := httptest.NewRequest("GET", "/user/2/data", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr1.Code)
	}

	if rr1.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", rr1.Body.String())
	}
}

// validate the token bucket implementation
func TestLimiterAllow(t *testing.T) {
	// Test case 1: Infinite limit should always allow
	limiter := NewLimiter(Inf, 1)
	if !limiter.Allow() {
		t.Errorf("Expected Allow to return true for infinite limit, got false")
	}

	// Test case 1: Zero limit should allow if burst tokens are available
	limiter = NewLimiter(0, 1)
	if !limiter.Allow() {
		t.Errorf("Expected Allow to return true for zero limit with available burst, got false")
	}
	// After consuming burst token, the next request should be denied
	if limiter.Allow() {
		t.Errorf("Expected Allow to return false after consuming burst token, got true")
	}

	// Test case 4: Rate-limited case with tokens available
	limiter = NewLimiter(1, 2)
	limiter.tokens = 2

	// First two requests should be allowed
	if !limiter.Allow() {
		t.Errorf("Expected Allow to return true for first request, got false")
	}
	if !limiter.Allow() {
		t.Errorf("Expected Allow to return true for second request, got false")
	}

	// Next request should be denied as tokens are exhausted
	if limiter.Allow() {
		t.Errorf("Expected Allow to return false after tokens are exhausted, got true")
	}
}
