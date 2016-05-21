package main

import "testing"

func TestEventTypeString(t *testing.T) {
	testcase := []struct {
		Event    EventType
		Expected string
	}{
		{MonitorPausedEvent, "Monitor Paused Event"},
		{MonitorCreatedEvent, "Monitor Created Event"},
		{MonitorDownEvent, "Monitor Down Event"},
		{MonitorStartedEvent, "Monitor Started Event"},
		{MonitorUpEvent, "Monitor Up Event"},
		{EventType(5), "Unkown Event"},
	}

	msg := "EventType(id=%d).String() => %v, wanted %v"
	for _, row := range testcase {
		if row.Event.String() != row.Expected {
			t.Errorf(msg, row.Event, row.Event.String(), row.Expected)
		}
	}
}

func TestEventTypeFullName(t *testing.T) {
	testcase := []struct {
		Event    EventType
		Expected string
	}{
		{MonitorPausedEvent, "Monitor has been paused"},
		{MonitorCreatedEvent, "Monitor has been created"},
		{MonitorDownEvent, "Server is down"},
		{MonitorStartedEvent, "Monitor has been started"},
		{MonitorUpEvent, "Server is up"},
		{EventType(5), "Unkown Event"},
	}

	msg := "EventType(id=%d).Fullname() => %v, wanted %v"
	for _, row := range testcase {
		if row.Event.FullName() != row.Expected {
			t.Errorf(msg, row.Event, row.Event.FullName(), row.Expected)
		}
	}
}

func TestEventTypeShortName(t *testing.T) {
	testcase := []struct {
		Event    EventType
		Expected string
	}{
		{MonitorPausedEvent, "Paused"},
		{MonitorCreatedEvent, "Created"},
		{MonitorDownEvent, "Down"},
		{MonitorStartedEvent, "Started"},
		{MonitorUpEvent, "Up"},
		{EventType(5), "Unkown Event"},
	}

	msg := "EventType(id=%d).ShortName() => %v, wanted %v"
	for _, row := range testcase {
		if row.Event.ShortName() != row.Expected {
			t.Errorf(msg, row.Event, row.Event.ShortName(), row.Expected)
		}
	}
}
