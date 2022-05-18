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

func NewDummyLogger(cfg Config) PrettyLogger {
	return &dummylogger{}
}

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
		log.SetOutput(ioutil.Discard)
	}

	return plog
}

// PrettyLogger overrides the logger using hooks and presents a clean scrolling UI
// There are also function to show progress or completed messages
type PrettyLogger interface {
	Stop()
	AddProgressMessage(format string, args ...interface{}) error
	AddSuccessMessage(status string, format string, args ...interface{}) error
	AddFailedMessage(status string, format string, args ...interface{}) error
	Log(level logrus.Level, format string, args ...interface{}) error
}

type dummylogger struct {
}

func (*dummylogger) AddProgressMessage(format string, args ...interface{}) error {
	return nil
}

func (*dummylogger) AddSuccessMessage(status, format string, args ...interface{}) error {
	return nil
}

func (*dummylogger) AddFailedMessage(status, format string, args ...interface{}) error {
	return nil
}

func (*dummylogger) Log(lvl logrus.Level, format string, args ...interface{}) error {
	return nil
}

func (*dummylogger) Stop() {}

type prettylogger struct {
	closed bool
	prog   *tea.Program
	cfg    Config
	model  model
	mu     sync.RWMutex // synchronize access to closed bool and for closing channels
}

func newprettylogger(cfg Config) *prettylogger {
	mod := newModel(cfg)
	prog := tea.NewProgram(mod)

	p := &prettylogger{
		prog:   prog,
		cfg:    cfg,
		closed: false,
		model:  mod,
	}

	go func() {
		_ = p.prog.Start()

		p.mu.Lock()
		p.closed = true
		close(p.model.logCh)
		close(p.model.progressCh)
		p.mu.Unlock()
	}()

	return p
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

func (p *prettylogger) Log(lvl logrus.Level, format string, args ...interface{}) error {
	return p.writeLog(lvl, format, args...)
}

func (p *prettylogger) AddProgressMessage(format string, args ...interface{}) error {
	return p.writeProgressMessage(ProgressMsg{
		Details: fmt.Sprintf(format, args...),
	})
}

func (p *prettylogger) AddSuccessMessage(status, format string, args ...interface{}) error {
	return p.writeProgressMessage(ProgressMsg{
		Completed: CompletedMessage{
			Status:  status,
			Success: true,
			Details: fmt.Sprintf(format, args...),
		},
	})
}

func (p *prettylogger) AddFailedMessage(status, format string, args ...interface{}) error {
	return p.writeProgressMessage(ProgressMsg{
		Completed: CompletedMessage{
			Status:  status,
			Success: false,
			Details: fmt.Sprintf(format, args...),
		},
	})
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

func (p *prettylogger) writeProgressMessage(msg ProgressMsg) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("closed channel")
	}

	p.model.progressCh <- msg

	return nil
}
