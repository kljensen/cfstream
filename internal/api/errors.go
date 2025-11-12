package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cloudflare/cloudflare-go/v3"
)

var (
	// ErrNotFound is returned when a video is not found (404).
	ErrNotFound = errors.New("video not found")

	// ErrUnauthorized is returned when authentication fails (401).
	ErrUnauthorized = errors.New("unauthorized: invalid API token or account ID")

	// ErrForbidden is returned when access is forbidden (403).
	ErrForbidden = errors.New("forbidden: insufficient permissions")

	// ErrRateLimit is returned when rate limited (429).
	ErrRateLimit = errors.New("rate limit exceeded: please wait before retrying")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")
)

// WrapError converts Cloudflare SDK errors into user-friendly errors.
func WrapError(err error) error {
	if err == nil {
		return nil
	}

	// Check if it's a Cloudflare API error
	var apiErr *cloudflare.Error
	if errors.As(err, &apiErr) {
		return wrapAPIError(apiErr)
	}

	// Return original error if not a Cloudflare error
	return err
}

// wrapAPIError converts Cloudflare API errors based on status code.
func wrapAPIError(apiErr *cloudflare.Error) error {
	statusCode := apiErr.StatusCode

	// Get the original error message safely
	// Use defer/recover to handle potential panics from apiErr.Error()
	var errMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				errMsg = ""
			}
		}()
		errMsg = apiErr.Error()
	}()

	switch statusCode {
	case http.StatusNotFound:
		if errMsg != "" {
			return fmt.Errorf("%w: %s", ErrNotFound, errMsg)
		}
		return fmt.Errorf("%w", ErrNotFound)
	case http.StatusUnauthorized:
		if errMsg != "" {
			return fmt.Errorf("%w: %s", ErrUnauthorized, errMsg)
		}
		return fmt.Errorf("%w", ErrUnauthorized)
	case http.StatusForbidden:
		if errMsg != "" {
			return fmt.Errorf("%w: %s", ErrForbidden, errMsg)
		}
		return fmt.Errorf("%w", ErrForbidden)
	case http.StatusTooManyRequests:
		if errMsg != "" {
			return fmt.Errorf("%w: %s", ErrRateLimit, errMsg)
		}
		return fmt.Errorf("%w", ErrRateLimit)
	case http.StatusBadRequest:
		if errMsg != "" {
			return fmt.Errorf("%w: %s", ErrInvalidInput, errMsg)
		}
		return fmt.Errorf("%w", ErrInvalidInput)
	default:
		if errMsg != "" {
			return fmt.Errorf("API error (status %d): %s", statusCode, errMsg)
		}
		return fmt.Errorf("API error (status %d)", statusCode)
	}
}
