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
	AddNewMessage(MessageType, string)
	UpdateMessage(MessageType, string)
}

type dummylogger struct {
}

func (*dummylogger) AddNewMessage(MessageType, string) {
}

func (*dummylogger) UpdateMessage(MessageType, string) {
}

type prettylogger struct {
	w    *term.Writer
	logW *term.Writer
	cfg  Config
	mu   sync.Mutex
}

func newprettylogger(cfg Config) *prettylogger {
	return &prettylogger{
		w:    term.New(os.Stdout),
		logW: term.New(os.Stdout),
		cfg:  cfg,
	}
}

func (p *prettylogger) Levels() []logrus.Level {
	return p.cfg.Levels
}

func (p *prettylogger) Fire(entry *logrus.Entry) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// hooks are protected by a mutex
	p.setLogColor(entry.Level)
	_, err := fmt.Fprintf(p.w, "[%s] ", strings.ToUpper(entry.Level.String()))
	if err != nil {
		return err
	}

	err = p.logW.Print()
	if err != nil {
		return err
	}

	ct.ResetColor()
	_, err = fmt.Fprintln(p.w, entry.Message)
	if err != nil {
		return err
	}

	return p.logW.Print()
}

func (p *prettylogger) AddNewMessage(tp MessageType, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.logW.Reset()
	p.w.Reset()
	p.setMessageColor(tp)
	fmt.Fprintln(p.w, message+"\n")
	p.w.Print()
	ct.ResetColor()
}

func (p *prettylogger) UpdateMessage(tp MessageType, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.logW.Reset()
	p.w.Clear()
	p.setMessageColor(tp)
	fmt.Fprintln(p.w, message+"\n")
	p.w.Print()
	ct.ResetColor()
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
