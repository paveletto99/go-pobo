package samplejob

import (
	"context"
	"strings"
	"testing"
)

func TestRunnerRunOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		runner := &Runner{
			config:  &Config{WorkItems: 3},
			metrics: nil,
		}
		if err := runner.RunOnce(ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("configured failure", func(t *testing.T) {
		t.Parallel()

		runner := &Runner{
			config:  &Config{Fail: true},
			metrics: nil,
		}
		if err := runner.RunOnce(ctx); err == nil || !strings.Contains(err.Error(), "configured to fail") {
			t.Fatalf("got %v, want configured failure", err)
		}
	})
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	if err := (&Config{WorkItems: -1}).Validate(); err == nil {
		t.Fatal("expected validation error")
	}
	if err := (&Config{WorkItems: 0}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestNewRunnerValidation(t *testing.T) {
	t.Parallel()

	_, err := NewRunner(&Config{WorkItems: -1})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
