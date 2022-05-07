package main

import (
	"prettylog"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg := prettylog.NewConfig()
	l := prettylog.NewPrettyGlobalLogger(cfg)
	ch := make(chan int)

	go printLogs(ch)

	runSequence(l, []string{
		"[...] Starting processing",
		"[...] Still doing it 1",
		"[...] Still doing it 2",
		"[...] Still doing it 3",
		"[...] Still doing it 4",
		"[...] Still doing it 5",
		"[OK] Finished",
	}, true)

	runSequence(l, []string{
		"[...] Starting processing again",
		"[...] Still doing it 1",
		"[...] Still doing it 2",
		"[...] Still doing it 3",
		"[...] Still doing it 4",
		"[...] Still doing it 5",
		"[OK] Failed to do the task",
	}, false)

	ch <- 0
	close(ch)

	time.Sleep(time.Second)
}

func runSequence(l prettylog.PrettyLogger, msgs []string, success bool) {
	for i, m := range msgs {
		time.Sleep(time.Second * 3)
		if i == 0 {
			l.AddNewMessage(prettylog.InProgress, m)
			logrus.Warn("initializing messages something like this will have to do for now. lorem epsum ditum?")
		} else {
			tp := prettylog.InProgress
			if i == len(msgs)-1 {
				tp = prettylog.Failed
				if success {
					tp = prettylog.Succeeded
				}
			}
			l.UpdateMessage(tp, m)
			logrus.Info("random updates coming in from the activity")
		}
		for i := 0; i < 25; i++ {
			logrus.Infof("%d more random updates baa baa black sheep, have you any wool?", i)
			time.Sleep(time.Millisecond * 200)
		}
	}
}

func printLogs(quit <-chan int) {
	for {
		select {
		case <-quit:
			return
		case <-time.After(time.Millisecond * 1000):
			logrus.Info("something happened at ", time.Now().String())
		}
	}
}
