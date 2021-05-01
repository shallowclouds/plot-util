package util

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func LogTimeCost(name ...string) func() {
	start := time.Now()
	return func() {
		logrus.Debugf("%s costs time: %.2fs",
			strings.Join(name, " "),
			float64(time.Since(start).Milliseconds()) / 1000,
		)
	}
}
