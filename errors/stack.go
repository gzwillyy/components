package errors

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// Frame 表示堆栈帧中的程序计数器.
// 由于历史原因，如果Frame被解释为uintptr
// 其值表示程序计数器+1.
type Frame uintptr

// pc返回该 Frame 的程序计数器；
// 多个 Frame 可以具有相同的PC值.
func (f Frame) pc() uintptr { return uintptr(f) - 1 }

// file 返回文件的完整路径，该文件包含此Frame的pc的功能.
func (f Frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// line 返回此Frame的pc的函数源代码的行号.
func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// name返回此函数的名称（如果已知）.
func (f Frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// Format 根据fmt.Formatter接口格式化帧.
// %s源文件
// %d源行
// %n函数名称
// %v相当于%s：%d
// Format接受改变某些动词打印的标志，如下所示：
// %+s函数名和源文件相对于编译时的路径
// GOPATH由\n\t分隔（<funcname>\n\t<path>）
// %+v相当于%+s:%d
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.name())
			io.WriteString(s, "\n\t")
			io.WriteString(s, f.file())
		default:
			io.WriteString(s, path.Base(f.file()))
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.line()))
	case 'n':
		io.WriteString(s, funcname(f.name()))
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// MarshalText 将堆栈跟踪帧格式化为文本字符串.
// 输出与fmt.Sprintf（“%+v”，f）的输出相同，但没有换行符或制表符.
func (f Frame) MarshalText() ([]byte, error) {
	name := f.name()
	if name == "unknown" {
		return []byte(name), nil
	}
	return []byte(fmt.Sprintf("%s %s:%d", name, f.file(), f.line())), nil
}

// StackTrace is stack of Frames from innermost (newest) to outermost (oldest).
type StackTrace []Frame

// Format 根据fmt.Formatter接口格式化帧堆栈.
// %s列出堆栈中每个帧的源文件
// %v列出堆栈中每个帧的源文件和行号
// Format接受改变某些动词打印的标志，如下所示：
// %+v打印堆栈中每个帧的文件名、函数和行号.
func (st StackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, f := range st {
				io.WriteString(s, "\n")
				f.Format(s, verb)
			}
		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", []Frame(st))
		default:
			st.formatSlice(s, verb)
		}
	case 's':
		st.formatSlice(s, verb)
	}
}

// formatSlice 将把这个StackTrace格式化到给定的缓冲区中，作为
// Frame，仅在使用“%s”或“%v”调用时有效.
func (st StackTrace) formatSlice(s fmt.State, verb rune) {
	io.WriteString(s, "[")
	for i, f := range st {
		if i > 0 {
			io.WriteString(s, " ")
		}
		f.Format(s, verb)
	}
	io.WriteString(s, "]")
}

// stack表示程序计数器的堆栈.
type stack []uintptr

func (s *stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := Frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

func (s *stack) StackTrace() StackTrace {
	f := make([]Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = Frame((*s)[i])
	}
	return f
}

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

// funcname 删除 func.name（）报告的函数名称的路径前缀组件.
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}
