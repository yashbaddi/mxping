package cli

import (
	"flag"
	"io"
	"time"

	"github.com/yashbaddi/mxping/internal/verifier"
)

const AppVersion = "0.1.0"

// Options represents the parsed command-line configuration.
type Options struct {
	MailFrom     string
	HeloDomain   string
	SMTPPort     int
	Timeout      time.Duration
	DNSServer    string
	SkipSMTP     bool
	SkipCatchAll bool
	JSON         bool
	NoColor      bool
	Verbose      bool
	Concurrency  int
	ShowVersion  bool
}

// ParseOptions parses command line arguments and returns Options and remaining positional arguments.
func ParseOptions(args []string, output io.Writer) (*Options, []string, error) {
	fs := flag.NewFlagSet("mxping", flag.ContinueOnError)
	fs.SetOutput(output)

	opts := &Options{}

	fs.StringVar(&opts.MailFrom, "from", "check@mxping.local", "Sender address for MAIL FROM probe")
	fs.StringVar(&opts.HeloDomain, "helo", "mxping.local", "HELO/EHLO domain to identify as")
	fs.IntVar(&opts.SMTPPort, "port", 25, "SMTP port to connect to")
	fs.DurationVar(&opts.Timeout, "timeout", 10*time.Second, "Network timeout duration")
	fs.StringVar(&opts.DNSServer, "dns", "", "Custom DNS resolver address (e.g. 8.8.8.8:53)")
	fs.BoolVar(&opts.SkipSMTP, "skip-smtp", false, "Validate syntax and DNS only; skip SMTP connection")
	fs.BoolVar(&opts.SkipCatchAll, "skip-catchall", false, "Skip catch-all probe detection")
	fs.BoolVar(&opts.JSON, "json", false, "Output results in JSON format")
	fs.BoolVar(&opts.NoColor, "no-color", false, "Disable colored terminal output")
	fs.BoolVar(&opts.Verbose, "v", false, "Enable verbose SMTP dialogue logging (printed to stderr)")
	fs.IntVar(&opts.Concurrency, "c", 4, "Concurrency limit for batch email checks")
	fs.BoolVar(&opts.ShowVersion, "version", false, "Print version information and exit")

	if err := fs.Parse(args); err != nil {
		return nil, nil, err
	}

	return opts, fs.Args(), nil
}

// ToVerifierConfig maps parsed CLI options to the verifier configuration.
func (o *Options) ToVerifierConfig(logger func(format string, args ...any)) verifier.Config {
	cfg := verifier.DefaultConfig()
	cfg.MailFrom = o.MailFrom
	cfg.HeloName = o.HeloDomain
	cfg.SMTPPort = o.SMTPPort
	cfg.SMTPTimeout = o.Timeout
	cfg.DNSTimeout = o.Timeout
	cfg.DNSServer = o.DNSServer
	cfg.CheckSMTP = !o.SkipSMTP
	cfg.CheckCatchAll = !o.SkipCatchAll
	cfg.VerboseLog = logger
	return cfg
}

