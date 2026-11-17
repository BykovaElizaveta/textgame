package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

var (
	fastMode   bool
	noColor    bool
	forceColor bool
)

func registerFlags(fs *flag.FlagSet) {
	fs.BoolVar(&fastMode, "fast", false, "печатать текст без эффекта машинки")
	fs.BoolVar(&noColor, "no-color", false, "не использовать ANSI-цвета")
	fs.BoolVar(&forceColor, "color", false, "использовать ANSI-цвета даже не в терминале")
}

const (
	colorBold   = "1"
	colorRed    = "31"
	colorGreen  = "32"
	colorYellow = "33"
	colorCyan   = "36"
	colorGray   = "90"

	typewriterDelay = 7 * time.Millisecond
)

type renderer struct {
	out     io.Writer
	color   bool
	animate bool
}

func newRenderer(out io.Writer) *renderer {
	return &renderer{
		out:     out,
		color:   shouldUseColor(out),
		animate: !fastMode && isTerminal(out),
	}
}

func isTerminal(out io.Writer) bool {
	f, ok := out.(*os.File)
	if !ok || f == nil {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func shouldUseColor(out io.Writer) bool {
	if noColor {
		return false
	}
	if forceColor {
		return true
	}
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	if f, ok := out.(*os.File); !ok || !enableVirtualTerminal(f) {
		return false
	}
	return true
}

func (r *renderer) paint(code, text string) string {
	if !r.color || text == "" {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func (r *renderer) write(text string) {
	_, _ = fmt.Fprint(r.out, text)
}

func (r *renderer) line(text string) {
	_, _ = fmt.Fprintln(r.out, text)
}

func (r *renderer) flush() {
	if f, ok := r.out.(interface{ Flush() error }); ok {
		_ = f.Flush()
	}
}

func (r *renderer) typeLine(text string) {
	if !r.animate {
		r.line(text)
		return
	}
	for _, ch := range text {
		r.write(string(ch))
		if ch != ' ' && ch != '\n' {
			r.flush()
			time.Sleep(typewriterDelay)
		}
	}
	r.line("")
	r.flush()
}
