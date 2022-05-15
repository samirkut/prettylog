package prettylog

import tea "github.com/charmbracelet/bubbletea"

// buffer impl to hold last x entries of messages
// TODO: rewrite using a ring buffer

type buffer struct {
	lines []tea.Msg
}

func newBuffer(size int) *buffer {
	return &buffer{
		lines: make([]tea.Msg, 0, size),
	}
}

func (b *buffer) Len() int {
	return len(b.lines)
}

func (b *buffer) Lines() []tea.Msg {
	return b.lines
}

func (b *buffer) Add(line tea.Msg) {
	if len(b.lines) < cap(b.lines) {
		b.lines = append(b.lines, line)
	} else {
		// this is quite inefficient but avoids memory leaks due to slicing approach
		for i := 1; i < len(b.lines); i++ {
			b.lines[i-1] = b.lines[i]
		}
		b.lines[len(b.lines)-1] = line
	}
}

func (b *buffer) PeekFirst() tea.Msg {
	if b.Len() == 0 {
		return nil
	}
	return b.lines[0]
}

func (b *buffer) PeekLast() tea.Msg {
	if b.Len() == 0 {
		return nil
	}
	return b.lines[b.Len()-1]
}
