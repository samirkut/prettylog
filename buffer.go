package prettylog

// buffer impl to hold logs. rewrite using a ring buffer

type buffer struct {
	lines []string
}

func newBuffer(size int) *buffer {
	return &buffer{
		lines: make([]string, 0, size),
	}
}

func (b *buffer) Len() int {
	return len(b.lines)
}
func (b *buffer) Lines() []string {
	return b.lines
}

func (b *buffer) PeekFirst() string {
	if b.Len() == 0 {
		return ""
	}
	return b.lines[0]
}

func (b *buffer) PeekLast() string {
	if b.Len() == 0 {
		return ""
	}
	return b.lines[len(b.lines)-1]
}

func (b *buffer) Add(line string) {
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
