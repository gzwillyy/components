package log

import (
	"fmt"
	"strings"

	"github.com/goccy/go-json"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	flagLevel             = "log.level"
	flagDisableCaller     = "log.disable-caller"
	flagDisableStacktrace = "log.disable-stacktrace"
	flagFormat            = "log.format"
	flagEnableColor       = "log.enable-color"
	flagOutputPaths       = "log.output-paths"
	flagErrorOutputPaths  = "log.error-output-paths"
	flagDevelopment       = "log.development"
	flagName              = "log.name"

	consoleFormat = "console"
	jsonFormat    = "json"
)

type Options struct {
	OutputPaths       []string `json:"output-paths"       mapstructure:"output-paths"`       // 支持输出到多个输出，用逗号分开.支持输出到标准输出（stdout）和文件
	ErrorOutputPaths  []string `json:"error-output-paths" mapstructure:"error-output-paths"` // zap内部(非业务)错误日志输出路径，多个输出，用逗号分开
	Level             string   `json:"level"              mapstructure:"level"`              // 日志级别，优先级从低到高依次为：Debug , Info , Warn , Error , Dpanic , Panic , Fatal
	Format            string   `json:"format"             mapstructure:"format"`             // 支持的日志输出格式，目前支持 Console 和 JSON 两种. Console 其实就是 Text 格式
	DisableCaller     bool     `json:"disable-caller"     mapstructure:"disable-caller"`     // 是否开启 caller，如果开启会在日志中显示调用日志所在的文件、函数和行号
	DisableStacktrace bool     `json:"disable-stacktrace" mapstructure:"disable-stacktrace"` // 是否在Panic及以上级别禁止打印堆栈信息
	EnableColor       bool     `json:"enable-color"       mapstructure:"enable-color"`       // 是否开启颜色输出，true ，是；false，否
	Development       bool     `json:"development"        mapstructure:"development"`        // 是否是开发模式.如果是开发模式，会对DPanicLevel进行堆栈跟踪
	Name              string   `json:"name"               mapstructure:"name"`               // Logger 的名字
}

// NewOptions 创建一个带有默认参数的 Options 对象.
func NewOptions() *Options {
	return &Options{
		Level:             zapcore.InfoLevel.String(),
		DisableCaller:     false,
		DisableStacktrace: false,
		Format:            consoleFormat,
		EnableColor:       false,
		Development:       false,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}
}

// Validate 验证选项字段.
func (o *Options) Validate() []error {
	var errs []error

	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(o.Level)); err != nil {
		errs = append(errs, err)
	}

	format := strings.ToLower(o.Format)
	if format != consoleFormat && format != jsonFormat {
		errs = append(errs, fmt.Errorf("not a valid log format: %q", o.Format))
	}

	return errs
}

// Build 方法可以根据Options构建一个全局的Logger
func (o Options) Build() error {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(o.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}
	encodeLevel := zapcore.CapitalLevelEncoder
	if o.Format == consoleFormat && o.EnableColor {
		encodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zc := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       o.Development,
		DisableCaller:     o.DisableCaller,
		DisableStacktrace: o.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: o.Format,
		EncoderConfig: zapcore.EncoderConfig{
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
			EncodeName:     zapcore.FullNameEncoder,
		},
		OutputPaths:      o.OutputPaths,
		ErrorOutputPaths: o.ErrorOutputPaths,
	}

	logger, err := zc.Build(zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		return err
	}
	zap.RedirectStdLog(logger.Named(o.Name))
	zap.ReplaceGlobals(logger)

	return nil
}

// AddFlags 方法可以将 Options 的各个字段追加到传入的 pflag.FlagSet变量中
func (o Options) AddFlags(fs *pflag.FlagSet) {
	//  定义命令行参数绑定到对应的变量
	fs.StringVar(&o.Level, flagLevel, o.Level, "Minimum log output `LEVEL`.")
	fs.BoolVar(&o.DisableCaller, flagDisableCaller, o.DisableCaller, "Disable output of caller information in the log.")
	fs.BoolVar(&o.DisableStacktrace, flagDisableStacktrace,
		o.DisableStacktrace, "Disable the log to record a stack trace for all messages at or above panic level.")
	fs.StringVar(&o.Format, flagFormat, o.Format, "Log output `FORMAT`, support plain or json format.")
	fs.BoolVar(&o.EnableColor, flagEnableColor, o.EnableColor, "Enable output ansi colors in plain format logs.")
	fs.StringSliceVar(&o.OutputPaths, flagOutputPaths, o.OutputPaths, "Output paths of log.")
	fs.StringSliceVar(&o.ErrorOutputPaths, flagErrorOutputPaths, o.ErrorOutputPaths, "Error output paths of log.")
	fs.BoolVar(
		&o.Development,
		flagDevelopment,
		o.Development,
		"Development puts the logger in development mode, which changes "+
			"the behavior of DPanicLevel and takes stacktraces more liberally.",
	)
	fs.StringVar(&o.Name, flagName, o.Name, "The name of the logger.")
}

// String 方法可以将 Options 的值以 JSON 格式字符串返回
func (o Options) String() string {
	data, _ := json.Marshal(o)
	return string(data)
}
