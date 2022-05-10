package main

import (
	"github.com/samirkut/prettylog"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg := prettylog.NewConfig()
	l := prettylog.NewPrettyGlobalLogger(cfg)

	runSequence(l, []string{
		"[...] Starting processing again",
		"[...] Still doing it 1",
		"[...] Still doing it 2",
		"[OK] Failed to do the task",
	}, false)

}

func runSequence(l prettylog.PrettyLogger, msgs []string, success bool) {
	for i, m := range msgs {
		if i == 0 {
			l.AddNewMessage(prettylog.InProgress, m)
			time.Sleep(time.Millisecond * 200)
			logrus.Warn("initializing messages something like this will have to do for now. lorem epsum ditum?")

			time.Sleep(time.Second * 1)

			logrus.Infof("%d more random updates...", i)
			time.Sleep(time.Millisecond * 200)
			logrus.Infof("%d more random updates...", i)
			time.Sleep(time.Millisecond * 200)
			logrus.Infof("%d more random updates...", i)
			time.Sleep(time.Millisecond * 200)
		} else {
			tp := prettylog.InProgress
			if i == len(msgs)-1 {
				tp = prettylog.Failed
				if success {
					tp = prettylog.Succeeded
				}

				time.Sleep(time.Millisecond * 200)
				logrus.Error("random updates after finish")
			}

			l.UpdateMessage(tp, m)
		}
	}
}
