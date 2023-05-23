package log_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gzwillyy/components/log"
)

func Test_Options_Validate(t *testing.T) {
	opts := &log.Options{
		Level:            "test",
		Format:           "test",
		EnableColor:      true,
		DisableCaller:    false,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	errs := opts.Validate()
	expected := `[unrecognized level: "test" not a valid log format: "test"]`
	assert.Equal(t, expected, fmt.Sprintf("%s", errs))
}
