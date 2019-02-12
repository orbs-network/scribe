package log

import "regexp"

const TEST_FAILED_ERROR = "Test failed due to unexpected errors being logged. If the error above is expected, please add it to the list of allowed errors by invoking TestOutput.AllowErrorsMatching"

type TLog interface {
	Fatal(args ...interface{})
	Log(args ...interface{})
	Error(args ...interface{})
}

func NewTestOutput(tb TLog, formatter LogFormatter) *TestOutput {
	return &TestOutput{tb: tb, formatter: formatter}
}

type TestOutput struct {
	formatter     LogFormatter
	tb            TLog
	stopLogging   bool
	allowedErrors []string
	hasErrors     bool
}

func (o *TestOutput) allowed(message string, fields []*Field) bool {
	for _, allowedMessage := range o.allowedErrors {
		if matched, _ := regexp.MatchString(allowedMessage, message); matched {
			return true
		}
		for _, f := range fields {
			if f.Key == "error" {
				if matched, _ := regexp.MatchString(allowedMessage, f.String()); matched {
					return true
				}
			}
		}
	}

	return false
}

func (o *TestOutput) AllowErrorsMatching(pattern string) {
	o.allowedErrors = append(o.allowedErrors, pattern)
}

func (o *TestOutput) HasErrors() bool {
	return o.hasErrors
}

func (o *TestOutput) recordError(line string) {
	o.hasErrors = true
	o.tb.Error(line)
	o.tb.Error(TEST_FAILED_ERROR)
}

// func (o *TestOutput) Append(level string, message string, fields ...*Field) moved to file t.go
