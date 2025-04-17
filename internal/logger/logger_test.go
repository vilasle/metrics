package logger

import (
	"testing"
)

type MockWriterSync struct {
	content [][]byte
}

func (w *MockWriterSync) Write(p []byte) (n int, err error) {
	w.content = append(w.content, p)
	return len(p), nil
}

func (w MockWriterSync) Sync() error {
	return nil
}

func (w MockWriterSync) last() []byte {
	if len(w.content) == 0 {
		return nil
	}
	return w.content[len(w.content)-1]
}

func TestLogger(t *testing.T) {
	// w := &MockWriterSync{}

	// Init(w, true)

	// checks := []struct {
	// 	message string
	// 	*regexp.Regexp
	// 	write func(msg string)
	// }{
	// 	{
	// 		message: "debug message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"debug","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"debug message".+}\n$`),
	// 		write: func(msg string) { Debug(msg) },
	// 	},
	// 	{
	// 		message: "info message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"info","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"info message".+}\n$`),
	// 		write: func(msg string) { Info(msg) },
	// 	},
	// 	{
	// 		message: "warn message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"warn","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"warn message".+}\n$`),
	// 		write: func(msg string) { Warn(msg) },
	// 	},
	// 	{
	// 		message: "error message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"error","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"error message".+}\n$`),
	// 		write: func(msg string) { Error(msg) },
	// 	},

	// 	{
	// 		message: "debug message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"debug","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"debug message".+}\n$`),
	// 		write: func(msg string) { Debugf(msg) },
	// 	},
	// 	{
	// 		message: "info message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"info","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"info message".+}\n$`),
	// 		write: func(msg string) { Infof(msg) },
	// 	},
	// 	{
	// 		message: "warn message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"warn","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"warn message".+}\n$`),
	// 		write: func(msg string) { Warnf(msg) },
	// 	},
	// 	{
	// 		message: "error message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"error","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"error message".+}\n$`),
	// 		write: func(msg string) { Errorf(msg) },
	// 	},

	// 	{
	// 		message: "debug message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"debug","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"debug message".+}\n$`),
	// 		write: func(msg string) { Debugw(msg) },
	// 	},
	// 	{
	// 		message: "info message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"info","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"info message".+}\n$`),
	// 		write: func(msg string) { Infow(msg) },
	// 	},
	// 	{
	// 		message: "warn message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"warn","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"warn message".+}\n$`),
	// 		write: func(msg string) { Warnw(msg) },
	// 	},
	// 	{
	// 		message: "error message",
	// 		Regexp: regexp.MustCompile(
	// 			`^{"level":"error","ts":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}\+\d{4}","msg":"error message".+}\n$`),
	// 		write: func(msg string) { Errorw(msg) },
	// 	},
	// }

	// for _, check := range checks {
	// 	check.write(check.message)
	// 	last := w.last()

	// 	if !check.Match(last) {
	// 		t.Errorf("expected %s, got %s", check.message, string(last))
	// 	}
	// }
}
