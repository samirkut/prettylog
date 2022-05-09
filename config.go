package prettylog

import (
	ct "github.com/daviddengcn/go-colortext"
	"github.com/sirupsen/logrus"
)

type MessageType int

const (
	NotStarted MessageType = iota
	InProgress
	Succeeded
	Failed
)

type Config struct {
	MessageColors   map[MessageType]ct.Color
	UseBrightColors bool
	Levels          []logrus.Level
	LevelColors     map[logrus.Level]ct.Color
	MaxLogRows      int
}

func NewConfig() Config {
	return Config{
		MessageColors: map[MessageType]ct.Color{
			InProgress: ct.Yellow,
			Succeeded:  ct.Green,
			Failed:     ct.Red,
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
		MaxLogRows: 5,
	}
}
