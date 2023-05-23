package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 定义公共日志字段.
const (
	KeyRequestID   string = "requestID"
	KeyUsername    string = "username"
	KeyWatcherName string = "watcher"
)

// Field 是底层日志框架中字段结构的别名.
type Field = zapcore.Field

// Level 是底层日志框架中级别结构的别名.
type Level = zapcore.Level

var (
	// DebugLevel 日志通常是大量的，并且通常在生产.
	DebugLevel = zapcore.DebugLevel
	// InfoLevel 是默认的日志记录优先级.
	InfoLevel = zapcore.InfoLevel
	// WarnLevel 日志比信息更重要，但不需要单独的人工审查.
	WarnLevel = zapcore.WarnLevel
	// ErrorLevel 日志具有高优先级. 如果应用程序运行平稳，则不应生成任何错误级别的日志.
	ErrorLevel = zapcore.ErrorLevel
	// PanicLevel 记录一条消息，然后 panics.
	PanicLevel = zapcore.PanicLevel
	// FatalLevel 记录一条消息，然后调用 os.Exit(1).
	FatalLevel = zapcore.FatalLevel
)

// zap 类型函数的别名.
var (
	Any         = zap.Any
	Array       = zap.Array
	Object      = zap.Object
	Binary      = zap.Binary
	Bool        = zap.Bool
	Bools       = zap.Bools
	ByteString  = zap.ByteString
	ByteStrings = zap.ByteStrings
	Complex64   = zap.Complex64
	Complex64s  = zap.Complex64s
	Complex128  = zap.Complex128
	Complex128s = zap.Complex128s
	Duration    = zap.Duration
	Durations   = zap.Durations
	Err         = zap.Error
	Errors      = zap.Errors
	Float32     = zap.Float32
	Float32s    = zap.Float32s
	Float64     = zap.Float64
	Float64s    = zap.Float64s
	Int         = zap.Int
	Ints        = zap.Ints
	Int8        = zap.Int8
	Int8s       = zap.Int8s
	Int16       = zap.Int16
	Int16s      = zap.Int16s
	Int32       = zap.Int32
	Int32s      = zap.Int32s
	Int64       = zap.Int64
	Int64s      = zap.Int64s
	Namespace   = zap.Namespace
	Reflect     = zap.Reflect
	Stack       = zap.Stack
	String      = zap.String
	Stringer    = zap.Stringer
	Strings     = zap.Strings
	Time        = zap.Time
	Times       = zap.Times
	Uint        = zap.Uint
	Uints       = zap.Uints
	Uint8       = zap.Uint8
	Uint8s      = zap.Uint8s
	Uint16      = zap.Uint16
	Uint16s     = zap.Uint16s
	Uint32      = zap.Uint32
	Uint32s     = zap.Uint32s
	Uint64      = zap.Uint64
	Uint64s     = zap.Uint64s
	Uintptr     = zap.Uintptr
	Uintptrs    = zap.Uintptrs
)
