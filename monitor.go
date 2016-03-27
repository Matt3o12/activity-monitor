package main

type Monitor struct {
	Name   string
	Type   string
	Uptime float32
}

func makeMonitors() []Monitor {
	return []Monitor{
		{"TCP/UDP Socket", "sock", 1},
		{"HTTP(s) Server", "http", 50},
		{"Main Server", "ping", 100},
	}
}
