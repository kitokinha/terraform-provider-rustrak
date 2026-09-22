package client

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIErrorClassification(t *testing.T) {
	for _, tt := range []struct {
		name               string
		err                error
		notFound, conflict bool
	}{
		{"nil", nil, false, false},
		{"ordinary", errors.New("404 not found"), false, false},
		{"missing", &APIError{StatusCode: 404}, true, false},
		{"wrapped missing", fmt.Errorf("read: %w", &APIError{StatusCode: 404}), true, false},
		{"conflict", &APIError{StatusCode: 409}, false, true},
		{"wrapped conflict", fmt.Errorf("create: %w", &APIError{StatusCode: 409}), false, true},
		{"server error", &APIError{StatusCode: 500}, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if IsNotFound(tt.err) != tt.notFound || IsConflict(tt.err) != tt.conflict {
				t.Fatalf("incorrect classification of %v", tt.err)
			}
		})
	}
}
