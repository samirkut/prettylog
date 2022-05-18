package prettylog

import tea "github.com/charmbracelet/bubbletea"

// buffer impl to hold last x entries of messages

type buffer struct {
	lines   []tea.Msg
	maxSize int
	size    int
	nextIdx int
}

func newBuffer(maxSize int) *buffer {
	return &buffer{
		lines:   make([]tea.Msg, maxSize, maxSize),
		maxSize: maxSize,
		size:    0,
		nextIdx: 0,
	}
}

func (b *buffer) Len() int {
	return b.size
}

func (b *buffer) Iterate(fn func(tea.Msg)) {
	ptr := b.getFirstPtr()
	for i := 0; i < b.size; i++ {
		fn(b.lines[ptr])
		ptr = (ptr + 1) % b.maxSize
	}
}

func (b *buffer) ReverseIterate(fn func(tea.Msg)) {
	ptr := b.getLastPtr()
	for i := 0; i < b.size; i++ {
		fn(b.lines[ptr])
		ptr--
		if ptr < 0 {
			ptr = b.maxSize - 1
		}
	}
}

func (b *buffer) Add(line tea.Msg) {
	b.lines[b.nextIdx] = line
	b.nextIdx = (b.nextIdx + 1) % b.maxSize
	if b.size < b.maxSize {
		b.size++
	}
}

func (b *buffer) PeekFirst() tea.Msg {
	if b.size == 0 {
		return nil
	}
	return b.lines[b.getFirstPtr()]
}

func (b *buffer) PeekLast() tea.Msg {
	if b.size == 0 {
		return nil
	}
	return b.lines[b.getLastPtr()]
}

func (b *buffer) getLastPtr() int {
	lastPtr := b.nextIdx - 1
	if lastPtr < 0 {
		lastPtr = b.maxSize - 1
	}
	return lastPtr
}

func (b *buffer) getFirstPtr() int {
	if b.size < b.maxSize {
		return 0
	}
	return b.nextIdx
}
