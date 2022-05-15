package prettylog

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

// PrettyLogger overrides the logger using hooks and presents a clean scrolling UI
// There are also function to show progress or completed messages
type PrettyLogger interface {
	Start()
	Stop()
	AddProgress(format string, args ...interface{}) error
	AddCompletedMessage(success bool, format string, args ...interface{}) error
	LogMessage(level logrus.Level, format string, args ...interface{}) error
}

type dummylogger struct {
}

func (*dummylogger) AddProgress(format string, args ...interface{}) error {
	return nil
}

func (*dummylogger) AddCompletedMessage(success bool, format string, args ...interface{}) error {
	return nil
}

func (*dummylogger) LogMessage(lvl logrus.Level, format string, args ...interface{}) error {
	return nil
}

func (*dummylogger) Start() {
}

func (*dummylogger) Stop() {}

type prettylogger struct {
	closed bool
	prog   *tea.Program
	cfg    Config
	model  model
	mu     sync.RWMutex // synchronize access to closed bool and for closing channels
}

func newprettylogger(cfg Config) (*prettylogger, error) {
	mod := newModel(cfg)
	prog := tea.NewProgram(mod)

	return &prettylogger{
		prog:   prog,
		cfg:    cfg,
		closed: false,
		model:  mod,
	}, nil
}

func (p *prettylogger) Start() {
	go func() {
		_ = p.prog.Start()

		p.mu.Lock()
		p.closed = true
		close(p.model.logCh)
		close(p.model.progressCh)
		p.mu.Unlock()
	}()
}

func (p *prettylogger) Stop() {
	//HACK: give a second or so to let the channels flush?
	time.Sleep(time.Second)
	p.prog.Quit()
	p.prog.Kill()
}

func (p *prettylogger) Levels() []logrus.Level {
	return p.cfg.LogLevels
}

func (p *prettylogger) Fire(entry *logrus.Entry) error {
	return p.writeLog(entry.Level, entry.Message)
}

func (p *prettylogger) LogMessage(lvl logrus.Level, format string, args ...interface{}) error {
	return p.writeLog(lvl, format, args...)
}

func (p *prettylogger) AddProgress(format string, args ...interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("closed channel")
	}

	p.model.progressCh <- ProgressMsg{
		Details: fmt.Sprintf(format, args...),
	}

	return nil
}

func (p *prettylogger) AddCompletedMessage(success bool, format string, args ...interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("closed channel")
	}

	p.model.progressCh <- ProgressMsg{
		Completed: CompletedMessage{
			Success: success,
			Details: fmt.Sprintf(format, args...),
		},
	}

	return nil
}

func (p *prettylogger) writeLog(lvl logrus.Level, format string, args ...interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("closed channel")
	}

	msg := fmt.Sprintf(format, args...)

	p.model.logCh <- LogMsg{lvl, msg}
	return nil
}
