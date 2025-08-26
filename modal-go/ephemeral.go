package modal

import (
	"context"
	"time"
)

const ephemeralObjectHeartbeatSleep = 300 * time.Second

func startEphemeralHeartbeat(ctx context.Context, heartbeatFn func() error) {
	go func() {
		t := time.NewTicker(ephemeralObjectHeartbeatSleep)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				_ = heartbeatFn()
			}
		}
	}()
}
