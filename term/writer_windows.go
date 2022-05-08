//go:build windows
// +build windows

package term

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"

	"github.com/mattn/go-isatty"
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var (
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
	procFillConsoleOutputCharacter = kernel32.NewProc("FillConsoleOutputCharacterW")
	procFillConsoleOutputAttribute = kernel32.NewProc("FillConsoleOutputAttribute")
)

type short int16
type dword uint32
type word uint16

type coord struct {
	x short
	y short
}

type smallRect struct {
	left   short
	top    short
	right  short
	bottom short
}

type consoleScreenBufferInfo struct {
	size              coord
	cursorPosition    coord
	attributes        word
	window            smallRect
	maximumWindowSize coord
}

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

	f, ok := io.Writer(w.Out).(*os.File)
	if ok && !isatty.IsTerminal(f.Fd()) {
		ok = false
	}

	if !ok {
		for i := 0; i < linesToClear; i++ {
			fmt.Fprintf(w.Out, "%c[%dA", esc, 0) // move the cursor up
			fmt.Fprintf(w.Out, "%c[2K\r", esc)   // clear the line
			w.lineCount--
		}
	} else {
		fd := f.Fd()
		var csbi consoleScreenBufferInfo
		procGetConsoleScreenBufferInfo.Call(fd, uintptr(unsafe.Pointer(&csbi)))

		for i := 0; i < linesToClear; i++ {
			// move the cursor up
			csbi.cursorPosition.y--
			procSetConsoleCursorPosition.Call(fd, uintptr(*(*int32)(unsafe.Pointer(&csbi.cursorPosition))))
			// clear the line
			cursor := coord{
				x: csbi.window.left,
				y: csbi.window.top + csbi.cursorPosition.y,
			}
			var count, w dword
			count = dword(csbi.size.x)
			procFillConsoleOutputCharacter.Call(fd, uintptr(' '), uintptr(count), *(*uintptr)(unsafe.Pointer(&cursor)), uintptr(unsafe.Pointer(&w)))

			w.lineCount--
		}
	}
}

// GetTermDimensions returns the width and height of the current terminal
func (w *Writer) GetTermDimensions() (int, int) {
	f, ok := io.Writer(w.Out).(*os.File)
	if !ok {
		return 80, 25
	}
	fd := f.Fd()
	var csbi consoleScreenBufferInfo
	procGetConsoleScreenBufferInfo.Call(fd, uintptr(unsafe.Pointer(&csbi)))
	return int(csbi.maximumWindowSize.x), int(csbi.maximumWindowSize.y)
}
