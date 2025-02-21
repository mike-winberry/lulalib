package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

type LogLevel int

const (
	WarnLevel LogLevel = iota
	InfoLevel
	DebugLevel
	TraceLevel

	TermWidth = 100
)

var NoProgress bool
var RuleLine = strings.Repeat("━", TermWidth)
var LogWriter io.Writer = os.Stderr
var logLevel = InfoLevel
var logFile *os.File
var useLogFile bool

type DebugWriter struct{}

func (d *DebugWriter) Write(raw []byte) (int, error) {
	debugPrinter(2, string(raw))
	return len(raw), nil
}

func init() {
	SetDefaultOutput(os.Stderr)
}

func SetDefaultOutput(w io.Writer) {
	LogWriter = w
}

func UseLogFile(inputLogFile *os.File) {
	ts := time.Now().Format("2006-01-02-15-04-05")

	var err error
	if inputLogFile != nil {
		LogWriter = io.MultiWriter(inputLogFile)
		SetDefaultOutput(LogWriter)
	} else {
		if logFile, err = os.CreateTemp("", fmt.Sprintf("lula-%s-*.log", ts)); err != nil {
			WarnErr(err, "Error saving a log file to a temporary directory")
		} else {
			useLogFile = true
			LogWriter = io.MultiWriter(os.Stderr, logFile)
			SetDefaultOutput(LogWriter)
			message := fmt.Sprintf("Saving log file to %s", logFile.Name())
			Note(message)
		}
	}
}

func UseBuffer(buf *bytes.Buffer) {
	LogWriter = io.MultiWriter(buf)
	SetDefaultOutput(LogWriter)
}

func SetLogLevel(lvl LogLevel) {
	logLevel = lvl
}

func GetLogLevel() LogLevel {
	return logLevel
}

func DisableColor() {
	// No-op for now, as fmt doesn't support color
}

// Debug prints a debug message.
func Debug(payload ...any) {
	if logLevel >= DebugLevel {
		message := fmt.Sprint(payload...)
		fmt.Fprintln(LogWriter, "DEBUG:", message)
	}
}

// Debugf prints a debug message with a given format.
func Debugf(format string, a ...any) {
	if logLevel >= DebugLevel {
		message := fmt.Sprintf(format, a...)
		fmt.Fprintln(LogWriter, "DEBUG:", message)
	}
}

func ErrorWebf(err any, w http.ResponseWriter, format string, a ...any) {
	debugPrinter(2, err)
	message := fmt.Sprintf(format, a...)
	Warn(message)
	http.Error(w, message, http.StatusInternalServerError)
}

func Warn(message string) {
	Warnf("%s", message)
}

func Warnf(format string, a ...any) {
	message := Paragraphn(TermWidth-10, format, a...)
	fmt.Fprintln(LogWriter)
	fmt.Fprintln(LogWriter, "WARNING:", message)
}

func WarnErr(err any, message string) {
	debugPrinter(2, err)
	Warnf("%s", message)
}

func WarnErrf(err any, format string, a ...any) {
	debugPrinter(2, err)
	Warnf(format, a...)
}

func Fatal(err any, message string) {
	debugPrinter(2, err)
	fmt.Fprintln(LogWriter, "FATAL:", message)
	debugPrinter(2, string(debug.Stack()))
	os.Exit(1)
}

func Fatalf(err any, format string, a ...any) {
	message := Paragraph(format, a...)
	Fatal(err, message)
}

func Info(message string) {
	Infof("%s", message)
}

func Infof(format string, a ...any) {
	if logLevel > 0 {
		message := Paragraph(format, a...)
		fmt.Fprintln(LogWriter, "INFO:", message)
	}
}

func Detail(message string) {
	Detailf("%s", message)
}

func Detailf(format string, a ...any) {
	if logLevel > 0 {
		message := fmt.Sprintf(format, a...)
		fmt.Fprintln(LogWriter, "DETAIL:", message)
	}
}

func Success(message string) {
	Successf("%s", message)
}

func Successf(format string, a ...any) {
	message := Paragraph(format, a...)
	fmt.Fprintln(LogWriter, "SUCCESS:", message)
}

func Fail(message string) {
	Failf("%s", message)
}

func Failf(format string, a ...any) {
	message := Paragraph(format, a...)
	fmt.Fprintln(LogWriter, "FAIL:", message)
}

func Question(text string) {
	Questionf("%s", text)
}

func Questionf(format string, a ...any) {
	fmt.Fprintln(LogWriter)
	message := Paragraph(format, a...)
	fmt.Fprintln(LogWriter, "QUESTION:", message)
}

func Note(text string) {
	Notef("%s", text)
}

func Notef(format string, a ...any) {
	fmt.Fprintln(LogWriter)
	message := Paragraphn(TermWidth-7, format, a...)
	fmt.Fprintln(LogWriter, "NOTE:", message)
}

func Title(title string, help string) {
	titleFormatted := fmt.Sprintf(" %s ", title)
	helpFormatted := help
	fmt.Fprintf(LogWriter, "%s  %s\n", titleFormatted, helpFormatted)
}

func HeaderInfof(format string, a ...any) {
	message := Truncate(fmt.Sprintf(format, a...), TermWidth, false)
	padding := TermWidth - len(message)
	fmt.Fprintln(LogWriter)
	fmt.Fprintf(LogWriter, "%s%s\n", message, strings.Repeat(" ", padding))
}

func HorizontalRule() {
	fmt.Fprintln(LogWriter)
	fmt.Fprintln(LogWriter, RuleLine)
}

func JSONValue(value any) string {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		debugPrinter(2, fmt.Sprintf("ERROR marshalling json: %s", err.Error()))
	}
	return string(bytes)
}

func Printf(format string, a ...any) {
	fmt.Fprintf(LogWriter, format, a...)
}

func Paragraph(format string, a ...any) string {
	return Paragraphn(TermWidth, format, a...)
}

func Paragraphn(n int, format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

func PrintDiff(textA, textB string) {
	fmt.Fprintln(LogWriter, "Diff not implemented")
}

func Truncate(text string, length int, invert bool) string {
	textEscaped := strings.ReplaceAll(text, "\n", "; ")
	if len(textEscaped) > length {
		if invert {
			start := len(textEscaped) - length + 3
			textEscaped = "..." + textEscaped[start:]
		} else {
			end := length - 3
			textEscaped = textEscaped[:end] + "..."
		}
	}
	return textEscaped
}

func Table(header []string, data [][]string, columnSize []int) error {
	fmt.Fprintln(LogWriter)
	termWidth := 100 - 10

	if len(columnSize) != len(header) {
		Warn("The number of columns does not match the number of headers")
		columnSize = make([]int, len(header))
		for i := range columnSize {
			columnSize[i] = (len(header) / termWidth) * 100
		}
	}

	for _, row := range data {
		for i, cell := range row {
			row[i] = addLineBreaks(strings.Replace(cell, "\n", " ", -1), (columnSize[i]*termWidth)/100)
		}
		fmt.Fprintln(LogWriter, row)
	}

	return nil
}

func addLineBreaks(input string, maxLineLength int) string {
	words := strings.Fields(input)
	var result strings.Builder
	currentLineLength := 0

	for _, word := range words {
		if currentLineLength+len(word) > maxLineLength {
			firstPart, secondPart := splitHyphenedWords(word, currentLineLength, maxLineLength)
			if firstPart != "" {
				if currentLineLength > 0 {
					result.WriteString(" ")
				}
				result.WriteString(firstPart + "-")
			}
			result.WriteString("\n")
			currentLineLength = 0

			if secondPart != "" {
				word = secondPart
			}
		}
		if currentLineLength > 0 {
			result.WriteString(" ")
			currentLineLength++
		}
		result.WriteString(word)
		currentLineLength += len(word)
	}

	return result.String()
}

func splitHyphenedWords(input string, currentLength int, maxLength int) (firstPart string, secondPart string) {
	hyphenIndicies := []int{}
	for i, char := range input {
		if char == '-' {
			hyphenIndicies = append(hyphenIndicies, i)
		}
	}

	if len(hyphenIndicies) != 0 {
		for i := len(hyphenIndicies) - 1; i >= 0; i-- {
			hyphenIndex := hyphenIndicies[i]
			firstPart = input[:hyphenIndex]
			secondPart = input[hyphenIndex+1:]
			if len(firstPart)+currentLength <= maxLength {
				return firstPart, secondPart
			}
		}
	}
	return "", input
}

func debugPrinter(offset int, a ...any) {
	now := time.Now().Format(time.RFC3339)
	a = append([]any{now, " - "}, a...)
	fmt.Fprintln(LogWriter, a...)

	if useLogFile {
		fmt.Fprintln(logFile, a...)
	}
}
