package schedule

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func DoWithSchedule(doInterval, errInterval string, immediately bool, f func() error) {
	d := mustParseInterval(doInterval, "do")
	e := mustParseInterval(errInterval, "err")

	for doIdx := 0; true; doIdx++ {
		if doIdx == 0 && !immediately {
			time.Sleep(d)
		}
		if err := f(); err != nil {
			log.Errorf("schedule func error: %v", err)
			time.Sleep(e)
			continue
		}
		time.Sleep(d)
	}
}

func mustParseInterval(interval, typ string) time.Duration {
	d, err := time.ParseDuration(interval)
	if err != nil {
		log.Fatalf("wrong %s interval format: %s", typ, interval)
	}
	return d
}
