package binance

// Logger is an optional structured logger. The default is silent.
// Implementations must never log API secrets, signatures, or PEM material.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

type nopLogger struct{}

func (nopLogger) Debug(string, ...any) {}
func (nopLogger) Info(string, ...any)  {}
func (nopLogger) Error(string, ...any) {}

// StdLogger adapts a printf-style function (for example log.Printf) to Logger.
type StdLogger struct {
	Printf func(format string, args ...any)
}

func (l StdLogger) Debug(msg string, args ...any) { l.log("DEBUG", msg, args...) }
func (l StdLogger) Info(msg string, args ...any)  { l.log("INFO", msg, args...) }
func (l StdLogger) Error(msg string, args ...any) { l.log("ERROR", msg, args...) }

func (l StdLogger) log(level, msg string, args ...any) {
	if l.Printf == nil {
		return
	}
	l.Printf("binance %s: %s %v", level, msg, args)
}
