package utils

import (
	"context"
	"time"
)

// NewContextWithTimeout creates new context with timeout/*
func NewContextWithTimeout(inSec time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), inSec)
}
