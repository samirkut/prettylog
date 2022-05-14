package prettylog

import (
	ct "github.com/daviddengcn/go-colortext"
	"github.com/sirupsen/logrus"
)

type MessageType int

const (
	Succeeded MessageType = iota
	Failed
)

type Config struct {
	ProgressColor   ct.Color
	MessageColors   map[MessageType]ct.Color
	UseBrightColors bool
	Levels          []logrus.Level
	LevelColors     map[logrus.Level]ct.Color
	MaxLogRows      int
}

func NewConfig() Config {
	return Config{
		ProgressColor: ct.Yellow,
		MessageColors: map[MessageType]ct.Color{
			Succeeded: ct.Green,
			Failed:    ct.Red,
		},
		UseBrightColors: true,
		Levels: []logrus.Level{
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
		},
		LevelColors: map[logrus.Level]ct.Color{
			//logrus.InfoLevel:  ct.Green,
			logrus.WarnLevel:  ct.Yellow,
			logrus.ErrorLevel: ct.Red,
		},
		MaxLogRows: 10,
	}
}
