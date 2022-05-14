package prettylog

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

func NewPrettyGlobalLogger(cfg Config) (PrettyLogger, error) {
	return NewPrettyLogger(logrus.StandardLogger(), cfg)
}

func NewPrettyLogger(logger *logrus.Logger, cfg Config) (PrettyLogger, error) {
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return &dummylogger{}, nil
	}

	plog, err := newprettylogger(cfg)
	if err != nil {
		return nil, err
	}

	logger.AddHook(plog)
	if logger.Out == os.Stdout || logger.Out == os.Stderr {
		logger.SetOutput(ioutil.Discard)
		log.SetOutput(ioutil.Discard)
	}

	return plog, nil
}

type PrettyLogger interface {
	Start() error
	Close()
	AddProgress(format string, args ...interface{}) error
	AppendMessage(tp MessageType, format string, args ...interface{}) error
	LogMessage(level logrus.Level, format string, args ...interface{}) error
}

type dummylogger struct {
}

func (*dummylogger) AddProgress(format string, args ...interface{}) error {
	return nil
}

func (*dummylogger) AppendMessage(tp MessageType, format string, args ...interface{}) error {
	return nil
}

func (*dummylogger) LogMessage(lvl logrus.Level, format string, args ...interface{}) error {
	return nil
}

func (*dummylogger) Start() error {
	return nil
}

func (*dummylogger) Close() {}

type prettylogger struct {
	closed bool
	p      *tea.Program
	cfg    Config
	model  model
	mu     sync.RWMutex
}

func newprettylogger(cfg Config) (*prettylogger, error) {
	mod := newModel(cfg)
	p := tea.NewProgram(mod)

	return &prettylogger{
		p:      p,
		cfg:    cfg,
		closed: false,
		model:  mod,
	}, nil
}

func (p *prettylogger) Start() error {
	go func() {
		_ = p.p.Start()

		p.mu.Lock()
		p.closed = true
		close(p.model.logCh)
		close(p.model.progressCh)
		close(p.model.messagesCh)
		p.mu.Unlock()
	}()
	return nil
}

func (p *prettylogger) Close() {
	//HACK: give a second or so to let the channels flush?
	time.Sleep(time.Second)
	p.p.Quit()
	p.p.Kill()
}

func (p *prettylogger) Levels() []logrus.Level {
	return p.cfg.Levels
}

func (p *prettylogger) Fire(entry *logrus.Entry) error {
	return p.writeLog(entry.Level, "[%s] %s", strings.ToUpper(entry.Level.String()), entry.Message)
}

func (p *prettylogger) LogMessage(lvl logrus.Level, format string, args ...interface{}) error {
	return p.writeLog(lvl, format, args...)
}

func (p *prettylogger) AddProgress(format string, args ...interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("closed")
	}

	p.model.progressCh <- ProgressMsg{
		Details: fmt.Sprintf(format, args...),
	}

	return nil
}

func (p *prettylogger) AppendMessage(tp MessageType, format string, args ...interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("closed")
	}

	p.model.messagesCh <- AppendMessage{
		Success: tp == Succeeded,
		Details: fmt.Sprintf(format, args...),
	}

	return nil
}

func (p *prettylogger) writeLog(lvl logrus.Level, format string, args ...interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("closed")
	}

	msg := fmt.Sprintf(format, args...)

	p.model.logCh <- LogMsg{lvl, msg}
	return nil
}
