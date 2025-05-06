package app

import (
	"strings"

	"github.com/spf13/pflag"
)

func initFlag() {
	pflag.CommandLine.SetNormalizeFunc(WordSepNormalizeFunc)
}

// WordSepNormalizeFunc 更改所有包含“_”分隔符的标志.
func WordSepNormalizeFunc(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
	}

	return pflag.NormalizedName(name)
}
