package main

import "time"

type EventType uint8

func (e EventType) getSafe(pos int) string {
	if int(e) >= len(eventToString) {
		return "Unkown Event"
	}

	return eventToString[e][pos]
}

func (e EventType) String() string {
	return e.getSafe(0)
}

func (e EventType) FullName() string {
	return e.getSafe(1)
}

func (e EventType) ShortName() string {
	return e.getSafe(2)
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
	MonitorCreatedEvent EventType = iota
	MonitorPausedEvent
	MonitorStartedEvent
	MonitorDownEvent
	MonitorUpEvent
)

type Monitor struct {
	Id   int
	Name string
	Type string
	Logs []MonitorLog
}

type MonitorLog struct {
	Id        int
	Event     EventType
	Date      time.Time
	MonitorId int
	Monitor   *Monitor
}

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
