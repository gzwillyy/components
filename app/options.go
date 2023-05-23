package app

import (
	cliflag "github.com/gzwillyy/components/pkg/cli/flag"
)

// CliOptions 从命令行提取用于读取参数的配置选项.
type CliOptions interface {
	// AddFlags 将标志添加到指定的FlagSet对象.
	// AddFlags(fs *pflag.FlagSet)
	Flags() (fss cliflag.NamedFlagSets)
	Validate() []error
}

// ConfigurableOptions 从配置文件中提取用于读取参数的配置选项.
type ConfigurableOptions interface {
	// ApplyFlags 将参数从命令行或配置文件解析到选项实例.
	ApplyFlags() []error
}

// CompleteableOptions 抽象可以完成的选项.
type CompleteableOptions interface {
	Complete() error
}

// PrintableOptions 可以打印的摘要选项.
type PrintableOptions interface {
	String() string
}
