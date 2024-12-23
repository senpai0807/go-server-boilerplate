package helpers

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var secretKeyOnce sync.Once
var secretKey string

func GetSecretKey() string {
	secretKeyOnce.Do(func() {
		secretKey = os.Getenv("SECRET_KEY")
		if secretKey == "" {
			panic("SECRET_KEY environment variable not set")
		}
	})
	return secretKey
}

func FormatDate(t time.Time) string {
	return t.Format("15:04:05 01/02/2006")
}

type ColorizedLogger struct {
	useColor bool
}

var colorCodes = map[string]string{
	"info":    "\033[34m",
	"verbose": "\033[36m",
	"warn":    "\033[33m",
	"http":    "\033[35m",
	"silly":   "\033[32m",
	"error":   "\033[31m",
	"reset":   "\033[0m",
}

func (l *ColorizedLogger) log(level, message string) {
	timestamp := FormatDate(time.Now())
	levelUpper := strings.ToUpper(level)
	color := colorCodes[level]
	reset := colorCodes["reset"]

	if !l.useColor {
		color = ""
		reset = ""
	}

	logMessage := fmt.Sprintf("%s[%s]: [%s] | %s%s\n", color, timestamp, levelUpper, message, reset)
	os.Stdout.WriteString(logMessage)
}

func NewColorizedLogger(useColor bool) *ColorizedLogger {
	return &ColorizedLogger{useColor: useColor}
}

func (l *ColorizedLogger) Info(message string) {
	l.log("info", message)
}

func (l *ColorizedLogger) Verbose(message string) {
	l.log("verbose", message)
}

func (l *ColorizedLogger) Warn(message string) {
	l.log("warn", message)
}

func (l *ColorizedLogger) Http(message string) {
	l.log("http", message)
}

func (l *ColorizedLogger) Silly(message string) {
	l.log("silly", message)
}

func (l *ColorizedLogger) Error(message string) {
	l.log("error", message)
}
