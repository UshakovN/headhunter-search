package utils

import (
	"fmt"
	"time"
)

func NowTimeUTC() time.Time {
	return time.Now().UTC()
}

func TimeStrCast(timeStr string, srcLayout, dscLayout string) (string, error) {
	t, err := time.Parse(srcLayout, timeStr)
	if err != nil {
		return "", fmt.Errorf("cannot parse time string: %v", err)
	}
	timeStr = t.Format(dscLayout)
	return timeStr, nil
}
