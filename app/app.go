package app

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/gzwillyy/components/errors"
	cliflag "github.com/gzwillyy/components/pkg/cli/flag"
	"github.com/gzwillyy/components/pkg/cli/globalflag"
	"github.com/gzwillyy/components/pkg/term"
	"github.com/gzwillyy/components/pkg/version"
	"github.com/gzwillyy/components/pkg/version/verflag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gzwillyy/components/log"
)

var progressMessage = color.GreenString("==>")

// App 是一个cli应用的主要结构.
// 建议使用 app.NewApp() 函数创建应用.
type App struct {
	basename    string
	name        string
	description string     // 设置应用的描述
	options     CliOptions // 初始化应用程序的可选参数
	runFunc     RunFunc    // 应用程序的启动回调函数
	silence     bool       // 将应用程序设置为静默模式，在该模式下程序启动控制台不打印配置信息和版本信息
	noVersion   bool       // 应用程序不提供版本标志
	noConfig    bool       // 应用程序不提供配置标志
	commands    []*Command
	args        cobra.PositionalArgs // 将验证函数设置为有效的非标志参数
	cmd         *cobra.Command
}

// Option 定义用于初始化应用程序的可选参数
// structure.
type Option func(*App)

// WithOptions 打开应用程序的功能以从命令行读取.
// 或者从配置文件中读取参数.
func WithOptions(opt CliOptions) Option {
	return func(a *App) {
		a.options = opt
	}
}

// RunFunc 定义应用程序的启动回调函数.
type RunFunc func(basename string) error

// WithRunFunc 用于设置应用启动回调函数选项.
func WithRunFunc(run RunFunc) Option {
	return func(a *App) {
		a.runFunc = run
	}
}

// WithDescription 用于设置应用的描述.
func WithDescription(desc string) Option {
	return func(a *App) {
		a.description = desc
	}
}

// WithSilence 将应用程序设置为静默模式，在该模式下程序启动
// 控制台不打印配置信息和版本信息.
func WithSilence() Option {
	return func(a *App) {
		a.silence = true
	}
}

// WithNoVersion 设置应用程序不提供版本标志.
func WithNoVersion() Option {
	return func(a *App) {
		a.noVersion = true
	}
}

// WithNoConfig 设置应用程序不提供配置标志.
func WithNoConfig() Option {
	return func(a *App) {
		a.noConfig = true
	}
}

// WithValidArgs 将验证函数设置为有效的非标志参数.
func WithValidArgs(args cobra.PositionalArgs) Option {
	return func(a *App) {
		a.args = args
	}
}

// WithDefaultValidArgs 将默认验证函数设置为有效的非标志参数.
func WithDefaultValidArgs() Option {
	return func(a *App) {
		a.args = func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}

			return nil
		}
	}
}

// NewApp 根据给定的应用程序名称创建一个新的应用程序实例
// 第 1 步: 构建应用
func NewApp(name string, basename string, opts ...Option) *App {
	a := &App{
		name:     name,
		basename: basename,
	}

	// 选项模式 动态地配置 APP
	for _, o := range opts {
		o(a)
	}

	a.buildCommand()

	return a
}

// 第 2 步：命令行程序构建
func (a *App) buildCommand() {
	cmd := cobra.Command{
		Use:   FormatBaseName(a.basename), // 不同操作系统下的可执行文件名
		Short: a.name,
		Long:  a.description,
		// 命令错误时停止打印使用
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          a.args,
	}
	// cmd.SetUsageTemplate(usageTemplate)
	cmd.SetOut(os.Stdout) // 设置使用消息到标准输出
	cmd.SetErr(os.Stderr) // 设置错误消息到标准错误·
	cmd.Flags().SortFlags = true
	cliflag.InitFlags(cmd.Flags())

	if len(a.commands) > 0 {
		for _, command := range a.commands {
			cmd.AddCommand(command.cobraCommand())
		}
		cmd.SetHelpCommand(helpCommand(FormatBaseName(a.basename)))
	}
	if a.runFunc != nil {
		cmd.RunE = a.runCommand
	}

	// 第 3 步：命令行参数解析
	// namedFlagSets 中引用了 Pflag 包，
	// a.options.Flags() 创建并返回了一批 FlagSet，
	// a.options.Flags() 函数会将 FlagSet 进行分组.
	// 通过一个 for 循环，将 namedFlagSets 中保存的 FlagSet 添加到 Cobra 应用框架中的 FlagSet 中
	var namedFlagSets cliflag.NamedFlagSets
	if a.options != nil {
		namedFlagSets = a.options.Flags()
		fs := cmd.Flags()
		for _, f := range namedFlagSets.FlagSets {
			fs.AddFlagSet(f)
		}
	}

	// 版本信息：打印应用的版本.
	// 通过 verflag.AddFlags 可以指定版本信息.
	// 例如，App 包通过 github.com/gzwillyy/components/pkg/version 指定了以下版本信息
	if !a.noVersion {
		verflag.AddFlags(namedFlagSets.FlagSet("global"))
	}
	// 第 4 步：配置文件解析
	if !a.noConfig {
		// 通过 addConfigFlag 调用，添加了 -c, –config FILE 命令行参数，用来指定配置文件
		addConfigFlag(a.basename, namedFlagSets.FlagSet("global"))
	}
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	// 将新的全局标志集添加到 cmd FlagSet
	cmd.Flags().AddFlagSet(namedFlagSets.FlagSet("global"))

	addCmdTemplate(&cmd, namedFlagSets)
	a.cmd = &cmd
}

// Run 用于启动应用程序.
func (a *App) Run() {
	if err := a.cmd.Execute(); err != nil {
		fmt.Printf("%v %v\n", color.RedString("Error:"), err)
		os.Exit(1)
	}
}

// Command 返回应用程序内的 cobra 命令实例.
func (a *App) Command() *cobra.Command {
	return a.cmd
}

func (a *App) runCommand(cmd *cobra.Command, args []string) error {
	printWorkingDir()
	cliflag.PrintFlags(cmd.Flags())
	if !a.noVersion {
		// 显示应用版本信息
		verflag.PrintAndExitIfRequested()
	}

	if !a.noConfig {
		// 命令执行时，会将配置文件中的配置项和命令行参数绑定，并将 Viper 的配置 Unmarshal 到传入的 Options 中
		// Viper 的配置是命令行参数和配置文件配置 merge 后的配置.
		// 如果在配置文件中指定了 MySQL 的 host 配置，并且也同时指定了 –mysql.host 参数，则会优先取命令行参数设置的值.
		// 这里需要注意的是，不同于 YAML 格式的分级方式，配置项是通过点号 . 来分级的
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		if err := viper.Unmarshal(a.options); err != nil {
			return err
		}
	}

	if !a.silence {
		log.Infof("%v Starting %s ...", progressMessage, a.name)
		if !a.noVersion {
			log.Infof("%v Version: `%s`", progressMessage, version.Get().ToJSON())
		}
		if !a.noConfig {
			log.Infof("%v Config file used: `%s`", progressMessage, viper.ConfigFileUsed())
		}
	}
	if a.options != nil {
		if err := a.applyOptionRules(); err != nil {
			return err
		}
	}
	// 运行应用程序
	if a.runFunc != nil {
		return a.runFunc(a.basename)
	}

	return nil
}

// 判断选项是否可补全和打印：如果可以补全，则补全选项；如果可以打印，则打印选项的内容
func (a *App) applyOptionRules() error {
	if completeableOptions, ok := a.options.(CompleteableOptions); ok {
		if err := completeableOptions.Complete(); err != nil {
			return err
		}
	}
	// 传入的 Options 是一个实现了 CliOptions 接口的结构体变量.
	// 调用 Validate 方法来校验参数是否合法
	if errs := a.options.Validate(); len(errs) != 0 {
		return errors.NewAggregate(errs)
	}

	if printableOptions, ok := a.options.(PrintableOptions); ok && !a.silence {
		log.Infof("%v Config: `%s`", progressMessage, printableOptions.String())
	}

	return nil
}

func printWorkingDir() {
	wd, _ := os.Getwd()
	log.Infof("%v WorkingDir: %s", progressMessage, wd)
}

func addCmdTemplate(cmd *cobra.Command, namedFlagSets cliflag.NamedFlagSets) {
	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())

	// 使用信息（可选）：当用户提供无效的标志或命令时，向用户显示“使用信息”.
	// 通过 cmd.SetUsageFunc 函数，可以指定使用信息.
	// 如果不想每次输错命令打印一大堆 usage 信息，你可以通过设置 SilenceUsage: true 来关闭掉 usage
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)

		return nil
	})
	// 帮助信息：执行 -h/–help 时，输出的帮助信息.通过 cmd.SetHelpFunc 函数可以指定帮助信息.
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})
}
