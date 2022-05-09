package prettylog

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"prettylog/term"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

func NewPrettyGlobalLogger(cfg Config) PrettyLogger {
	return NewPrettyLogger(logrus.StandardLogger(), cfg)
}

func NewPrettyLogger(logger *logrus.Logger, cfg Config) PrettyLogger {
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return &dummylogger{}
	}

	plog := newprettylogger(cfg)
	logger.AddHook(plog)
	if logger.Out == os.Stdout || logger.Out == os.Stderr {
		logger.SetOutput(ioutil.Discard)
	}

	return plog
}

type PrettyLogger interface {
	AddNewMessage(MessageType, string) error
	UpdateMessage(MessageType, string) error
}

type dummylogger struct {
}

func (*dummylogger) AddNewMessage(MessageType, string) error {
	return nil
}

func (*dummylogger) UpdateMessage(MessageType, string) error {
	return nil
}

type prettylogger struct {
	w            *term.Writer
	cfg          Config
	mu           sync.Mutex
	msgLineCount int
	logLineCount int
}

func newprettylogger(cfg Config) *prettylogger {
	return &prettylogger{
		w:   term.New(os.Stdout),
		cfg: cfg,
	}
}

func (p *prettylogger) Levels() []logrus.Level {
	return p.cfg.Levels
}

func (p *prettylogger) Fire(entry *logrus.Entry) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.logLineCount > p.cfg.MaxLogRows {
		p.clearLogs()
	}

	if p.logLineCount == 0 {
		// add a blank line before
		err := p.writeLog("\n")
		if err != nil {
			return err
		}
	}

	p.setLogColor(entry.Level)
	err := p.writeLog("[%s] %s\n", strings.ToUpper(entry.Level.String()), entry.Message)
	if err != nil {
		return err
	}
	ct.ResetColor()
	return nil
}

func (p *prettylogger) AddNewMessage(tp MessageType, message string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.clearLogs()

	p.w.Reset()
	p.setMessageColor(tp)
	defer ct.ResetColor()
	fmt.Fprintln(p.w, message)

	lineCount, err := p.w.Print()
	if err != nil {
		return err
	}

	p.msgLineCount = lineCount
	return nil
}

func (p *prettylogger) UpdateMessage(tp MessageType, message string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.clearLogs()

	p.w.Clear(-1)

	p.setMessageColor(tp)
	defer ct.ResetColor()
	fmt.Fprintln(p.w, message)

	lineCount, err := p.w.Print()
	if err != nil {
		return err
	}

	p.msgLineCount = lineCount
	return nil
}

func (p *prettylogger) writeLog(format string, args ...interface{}) error {
	fmt.Fprintf(p.w, format, args...)
	lineCount, err := p.w.Print()
	if err != nil {
		return err
	}

	p.logLineCount = lineCount - p.msgLineCount
	return nil
}

func (p *prettylogger) clearLogs() {
	p.w.Clear(p.logLineCount)
	p.logLineCount = 0
}

func (p *prettylogger) setMessageColor(tp MessageType) {
	if col, found := p.cfg.MessageColors[tp]; found {
		ct.Foreground(col, p.cfg.UseBrightColors)
	}
}

func (p *prettylogger) setLogColor(lvl logrus.Level) {
	if col, found := p.cfg.LevelColors[lvl]; found {
		ct.Foreground(col, p.cfg.UseBrightColors)
	}
}
