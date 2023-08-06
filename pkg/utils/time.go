package utils

import "time"

func NowTimeUTC() time.Time {
  return time.Now().UTC()
}
