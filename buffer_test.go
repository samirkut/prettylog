package prettylog

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestBuffer(t *testing.T) {
	size := 10
	buffer := newBuffer(size)

	for i := 0; i < 100; i++ {
		line := fmt.Sprintf("line %d", i)
		buffer.Add(line)

		assert.Equal(t, line, buffer.PeekLast().(string))

		lst := []string{}
		buffer.Iterate(func(res tea.Msg) {
			lst = append(lst, res.(string))
		})

		if i < size-1 {
			assert.Equal(t, i+1, buffer.Len())
			assert.Equal(t, "line 0", buffer.PeekFirst().(string))

			for j := 0; j < buffer.Len(); j++ {
				assert.Equal(t, fmt.Sprintf("line %d", j), lst[j])
			}
		} else {
			assert.Equal(t, size, buffer.Len())
			assert.Equal(t, fmt.Sprintf("line %d", i-(size-1)), buffer.PeekFirst().(string))

			for j := 0; j < buffer.Len(); j++ {
				assert.Equal(t, fmt.Sprintf("line %d", i-(size-1)+j), lst[j])
			}
		}
	}
}
