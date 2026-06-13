package online

import "time"

type ServerOptions struct {
	ReconnectWindow time.Duration
	RoomIdleTTL     time.Duration
}

func (o ServerOptions) withDefaults() ServerOptions {
	if o.ReconnectWindow <= 0 {
		o.ReconnectWindow = 2 * time.Minute
	}
	if o.RoomIdleTTL <= 0 {
		o.RoomIdleTTL = 10 * time.Minute
	}
	return o
}
