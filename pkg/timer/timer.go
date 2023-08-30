package timer

import "time"

type RefreshTimer interface {
	Stop() bool
	Wait() <-chan time.Time
}

type wrappedTimer struct {
	d     time.Duration
	timer *time.Timer
}

func NewRefreshTimer(d time.Duration, immediately bool) RefreshTimer {
	var dur time.Duration
	if !immediately {
		dur = d
	}
	return &wrappedTimer{
		d:     d,
		timer: time.NewTimer(dur),
	}
}

func (t *wrappedTimer) Stop() bool {
	return t.timer.Stop()
}

func (t *wrappedTimer) Wait() <-chan time.Time {
	defer t.timer.Reset(t.d)
	return t.timer.C
}
