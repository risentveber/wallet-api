package integration

import (
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func Retry(pause time.Duration, maxCount uint, call func() error, logger log.Logger) error {
	var err error
	for i := uint(0); i < maxCount; i++ {
		err = call()
		if err == nil {
			return nil
		}
		_ = level.Warn(logger).Log(
			"msg", "retry after: "+
				pause.String()+", retries left: "+
				strconv.Itoa(int(maxCount-i)-1)+", error: "+err.Error())
		time.Sleep(pause)
	}

	return err
}
