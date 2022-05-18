package prettylog

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ProgressColor       lipgloss.Color
	SuccessMessageColor lipgloss.Color
	FailedMessageColor  lipgloss.Color
	MaxMessageRows      int
	LogLevels           []logrus.Level
	LogLevelColors      map[logrus.Level]lipgloss.Color
	LogTextColor        lipgloss.Color
	MaxLogRows          int
}

func NewConfig() Config {
	return Config{
		ProgressColor:       lipgloss.Color("178"),
		SuccessMessageColor: lipgloss.Color("82"),
		FailedMessageColor:  lipgloss.Color("160"),
		MaxMessageRows:      10,
		LogLevels: []logrus.Level{
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
		},
		LogTextColor: lipgloss.Color("7"),
		LogLevelColors: map[logrus.Level]lipgloss.Color{
			logrus.InfoLevel:  lipgloss.Color("2"),
			logrus.WarnLevel:  lipgloss.Color("3"),
			logrus.ErrorLevel: lipgloss.Color("1"),
		},
		MaxLogRows: 10,
	}
}
