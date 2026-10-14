package cli

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// ReadEmails extracts email addresses from command line arguments or standard input.
// If args has emails, it uses those. If args is empty and stdin is piped, it reads line-by-line.
func ReadEmails(args []string, stdin io.Reader) ([]string, error) {
	if len(args) > 0 {
		emails := make([]string, 0, len(args))
		for _, a := range args {
			trimmed := strings.TrimSpace(a)
			if trimmed != "" {
				emails = append(emails, trimmed)
			}
		}
		return emails, nil
	}

	// Check if stdin is a piped file/stream or has data
	if stdin == nil {
		return nil, nil
	}

	if f, ok := stdin.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return nil, err
		}
		// If stdin is a terminal character device and no args were given, don't block
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return nil, nil
		}
	}

	var emails []string
	scanner := bufio.NewScanner(stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			emails = append(emails, line)
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	return emails, nil
}
