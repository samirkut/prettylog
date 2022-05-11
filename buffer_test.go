package prettylog

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuffer(t *testing.T) {
	buffer := newBuffer(5)

	for i := 0; i < 15; i++ {
		line := fmt.Sprintf("line %d", i)
		buffer.Add(line)
		assert.Equal(t, line, buffer.PeekLast())
		if i == 0 {
			assert.Equal(t, line, buffer.PeekFirst())
		}
		if i > 4 {
			assert.Equal(t, 5, buffer.Len())
			assert.Equal(t, fmt.Sprintf("line %d", i-4), buffer.PeekFirst())
		}
	}
}
