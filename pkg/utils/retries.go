package utils

import (
  "errors"
  "time"

  log "github.com/sirupsen/logrus"
)

var ErrDoRetry = errors.New("do retry")

func DoWithRetries(count int, wait time.Duration, f func() error) error {
  var err error

  for tryIdx := 0; tryIdx < count; tryIdx++ {
    if err = f(); err != nil {
      if errors.Is(err, ErrDoRetry) {
        log.Warnf("try: %d. error: %v", tryIdx, err)
        time.Sleep(wait)
        continue
      }
      return err
    }
    break
  }
  return err
}
