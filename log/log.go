package log

import (
	"context"
	"fmt"
	"log"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gzwillyy/components/log/klog"
)

// InfoLogger 表示以特定的详细程度记录非错误消息的能力.
type InfoLogger interface {

	// Info 以给定的键/值对作为上下文记录非错误消息.
	// 应该使用 msg 参数来添加一些常量描述日志行.
	// 然后可以使用键/值对添加额外的变量信息. 键/值对应该交替字符串键和任意值

	Info(msg string, fields ...Field)
	Infof(format string, v ...interface{})
	Infow(msg string, keysAndValues ...interface{})

	// Enabled 测试此 InfoLogger 是否已启用.
	// 例如 命令行标志可用于设置日志记录详细程度和禁用一些信息日志.
	Enabled() bool
}

// Logger 表示记录消息的能力，包括错误和非错误.
type Logger interface {

	// InfoLogger 所有记录器都实现了 InfoLogger.
	// 直接在 Logger 值上调用 InfoLogger 方法等同于在 V(0) InfoLogger 上调用它们.
	// 例如，logger.Info() 产生与 logger.V(0).Info 相同的结果.

	InfoLogger
	Debug(msg string, fields ...Field)
	Debugf(format string, v ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Warn(msg string, fields ...Field)
	Warnf(format string, v ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Error(msg string, fields ...Field)
	Errorf(format string, v ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Panic(msg string, fields ...Field)
	Panicf(format string, v ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Fatal(msg string, fields ...Field)
	Fatalf(format string, v ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	// V 返回特定详细级别的 InfoLogger 值.
	// 更高的详细级别意味着日志消息不太重要. 传递小于零的日志级别是非法的.
	V(level Level) InfoLogger
	Write(p []byte) (n int, err error)

	// WithValues 将一些上下文的键值对添加到记录器.
	// 有关键/值对如何工作的文档，请参阅信息.
	WithValues(keysAndValues ...interface{}) Logger

	// WithName 将新元素添加到记录器的名称中.
	// 连续调用 WithName 继续追加
	// 记录器名称的后缀. 强烈推荐
	// 该名称段仅包含字母、数字和连字符
	//（有关更多信息，请参阅包文档）.
	WithName(name string) Logger

	// WithContext 返回设置日志值的上下文副本.
	WithContext(ctx context.Context) context.Context

	// Flush 调用底层 Core 的 Sync 方法，刷新任何缓冲
	// 日志条目. 应用程序应注意在退出前调用 Sync.
	Flush()
}

// zapLogger 是一个使用 Zap 进行日志记录的 logr.Logger.
type zapLogger struct {
	// 注意：这看起来与 zap.SugaredLogger 非常相似，但它满足了我们对多个详细级别的需求.
	zapLogger *zap.Logger
	infoLogger
}

// noopInfoLogger 是一个 logr.InfoLogger，它总是被禁用，什么都不做.
type noopInfoLogger struct{}

func (l *noopInfoLogger) Enabled() bool                    { return false }
func (l *noopInfoLogger) Info(_ string, _ ...Field)        {}
func (l *noopInfoLogger) Infof(_ string, _ ...interface{}) {}
func (l *noopInfoLogger) Infow(_ string, _ ...interface{}) {}

var disabledInfoLogger = &noopInfoLogger{}

// 注意：现在，我们总是使用糖化日志记录的等价物.
// 这是必要的，因为 logr 没有定义非推荐类型，并且使用特定于 zap 的非推荐类型会使使用绑定直接给Zap.

// infoLogger 是一个 logr.InfoLogger，它使用 Zap 在特定的位置记录等级.
// 该关卡已经转换为 Zap 关卡，这 就是说 `logrLevel = -1*zapLevel`.
type infoLogger struct {
	level zapcore.Level
	log   *zap.Logger
}

func (l *infoLogger) Enabled() bool { return true }
func (l *infoLogger) Info(msg string, fields ...Field) {
	if checkedEntry := l.log.Check(l.level, msg); checkedEntry != nil {
		checkedEntry.Write(fields...)
	}
}
func (l *infoLogger) Infof(format string, args ...interface{}) {
	if checkedEntry := l.log.Check(l.level, fmt.Sprintf(format, args...)); checkedEntry != nil {
		checkedEntry.Write()
	}
}

func (l *infoLogger) Infow(msg string, keysAndValues ...interface{}) {
	if checkedEntry := l.log.Check(l.level, msg); checkedEntry != nil {
		checkedEntry.Write(handleFields(l.log, keysAndValues)...)
	}
}

// handleFields 将一堆任意键值对转换为 Zap 字段. 它需要额外的预转换 Zap 字段，用于自动附加字段，如 `error`.
func handleFields(l *zap.Logger, args []interface{}, additional ...zap.Field) []zap.Field {
	// zap.SugaredLogger.sweetenFields 的略微修改版本
	if len(args) == 0 {
		// fast-return if we have no suggared fields.
		return additional
	}

	// 与 Zap 不同，我们可以非常确定用户不会传递结构化字段（因为 logr 没有这方面的概念），所以猜测我们需要的空间要少一些.
	fields := make([]zap.Field, 0, len(args)/2+len(additional))
	for i := 0; i < len(args); {
		// 检查以防强类型的 Zap 字段，这是非法的（因为它打破了实现不可知论），所以我们可以给出更好的错误消息.
		if _, ok := args[i].(zap.Field); ok {
			l.DPanic("strongly-typed Zap Field passed to logr", zap.Any("zap field", args[i]))

			break
		}

		// 确保这不是不匹配的密钥
		if i == len(args)-1 {
			l.DPanic("odd number of arguments passed as key-value pairs for logging", zap.Any("ignored key", args[i]))

			break
		}

		// 处理一个键值对，
		// 确保键是一个字符串
		key, val := args[i], args[i+1]
		keyStr, isString := key.(string)
		if !isString {
			// 如果密钥不是字符串，则 DPanic 并停止记录
			l.DPanic(
				"non-string key argument passed to logging, ignoring all later arguments",
				zap.Any("invalid key", key),
			)

			break
		}

		fields = append(fields, zap.Any(keyStr, val))
		i += 2
	}

	return append(fields, additional...)
}

// New 通过 opts 新建记录器，可以通过命令参数自定义.
func New(opts *Options) *zapLogger {
	if opts == nil {
		opts = NewOptions()
	}

	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(opts.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}
	encodeLevel := zapcore.CapitalLevelEncoder
	// when output to local path, with color is forbidden
	if opts.Format == consoleFormat && opts.EnableColor {
		encodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "timestamp",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    encodeLevel,
		EncodeTime:     timeEncoder,
		EncodeDuration: milliSecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	loggerConfig := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       opts.Development,
		DisableCaller:     opts.DisableCaller,
		DisableStacktrace: opts.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         opts.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      opts.OutputPaths,
		ErrorOutputPaths: opts.ErrorOutputPaths,
	}

	var err error
	l, err := loggerConfig.Build(zap.AddStacktrace(zapcore.PanicLevel), zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	logger := &zapLogger{
		zapLogger: l.Named(opts.Name),
		infoLogger: infoLogger{
			log:   l,
			level: zap.InfoLevel,
		},
	}
	klog.InitLogger(l)
	zap.RedirectStdLog(l)

	return logger
}

func Init(opts *Options) {
	mu.Lock()
	defer mu.Unlock()
	std = New(opts)
}

var (
	std = New(NewOptions())
	mu  sync.Mutex
)

// SugaredLogger 返回全局加糖记录器.
func SugaredLogger() *zap.SugaredLogger {
	return std.zapLogger.Sugar()
}

// StdErrLogger 返回标准库的记录器，它在错误级别写入提供的 zap 记录器.
func StdErrLogger() *log.Logger {
	if std == nil {
		return nil
	}
	if l, err := zap.NewStdLogAt(std.zapLogger, zapcore.ErrorLevel); err == nil {
		return l
	}

	return nil
}

// StdInfoLogger 返回标准库的记录器，它在信息级别写入提供的 zap 记录器.
func StdInfoLogger() *log.Logger {
	if std == nil {
		return nil
	}
	if l, err := zap.NewStdLogAt(std.zapLogger, zapcore.InfoLevel); err == nil {
		return l
	}

	return nil
}

func (l *zapLogger) V(level Level) InfoLogger {
	if l.zapLogger.Core().Enabled(level) {
		return &infoLogger{
			level: level,
			log:   l.zapLogger,
		}
	}

	return disabledInfoLogger
}

// V 返回一个调整过的 InfoLogger
func V(level Level) InfoLogger { return std.V(level) }

func (l *zapLogger) Write(p []byte) (n int, err error) {
	l.zapLogger.Info(string(p))

	return len(p), nil
}

// WithName 将一个新的路径段添加到记录器的名称中.分段由句点连接.默认情况下，Logger是未命名的.
func WithName(s string) Logger { return std.WithName(s) }
func (l *zapLogger) WithName(name string) Logger {
	newLogger := l.zapLogger.Named(name)

	return NewLogger(newLogger)
}

// WithValues 创建一个子记录器并添加Zap字段.
func WithValues(keysAndValues ...interface{}) Logger { return std.WithValues(keysAndValues...) }
func (l *zapLogger) WithValues(keysAndValues ...interface{}) Logger {
	newLogger := l.zapLogger.With(handleFields(l.zapLogger, keysAndValues)...)

	return NewLogger(newLogger)
}

// Flush 调用底层Core的Sync方法，刷新任何缓冲的日志条目.应用程序应注意在退出之前调用Sync.
func Flush() { std.Flush() }
func (l *zapLogger) Flush() {
	_ = l.zapLogger.Sync()
}

// NewLogger 创建一个新的logr.Logger，使用给定的Zap Logger进行日志记录.
func NewLogger(l *zap.Logger) Logger {
	return &zapLogger{
		zapLogger: l,
		infoLogger: infoLogger{
			log:   l,
			level: zap.InfoLevel,
		},
	}
}

// ZapLogger 用于其他日志包装器，如klog.
func ZapLogger() *zap.Logger {
	return std.zapLogger
}

// CheckIntLevel 用于其他日志包装器，如klog，如果启用了在指定级别记录消息，则会返回该日志包装器.
func CheckIntLevel(level int32) bool {
	var lvl zapcore.Level
	if level < 5 {
		lvl = zapcore.InfoLevel
	} else {
		lvl = zapcore.DebugLevel
	}
	checkEntry := std.zapLogger.Check(lvl, "")

	return checkEntry != nil
}

// Debug 调试方法输出调试级别日志.
func Debug(msg string, fields ...Field) {
	std.zapLogger.Debug(msg, fields...)
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.zapLogger.Debug(msg, fields...)
}

// Debugf 方法输出调试级别日志.
func Debugf(format string, v ...interface{}) {
	std.zapLogger.Sugar().Debugf(format, v...)
}

func (l *zapLogger) Debugf(format string, v ...interface{}) {
	l.zapLogger.Sugar().Debugf(format, v...)
}

// Debugw 方法输出调试级别日志.
func Debugw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Debugw(msg, keysAndValues...)
}

func (l *zapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Debugw(msg, keysAndValues...)
}

// Info 方法输出信息级别日志。
func Info(msg string, fields ...Field) {
	std.zapLogger.Info(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.zapLogger.Info(msg, fields...)
}

// Infof 方法输出信息级别日志。
func Infof(format string, v ...interface{}) {
	std.zapLogger.Sugar().Infof(format, v...)
}

func (l *zapLogger) Infof(format string, v ...interface{}) {
	l.zapLogger.Sugar().Infof(format, v...)
}

// Infow 方法输出信息级别日志。
func Infow(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Infow(msg, keysAndValues...)
}

func (l *zapLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Infow(msg, keysAndValues...)
}

// Warn 方法输出警告级别日志。
func Warn(msg string, fields ...Field) {
	std.zapLogger.Warn(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.zapLogger.Warn(msg, fields...)
}

// Warnf 方法输出警告级别日志。
func Warnf(format string, v ...interface{}) {
	std.zapLogger.Sugar().Warnf(format, v...)
}

func (l *zapLogger) Warnf(format string, v ...interface{}) {
	l.zapLogger.Sugar().Warnf(format, v...)
}

// Warnw 方法输出警告级别日志。
func Warnw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Warnw(msg, keysAndValues...)
}

func (l *zapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Warnw(msg, keysAndValues...)
}

// Error 方法输出错误级别日志。
func Error(msg string, fields ...Field) {
	std.zapLogger.Error(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.zapLogger.Error(msg, fields...)
}

// Errorf 方法输出错误级别日志。
func Errorf(format string, v ...interface{}) {
	std.zapLogger.Sugar().Errorf(format, v...)
}

func (l *zapLogger) Errorf(format string, v ...interface{}) {
	l.zapLogger.Sugar().Errorf(format, v...)
}

// Errorw 方法输出错误级别日志。
func Errorw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Errorw(msg, keysAndValues...)
}

func (l *zapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Errorw(msg, keysAndValues...)
}

// Panic 方法输出紧急级别日志并关闭应用程序。
func Panic(msg string, fields ...Field) {
	std.zapLogger.Panic(msg, fields...)
}

func (l *zapLogger) Panic(msg string, fields ...Field) {
	l.zapLogger.Panic(msg, fields...)
}

// Panicf 方法输出紧急级别日志并关闭应用程序。
func Panicf(format string, v ...interface{}) {
	std.zapLogger.Sugar().Panicf(format, v...)
}

func (l *zapLogger) Panicf(format string, v ...interface{}) {
	l.zapLogger.Sugar().Panicf(format, v...)
}

// Panicw 方法输出死机级别日志。
func Panicw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Panicw(msg, keysAndValues...)
}

func (l *zapLogger) Panicw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Panicw(msg, keysAndValues...)
}

// Fatal 方法输出致命级别日志。
func Fatal(msg string, fields ...Field) {
	std.zapLogger.Fatal(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.zapLogger.Fatal(msg, fields...)
}

// Fatalf 方法输出致命级别日志。
func Fatalf(format string, v ...interface{}) {
	std.zapLogger.Sugar().Fatalf(format, v...)
}

func (l *zapLogger) Fatalf(format string, v ...interface{}) {
	l.zapLogger.Sugar().Fatalf(format, v...)
}

// Fatalw 方法输出Fatalw级别日志。
func Fatalw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Fatalw(msg, keysAndValues...)
}

func (l *zapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Fatalw(msg, keysAndValues...)
}

// L 具有指定上下文值的方法输出。
func L(ctx context.Context) *zapLogger {
	return std.L(ctx)
}

func (l *zapLogger) L(ctx context.Context) *zapLogger {
	lg := l.clone()

	if requestID := ctx.Value(KeyRequestID); requestID != nil {
		lg.zapLogger = lg.zapLogger.With(zap.Any(KeyRequestID, requestID))
	}
	if username := ctx.Value(KeyUsername); username != nil {
		lg.zapLogger = lg.zapLogger.With(zap.Any(KeyUsername, username))
	}
	if watcherName := ctx.Value(KeyWatcherName); watcherName != nil {
		lg.zapLogger = lg.zapLogger.With(zap.Any(KeyWatcherName, watcherName))
	}

	return lg
}

//nolint:predeclared
func (l *zapLogger) clone() *zapLogger {
	copy := *l

	return &copy
}
