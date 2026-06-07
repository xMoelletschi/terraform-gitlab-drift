package cmd

import (
	"context"
	"testing"
	"time"
)

func TestWithTimeout_ZeroMeansNoDeadline(t *testing.T) {
	ctx, cancel := withTimeout(context.Background(), 0)
	defer cancel()
	if _, ok := ctx.Deadline(); ok {
		t.Error("expected no deadline for zero timeout")
	}
}

func TestWithTimeout_PositiveSetsDeadline(t *testing.T) {
	ctx, cancel := withTimeout(context.Background(), time.Minute)
	defer cancel()
	if _, ok := ctx.Deadline(); !ok {
		t.Error("expected a deadline for a positive timeout")
	}
}
