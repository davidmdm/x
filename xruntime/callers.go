package xruntime

import (
	"fmt"
	"runtime"
	"strings"
)

type Stack struct {
	Frames []runtime.Frame
}

func (stack Stack) String() string {
	var builder strings.Builder
	for _, frame := range stack.Frames {
		builder.WriteString(fmt.Sprintf("%s\n  %s:%d\n", frame.Function, frame.File, frame.Line))
	}
	return builder.String()
}

// GetCallStack returns the stack starting from where GetCallStack and going back as deep as specified by depth.
// The xruntime.Stack value satisfies the stringer interface for ease of use in common programs.
// If depth is less than 1 it defaults to a max depth of 100.
func GetCallStack(depth int) (stack Stack) {
	if depth < 1 {
		depth = 100
	}
	pc := make([]uintptr, depth)

	max := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:max])

	stack.Frames = make([]runtime.Frame, 0, max)

	emptyFrame := runtime.Frame{}
	for {
		frame, more := frames.Next()
		if frame != emptyFrame {
			stack.Frames = append(stack.Frames, frame)
		}
		if !more {
			break
		}
	}

	return
}