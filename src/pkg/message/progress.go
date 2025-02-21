// Package message provides a rich set of functions for displaying messages to the user.
package message

import (
	"fmt"
)

const padding = "    "

// ProgressBar is a struct used to drive a pterm ProgressbarPrinter.
type ProgressBar struct {
	total     int64
	current   int64
	startText string
}

// NewProgressBar creates a new ProgressBar instance from a total value and a format.
func NewProgressBar(total int64, text string) *ProgressBar {
	if NoProgress {
		Info(text)
	} else {
		fmt.Printf("%s%s\n", padding, text)
	}

	return &ProgressBar{
		total:     total,
		startText: text,
	}
}

// Update updates the ProgressBar with completed progress and new text.
func (p *ProgressBar) Update(complete int64, text string) {
	if NoProgress {
		debugPrinter(2, text)
		return
	}
	p.current = complete
	fmt.Printf("\r%s%s [%d/%d]", padding, text, p.current, p.total)
}

// UpdateTitle updates the ProgressBar with new text.
func (p *ProgressBar) UpdateTitle(text string) {
	if NoProgress {
		debugPrinter(2, text)
		return
	}
	fmt.Printf("\r%s%s", padding, text)
}

// Add updates the ProgressBar with completed progress.
func (p *ProgressBar) Add(n int) {
	p.current += int64(n)
	if p.current > p.total {
		p.current = p.total
	}
	fmt.Printf("\r%s [%d/%d]", p.startText, p.current, p.total)
}

// Write updates the ProgressBar with the number of bytes in a buffer as the completed progress.
func (p *ProgressBar) Write(data []byte) (int, error) {
	n := len(data)
	p.Add(n)
	return n, nil
}

// Successf marks the ProgressBar as successful in the CLI.
func (p *ProgressBar) Successf(format string, a ...any) {
	p.Stop()
	fmt.Printf(format, a...)
}

// Stop stops the ProgressBar from continuing.
func (p *ProgressBar) Stop() {
	fmt.Println()
}

// Errorf marks the ProgressBar as failed in the CLI.
func (p *ProgressBar) Errorf(err error, format string, a ...any) {
	p.Stop()
	WarnErrf(err, format, a...)
}
