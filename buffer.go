package prettylog

// buffer impl to hold logs.
// TODO: rewrite using a ring buffer

type buffer struct {
	lines []LogMsg
}

func newBuffer(size int) *buffer {
	return &buffer{
		lines: make([]LogMsg, 0, size),
	}
}

func (b *buffer) Len() int {
	return len(b.lines)
}

func (b *buffer) Lines() []LogMsg {
	return b.lines
}

func (b *buffer) Add(line LogMsg) {
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

func (b *buffer) PeekFirst() LogMsg {
	if b.Len() == 0 {
		return LogMsg{}
	}
	return b.lines[0]
}

func (b *buffer) PeekLast() LogMsg {
	if b.Len() == 0 {
		return LogMsg{}
	}
	return b.lines[b.Len()-1]
}
