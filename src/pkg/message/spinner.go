// Package message provides a rich set of functions for displaying messages to the user.
package message

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"time"
)

var activeSpinner *Spinner

var sequence = []string{`  ⠋ `, `  ⠙ `, `  ⠹ `, `  ⠸ `, `  ⠼ `, `  ⠴ `, `  ⠦ `, `  ⠧ `, `  ⠇ `, `  ⠏ `}

// Spinner is a wrapper around pterm.SpinnerPrinter.
type Spinner struct {
	text           string
	stopChan       chan struct{}
	preserveWrites bool
}

// NewProgressSpinner creates a new progress spinner.
func NewProgressSpinner(format string, a ...any) *Spinner {
	if activeSpinner != nil {
		activeSpinner.Updatef(format, a...)
		debugPrinter(2, "Active spinner already exists")
		return activeSpinner
	}

	text := fmt.Sprintf(format, a...)
	if NoProgress {
		Info(text)
	} else {
		fmt.Print(text)
	}

	spinner := &Spinner{
		text:     text,
		stopChan: make(chan struct{}),
	}

	go spinner.start()

	activeSpinner = spinner
	return spinner
}

// EnablePreserveWrites enables preserving writes to the terminal.
func (p *Spinner) EnablePreserveWrites() {
	p.preserveWrites = true
}

// DisablePreserveWrites disables preserving writes to the terminal.
func (p *Spinner) DisablePreserveWrites() {
	p.preserveWrites = false
}

// Write the given text to the spinner.
func (p *Spinner) Write(raw []byte) (int, error) {
	size := len(raw)
	if NoProgress {
		if p.preserveWrites {
			fmt.Println(string(raw))
		}

		return size, nil
	}

	// Split the text into lines and update the spinner for each line.
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		// Only be fancy if preserve writes is enabled.
		if p.preserveWrites {
			text := fmt.Sprintf("     %s", scanner.Text())
			fmt.Println(strings.Repeat(" ", TermWidth))
			fmt.Println(text)
		} else {
			// Otherwise just update the spinner text.
			p.text = scanner.Text()
		}
	}

	return size, nil
}

// Updatef updates the spinner text.
func (p *Spinner) Updatef(format string, a ...any) {
	if NoProgress {
		debugPrinter(2, fmt.Sprintf(format, a...))
		return
	}
	p.text = fmt.Sprintf(format, a...)
}

// Stop the spinner.
func (p *Spinner) Stop() {
	select {
	case <-p.stopChan:
		// Channel is already closed, do nothing
	default:
		close(p.stopChan)
		fmt.Println()
		activeSpinner = nil
	}
}

// Success prints a success message and stops the spinner.
func (p *Spinner) Success() {
	p.Successf("%s", p.text)
}

// Successf prints a success message with the spinner and stops it.
func (p *Spinner) Successf(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	fmt.Println(text)
	p.Stop()
}

// Warnf prints a warning message with the spinner.
func (p *Spinner) Warnf(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	fmt.Println("WARNING:", text)
}

// Errorf prints an error message with the spinner.
func (p *Spinner) Errorf(err error, format string, a ...any) {
	p.Warnf(format, a...)
	debugPrinter(2, err)
}

// Fatal calls message.Fatalf with the given error.
func (p *Spinner) Fatal(err error) {
	p.Fatalf(err, "%s", p.text)
}

// Fatalf calls message.Fatalf with the given error and format.
func (p *Spinner) Fatalf(err error, format string, a ...any) {
	p.Stop()
	Fatalf(err, format, a...)
}

// Pause the spinner.
func (p *Spinner) Pause() string {
	p.Stop()
	return p.text
}

func (p *Spinner) start() {
	i := 0
	for {
		select {
		case <-p.stopChan:
			return
		default:
			fmt.Printf("\r%s %s", p.text, sequence[i%len(sequence)])
			time.Sleep(100 * time.Millisecond)
			i++
		}
	}
}
