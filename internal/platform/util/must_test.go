package util

import (
	"errors"
	"testing"
)

func TestMust_Success(t *testing.T) {
	got := Must(42, nil)
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestMust_PanicsOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, got nil")
		}
	}()
	_ = Must(0, errors.New("boom"))
}
