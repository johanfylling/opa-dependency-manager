package printer

import (
	"fmt"
	"io"
	"os"
)

const (
	OutputLevel = iota
	InfoLevel
	DebugLevel
	TraceLevel
)

var (
	LogLevel              = OutputLevel
	PrintWriter io.Writer = os.Stdout
	LogWriter   io.Writer = os.Stderr
	noOpWriter  io.Writer = nil
)

func out(writer io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(writer, format+"\n", args...)
}

func Output(format string, args ...any) {
	out(PrintWriter, format, args...)
}

func Info(format string, args ...any) {
	if LogLevel >= InfoLevel {
		out(LogWriter, format, args...)
	}
}

func InfoPrinter() io.Writer {
	if LogLevel >= InfoLevel {
		return LogWriter
	} else {
		return noOpWriter
	}
}

func Debug(format string, args ...any) {
	if LogLevel >= DebugLevel {
		out(LogWriter, format, args...)
	}
}

func DebugPrinter() io.Writer {
	if LogLevel >= DebugLevel {
		return LogWriter
	} else {
		return noOpWriter
	}
}

func Trace(format string, args ...any) {
	if LogLevel >= TraceLevel {
		out(LogWriter, format, args...)
	}
}

func TracePrinter() io.Writer {
	if LogLevel >= TraceLevel {
		return LogWriter
	} else {
		return noOpWriter
	}
}
