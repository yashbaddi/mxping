package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestApp_Run_Version(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	app := &App{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
	}

	exitCode := app.Run(context.Background(), []string{"-version"})
	if exitCode != ExitSuccess {
		t.Fatalf("expected ExitSuccess (0), got %d", exitCode)
	}

	if !strings.Contains(stdout.String(), "mxping version") {
		t.Errorf("expected version output, got: %s", stdout.String())
	}
}

func TestApp_Run_NoArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	app := &App{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
	}

	exitCode := app.Run(context.Background(), []string{})
	if exitCode != ExitErrorOrBlocked {
		t.Fatalf("expected ExitErrorOrBlocked (2), got %d", exitCode)
	}

	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected usage message in stderr, got: %s", stderr.String())
	}
}

func TestApp_Run_SingleInvalid(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	app := &App{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
	}

	exitCode := app.Run(context.Background(), []string{"-skip-smtp", "invalid-email"})
	if exitCode != ExitInvalidEmail {
		t.Fatalf("expected ExitInvalidEmail (1), got %d", exitCode)
	}

	if !strings.Contains(stdout.String(), "INVALID") {
		t.Errorf("expected INVALID in output, got: %s", stdout.String())
	}
}

