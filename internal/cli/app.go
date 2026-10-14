package cli

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/yashbaddi/mxping/internal/output"
	"github.com/yashbaddi/mxping/internal/verifier"
)

// Exit codes
const (
	ExitSuccess       = 0
	ExitInvalidEmail  = 1
	ExitErrorOrBlocked = 2
)

// App manages the execution lifecycle of the mxping CLI.
type App struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

// NewApp creates a new App configured with standard operating system streams.
func NewApp() *App {
	return &App{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}
}

// Run executes the application logic based on arguments and returns an exit code.
func (a *App) Run(ctx context.Context, args []string) int {
	opts, positionalArgs, err := ParseOptions(args, a.Stderr)
	if err != nil {
		return ExitErrorOrBlocked
	}

	if opts.ShowVersion {
		fmt.Fprintf(a.Stdout, "mxping version %s\n", AppVersion)
		return ExitSuccess
	}

	emails, err := ReadEmails(positionalArgs, a.Stdin)
	if err != nil {
		fmt.Fprintf(a.Stderr, "Error reading input: %v\n", err)
		return ExitErrorOrBlocked
	}

	if len(emails) == 0 {
		a.printHelp()
		return ExitErrorOrBlocked
	}

	var logger func(format string, args ...any)
	if opts.Verbose {
		l := log.New(a.Stderr, "[SMTP] ", log.LstdFlags)
		logger = func(format string, args ...any) {
			l.Printf(format, args...)
		}
	}

	cfg := opts.ToVerifierConfig(logger)
	v := verifier.NewVerifier(cfg)
	formatter := output.NewFormatter(a.Stdout, opts.JSON, opts.NoColor)

	if len(emails) == 1 {
		return a.runSingle(ctx, v, formatter, emails[0])
	}

	return a.runBatch(ctx, v, formatter, emails, opts.Concurrency)
}

func (a *App) runSingle(ctx context.Context, v *verifier.Verifier, formatter *output.Formatter, email string) int {
	result := v.Verify(ctx, email)
	formatter.PrintResult(result)

	switch result.Status {
	case verifier.StatusValid:
		return ExitSuccess
	case verifier.StatusBlocked:
		return ExitErrorOrBlocked
	default:
		return ExitInvalidEmail
	}
}

func (a *App) runBatch(ctx context.Context, v *verifier.Verifier, formatter *output.Formatter, emails []string, concurrency int) int {
	if concurrency < 1 {
		concurrency = 1
	}

	type emailJob struct {
		index int
		email string
	}

	jobs := make(chan emailJob, len(emails))
	results := make([]verifier.Result, len(emails))
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				results[job.index] = v.Verify(ctx, job.email)
			}
		}()
	}

	for i, email := range emails {
		jobs <- emailJob{index: i, email: email}
	}
	close(jobs)

	wg.Wait()

	formatter.PrintBatchSummary(results)

	for _, r := range results {
		if r.Status != verifier.StatusValid {
			return ExitInvalidEmail
		}
	}

	return ExitSuccess
}

func (a *App) printHelp() {
	fmt.Fprintf(a.Stderr, "mxping - Fast, accurate email verification & mailbox pinging CLI\n\n")
	fmt.Fprintf(a.Stderr, "Usage:\n")
	fmt.Fprintf(a.Stderr, "  mxping [flags] <email> [email2 ...]\n")
	fmt.Fprintf(a.Stderr, "  cat emails.txt | mxping [flags]\n\n")
	fmt.Fprintf(a.Stderr, "Run 'mxping -h' to see all available flags and options.\n")
}

