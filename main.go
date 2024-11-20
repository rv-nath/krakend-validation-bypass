package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var (
	pluginName        = "krakend-common-middleware"
	HandlerRegisterer = registerer(pluginName)
)

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
),
) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(_ context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {
	// If the plugin requires some configuration, it should be under the name of the plugin.
	cfg, ok := extra[pluginName].(map[string]interface{})
	if !ok {
		return h, errors.New("configuration not found for common middleware")
	}

	// Extract exceptions list from configuration
	logger.Info("Extracting exception list...")
	exceptions, _ := cfg["exceptions"].([]interface{})
	var exceptionPatterns []string
	for _, url := range exceptions {
		// Convert to string and create regex pattern
		exception := url.(string)
		regexPattern := convertWildcardToRegex(exception)
		exceptionPatterns = append(exceptionPatterns, regexPattern)
	}
	logger.Debug("Configured exceptionPatterns:", exceptionPatterns)

	logger.Debug("Validation-bypass middleware registered")
	return middleware(h, exceptionPatterns), nil
}

// Convert wildcard pattern to regex
func convertWildcardToRegex(path string) string {
	// Replace '*' with '.*' to match any value
	path = strings.ReplaceAll(path, "*", ".*")
	return "^" + path + "$"
}

// Middleware function to mark requests for bypass if they match an exception pattern
func middleware(next http.Handler, exceptionPatterns []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info(fmt.Sprintf("[PLUGIN: %s] Common Middleware executing..", HandlerRegisterer))

		// Check if the request URL matches any of the exception patterns
		for _, pattern := range exceptionPatterns {
			matched, _ := regexp.MatchString(pattern, r.URL.Path)
			if matched {
				// Add a bypass flag to the request context
				ctx := context.WithValue(r.Context(), "bypassValidation", true)
				r = r.WithContext(ctx)
				logger.Info("[PLUGIN: Common Middleware] Request bypassed due to matching exception pattern")
				next.ServeHTTP(w, r)
				return
			}
		}

		// Proceed with the next handler if not bypassed
		next.ServeHTTP(w, r)
	})
}

func main() {}

// This logger is replaced by the RegisterLogger method to load the one from krakenD.
var logger Logger = noopLogger{}

func (registerer) RegisterLogger(v interface{}) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
}

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}

// Empty logger implementation
type noopLogger struct{}

func (n noopLogger) Debug(_ ...interface{})    {}
func (n noopLogger) Info(_ ...interface{})     {}
func (n noopLogger) Warning(_ ...interface{})  {}
func (n noopLogger) Error(_ ...interface{})    {}
func (n noopLogger) Critical(_ ...interface{}) {}
func (n noopLogger) Fatal(_ ...interface{})    {}
