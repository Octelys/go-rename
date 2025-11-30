package entities

import "time"

type Audit struct {
	Severity  Severity
	Timestamp time.Time
	Text      string
}

type Severity int

const (
	Debug Severity = iota
	Information
	Warning
	Error
)

func (s Severity) String() string {
	switch s {
	case Debug:
		return "DBG"
	case Information:
		return "INF"
	case Warning:
		return "WAR"
	case Error:
		return "ERR"
	default:
		return "XXX"
	}
}
