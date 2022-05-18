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

	ch := make(chan int)

	go printLogs(ch)

	runSequence(l, []string{
		"[...] A Starting processing",
		"[...] A Still doing it 1",
		"[...] A Still doing it 2",
		"[...] A Still doing it 3",
		"[...] A Still doing it 4",
		"[...] A Still doing it 5",
		"[OK] A Finished",
	}, true)

	runSequence(l, []string{
		"[...] B Starting processing again",
		"[...] B Still doing it 1",
		"[...] B Still doing it 2",
		"[...] B Still doing it 3",
		"[...] B Still doing it 4",
		"[...] B Still doing it 5",
		"[FAILED] B Failed to do the task",
	}, false)

	ch <- 0
	close(ch)
}

func runSequence(l prettylog.PrettyLogger, msgs []string, success bool) {
	for i, m := range msgs {
		time.Sleep(time.Millisecond * 300)
		if i == 0 {
			l.AddProgress(m)
			logrus.Warn("initializing messages something like this will have to do for now. lorem epsum ditum?")
		} else {

			if i == len(msgs)-1 {
				l.AddCompletedMessage(success, m)
			} else {
				l.AddProgress(m)
			}

			logrus.Info("random updates coming in from the activity")
		}
		for i := 0; i < 10; i++ {
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
