package db

import (
	"context"
	"testing"
	"time"
)

func TestNew_InvalidConnString(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := New(ctx, "invalid://")
	if err == nil {
		t.Fatal("expected error for invalid connection string")
	}
}
