//go:build linux
// +build linux

package term

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// accepts the number of lines to clear. -1 implies all
// Flush moves the cursor to location where last write started and clears the text written using previous Write.
func (w *Writer) Clear(linesToClear int) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	if linesToClear == 0 {
		return
	}

	if linesToClear < 0 || linesToClear > w.lineCount {
		linesToClear = w.lineCount
	}

	for i := 0; i < linesToClear; i++ {
		fmt.Fprintf(w.Out, "%c[%dA", esc, 0) // move the cursor up
		fmt.Fprintf(w.Out, "%c[2K\r", esc)   // clear the line

		w.lineCount--
	}
}

// GetTermDimensions returns the width and height of the current terminal
func (w *Writer) GetTermDimensions() (int, int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80, 25
	}
	splits := strings.Split(strings.Trim(string(out), "\n"), " ")
	height, err := strconv.ParseInt(splits[0], 0, 0)
	width, err1 := strconv.ParseInt(splits[1], 0, 0)
	if err != nil || err1 != nil {
		return 80, 25
	}
	return int(width), int(height)
}
