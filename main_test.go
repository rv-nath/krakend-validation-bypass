package main

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

// TestConvertToRegex checks that {var} patterns are correctly converted into regex patterns.
func TestConvertToRegex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/v1/profile/{accid}/setpassword", "^/v1/profile/[^/]+/setpassword$"},
		{"/v1/login", "^/v1/login$"},
		{"/v1/verifyaccount/{userid}", "^/v1/verifyaccount/[^/]+$"},
	}

	for _, test := range tests {
		result := convertToRegex(test.input)
		if result != test.expected {
			t.Errorf("convertToRegex(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

// TestMiddlewareBypass checks if the middleware correctly sets the bypass flag for matching exceptions.
func TestMiddlewareBypass(t *testing.T) {
	exceptionPatterns := []string{
		"^/v1/profile/[^/]+/setpassword$",
		"^/v1/login$",
	}

	// Create an instance of the middleware
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify if the context has the bypassValidation flag set
		bypass, ok := r.Context().Value("bypassValidation").(bool)
		if !ok || !bypass {
			t.Errorf("expected bypassValidation flag to be true; got false")
		}
		w.WriteHeader(http.StatusOK)
	}), exceptionPatterns)

	// Create a request that should match the bypass pattern
	req := httptest.NewRequest("POST", "/v1/profile/12345/setpassword", nil)
	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	// Check the response code
	if w.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", w.Code)
	}
}

// TestMiddlewareNoBypass checks if the middleware correctly proceeds without bypassing for non-matching URLs.
func TestMiddlewareNoBypass(t *testing.T) {
	exceptionPatterns := []string{
		"^/v1/profile/[^/]+/setpassword$",
		"^/v1/login$",
	}

	// Create an instance of the middleware
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify if the context does not have the bypassValidation flag set
		if bypass, ok := r.Context().Value("bypassValidation").(bool); ok && bypass {
			t.Errorf("expected bypassValidation flag to be false; got true")
		}
		w.WriteHeader(http.StatusOK)
	}), exceptionPatterns)

	// Create a request that should not match the bypass pattern
	req := httptest.NewRequest("POST", "/v1/profile/updatepassword", nil)
	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	// Check the response code
	if w.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", w.Code)
	}
}

// TestPatternMatching checks if regex patterns match the expected URLs.
func TestPatternMatching(t *testing.T) {
	patterns := []string{
		"^/v1/profile/[^/]+/setpassword$",
		"^/v1/login$",
	}

	tests := []struct {
		url         string
		shouldMatch bool
	}{
		{"/v1/profile/12345/setpassword", true},
		{"/v1/login", true},
		{"/v1/profile/updatepassword", false},
		{"/v1/profile/12345/updatepassword", false},
	}

	for _, test := range tests {
		matched := false
		for _, pattern := range patterns {
			if ok, _ := regexp.MatchString(pattern, test.url); ok {
				matched = true
				break
			}
		}
		if matched != test.shouldMatch {
			t.Errorf("for URL %q, expected match: %v, got: %v", test.url, test.shouldMatch, matched)
		}
	}
}
