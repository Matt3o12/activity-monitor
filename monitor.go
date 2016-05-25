package main

import "time"

// EventType contains information about the server's
// state (like monitioring has been stated, or that
// the server is up.
// This status is saved in the Database table `monitor_logs`
// as `Event`.
type EventType uint8

func (e EventType) getSafe(pos int) string {
	if int(e) >= len(eventToString) {
		return "Unkown Event"
	}

	return eventToString[e][pos]
}

// String returns the EventName like this:
// Monitor Created Event.
func (e EventType) String() string {
	return e.getSafe(0)
}

// FullName returns the full description of the event
// such as "Montior has been created.
func (e EventType) FullName() string {
	return e.getSafe(1)
}

// ShortName returns the short description of the event
// (such as Created or Started).
func (e EventType) ShortName() string {
	return e.getSafe(2)
}

func (e EventType) CSSColor() string {
	switch e {
	case MonitorDownEvent:
		return "red"

	case MonitorUpEvent:
		return "green"

	default:
		return "#FFC107"
	}
}

var eventToString = [][]string{
	{
		"Monitor Created Event",
		"Monitor has been created",
		"Created",
	}, {
		"Monitor Paused Event",
		"Monitor has been paused",
		"Paused",
	}, {
		"Monitor Started Event",
		"Monitor has been started",
		"Started",
	}, {
		"Monitor Down Event",
		"Server is down",
		"Down",
	}, {
		"Monitor Up Event",
		"Server is up",
		"Up",
	},
}

const (
	// MonitorCreatedEvent indicates that the monitor
	// has been created.
	MonitorCreatedEvent EventType = iota

	// MonitorPausedEvent indicated that the monitor
	// has been paused.
	MonitorPausedEvent

	// MonitorStartedEvent indicated that monitroing
	// has been started.
	MonitorStartedEvent

	// MonitorDownEvent indicates that the server
	// being monitored is down.
	MonitorDownEvent

	// MonitorUpEvent indicates that the server is up.
	MonitorUpEvent

	// MonitorMax is the monitor with the highest event number
	// (currently MonitorUpEvent)
	MontiorMax = MonitorUpEvent
)

// A Montior holds basic information about the server
// being monitored such as the type, name and a reference
// to all logs.
type Monitor struct {
	Id   int
	Name string
	Type string
	Logs []MonitorLog
}

// MonitorLog is a log entery for any EventType
// that has occurred.
type MonitorLog struct {
	Id        int
	Event     EventType
	Date      time.Time
	MonitorId int
	Monitor   *Monitor
}

// SUpportedTypes is a slice with all monitoring
// types such as pinging or checking for a word
// in an http page.
var SupportedTypes = []string{
	"Socket", "http", "ping",
}

func makeMonitors() []Monitor {
	return []Monitor{
		{0, "TCP/UDP Socket", "sock", []MonitorLog{}},
		{1, "HTTP(s) Server", "http", []MonitorLog{}},
		{2, "Main Server", "ping", []MonitorLog{}},
		{3, "Down server", "ping", []MonitorLog{}},
	}
}
