package main

import (
	"time"

	"github.com/samirkut/prettylog"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg := prettylog.NewConfig()
	l := prettylog.NewPrettyGlobalLogger(cfg)
	defer l.Stop()

	runSequence(l, []string{
		"[...] Starting processing again",
		"[...] Still doing it 1",
		"[...] Still doing it 2",
		"[OK] Failed to do the task",
	}, false)
}

func runSequence(l prettylog.PrettyLogger, msgs []string, success bool) {
	for i, m := range msgs {
		if i == len(msgs)-1 {
			l.AddCompletedMessage(success, m)
			time.Sleep(time.Millisecond * 200)
			logrus.Error("random updates after finish")
		} else {
			l.AddProgress(m)

			time.Sleep(time.Millisecond * 200)
			logrus.Warn("initializing messages something like this will have to do for now. lorem epsum ditum?")

			time.Sleep(time.Second * 1)

			logrus.Infof("%d more random updates...", i)
			time.Sleep(time.Millisecond * 200)
			logrus.Infof("%d more random updates...", i)
			time.Sleep(time.Millisecond * 200)
			logrus.Infof("%d more random updates...", i)
			time.Sleep(time.Millisecond * 200)
		}
	}
}
