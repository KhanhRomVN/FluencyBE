package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelSuccess
	LevelWarning
	LevelError
	LevelCritical
)

type PrettyLogger struct {
	level        LogLevel
	output       io.Writer
	service      string
	colorful     bool
	mu           sync.Mutex
	activeLevels map[LogLevel]bool
}

type DiscordWebhook struct {
	Content string         `json:"content"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordEmbed struct {
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	Color       int            `json:"color,omitempty"`
	Fields      []DiscordField `json:"fields,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty"`
	Footer      *DiscordFooter `json:"footer,omitempty"`
	Author      *DiscordAuthor `json:"author,omitempty"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

type DiscordAuthor struct {
	Name    string `json:"name"`
	IconURL string `json:"icon_url,omitempty"`
}

type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	EventCode string
	Fields    map[string]interface{}
	Message   string
}

func NewPrettyLogger(service string, level LogLevel, colorful bool) *PrettyLogger {
	return &PrettyLogger{
		level:        level,
		output:       os.Stdout,
		service:      service,
		colorful:     colorful,
		activeLevels: make(map[LogLevel]bool),
	}
}

func (l *PrettyLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *PrettyLogger) checkLogLevel(level LogLevel) bool {
	logLevelsStr := os.Getenv("LOG_LEVELS")
	if logLevelsStr != "" {
		currentLevel := strings.ToUpper(levelToString(level))
		levels := strings.Split(strings.ToUpper(logLevelsStr), ",")
		for _, l := range levels {
			if strings.TrimSpace(l) == currentLevel {
				return true
			}
		}
		return false
	}
	return true
}

func (l *PrettyLogger) Debug(eventCode string, fields map[string]interface{}, message string) {
	if !l.checkLogLevel(LevelDebug) {
		return
	}
	l.log(LevelDebug, eventCode, fields, message)
}

func (l *PrettyLogger) Info(eventCode string, fields map[string]interface{}, message string) {
	if !l.checkLogLevel(LevelInfo) {
		return
	}
	l.log(LevelInfo, eventCode, fields, message)
}

func (l *PrettyLogger) Success(eventCode string, fields map[string]interface{}, message string) {
	if !l.checkLogLevel(LevelSuccess) {
		return
	}
	l.log(LevelSuccess, eventCode, fields, message)
}

func (l *PrettyLogger) Warning(eventCode string, fields map[string]interface{}, message string) {
	if !l.checkLogLevel(LevelWarning) {
		return
	}
	l.log(LevelWarning, eventCode, fields, message)
}

func (l *PrettyLogger) Error(eventCode string, fields map[string]interface{}, message string) {
	if !l.checkLogLevel(LevelError) {
		return
	}
	l.log(LevelError, eventCode, fields, message)
}

func (l *PrettyLogger) Critical(eventCode string, fields map[string]interface{}, message string) {
	if !l.checkLogLevel(LevelCritical) {
		return
	}
	l.log(LevelCritical, eventCode, fields, message, false)
}

func (l *PrettyLogger) log(level LogLevel, eventCode string, fields map[string]interface{}, message string, systemWideAlert ...bool) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		EventCode: eventCode,
		Fields:    fields,
		Message:   message,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	env := os.Getenv("ENV")
	if env == "production" {
		l.logToDiscord(entry, len(systemWideAlert) > 0 && systemWideAlert[0])
	} else {
		_, file, line, _ := runtime.Caller(2)
		caller := fmt.Sprintf("%s:%d", shortenFilePath(file), line)
		if l.colorful {
			l.printColorful(entry, caller)
		} else {
			l.printPlain(entry, caller)
		}
	}
}

func (l *PrettyLogger) logToDiscord(entry LogEntry, systemWideAlert bool) {
	webhookURL := os.Getenv("DISCORD_ACCOUNT_WEBHOOK_URL")
	if webhookURL == "" {
		return
	}

	levelColor := 0x7289DA
	levelEmoji := "i️"
	mentionContent := ""

	switch entry.Level {
	case LevelCritical:
		levelColor = 0xFF0000
		levelEmoji = "🔴"
		mentionContent = "<@&1335118454113566720>"
	case LevelError:
		levelColor = 0xFF0000
		levelEmoji = "🔴"
	case LevelWarning:
		levelColor = 0xFFA500
		levelEmoji = "⚠️"
	case LevelInfo:
		levelColor = 0x00FF00
	case LevelSuccess:
		levelColor = 0x00FF00
		levelEmoji = "✅"
	case LevelDebug:
		levelColor = 0x7289DA
		levelEmoji = "🔍"
	}

	var fieldsList []DiscordField
	if entry.EventCode != "" {
		fieldsList = append(fieldsList, DiscordField{
			Name:   "Event Code",
			Value:  fmt.Sprintf("`%s`", entry.EventCode),
			Inline: true,
		})
	}

	fieldsList = append(fieldsList, DiscordField{
		Name:   "Timestamp",
		Value:  fmt.Sprintf("<t:%d:F>", entry.Timestamp.Unix()),
		Inline: true,
	})

	for k, v := range entry.Fields {
		value := fmt.Sprintf("%v", v)
		if strings.ContainsAny(value, "{}[]()") || strings.Contains(k, "id") {
			value = fmt.Sprintf("`%s`", value)
		}
		fieldsList = append(fieldsList, DiscordField{
			Name:   strings.Title(k),
			Value:  value,
			Inline: true,
		})
	}

	webhook := DiscordWebhook{
		Content: mentionContent,
		Embeds: []DiscordEmbed{
			{
				Author: &DiscordAuthor{
					Name: fmt.Sprintf("%s %s", levelEmoji, strings.ToUpper(levelToString(entry.Level))),
				},
				Title:  entry.Message,
				Color:  levelColor,
				Fields: fieldsList,
				Footer: &DiscordFooter{
					Text: fmt.Sprintf("Service: %s", l.service),
				},
				Timestamp: entry.Timestamp.Format(time.RFC3339),
			},
		},
	}

	payload, err := json.Marshal(webhook)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling Discord webhook: %v\n", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending to Discord: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

func (l *PrettyLogger) printColorful(entry LogEntry, caller string) {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
	levelStr := strings.ToUpper(levelToString(entry.Level))

	timestampColor := color.New(color.FgHiBlue).Add(color.Faint)
	serviceColor := color.New(color.FgHiMagenta)
	callerColor := color.New(color.FgHiBlue).Add(color.Faint)
	messageColor := color.New(color.FgWhite)
	fieldsColor := color.New(color.FgHiBlack)
	eventCodeColor := color.New(color.FgCyan)

	var levelColor *color.Color
	switch entry.Level {
	case LevelDebug:
		levelColor = color.New(color.FgHiCyan)
	case LevelInfo:
		levelColor = color.New(color.FgHiGreen)
	case LevelSuccess:
		levelColor = color.New(color.FgHiGreen)
	case LevelWarning:
		levelColor = color.New(color.FgHiYellow)
	case LevelError:
		levelColor = color.New(color.FgHiRed)
	case LevelCritical:
		levelColor = color.New(color.BgRed, color.FgHiWhite)
	}

	coloredTimestamp := timestampColor.Sprint(timestamp)
	coloredLevel := levelColor.Sprint(levelStr)
	coloredService := serviceColor.Sprintf("[%s]", l.service)
	coloredCaller := callerColor.Sprint(caller)
	coloredMessage := messageColor.Sprint(entry.Message)

	logLine := fmt.Sprintf("%s [%s] %s %s - %s",
		coloredTimestamp,
		coloredLevel,
		coloredService,
		coloredCaller,
		coloredMessage,
	)

	if len(entry.Fields) > 0 {
		fields := make([]string, 0, len(entry.Fields))
		for k, v := range entry.Fields {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		logLine += fieldsColor.Sprintf(" | %s", strings.Join(fields, " "))
	}

	if entry.EventCode != "" {
		logLine = eventCodeColor.Sprintf("[%s] ", entry.EventCode) + logLine
	}

	fmt.Fprintln(l.output, logLine)
}

func (l *PrettyLogger) printPlain(entry LogEntry, caller string) {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
	levelStr := strings.ToUpper(levelToString(entry.Level))

	logLine := fmt.Sprintf("%s [%s] [%s] %s - %s",
		timestamp,
		levelStr,
		l.service,
		caller,
		entry.Message,
	)

	if len(entry.Fields) > 0 {
		fields := make([]string, 0, len(entry.Fields))
		for k, v := range entry.Fields {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		logLine += fmt.Sprintf(" | %s", strings.Join(fields, " "))
	}

	if entry.EventCode != "" {
		logLine = fmt.Sprintf("[%s] %s", entry.EventCode, logLine)
	}

	fmt.Fprintln(l.output, logLine)
}

func levelToString(level LogLevel) string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelSuccess:
		return "SUCCESS"
	case LevelWarning:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

func shortenFilePath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], "/")
	}
	return path
}
