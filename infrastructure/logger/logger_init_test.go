package logger_test

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

// Test_Logger_Init_Subprocess covers different init() branches by running
// subprocess tests with various LOG_LEVEL/LOG_MODE/VERIFY_MODE env values.
// Each subprocess imports the logger package, triggering init() with the given env.
func Test_Logger_Init_Subprocess(t *testing.T) { //nolint:paralleltest
	if os.Getenv("LOGGER_INIT_SUBPROCESS") == "1" {
		// We are the subprocess. Just import the logger (already done by package import).
		// If init() panics, the test will fail.
		return
	}

	cases := []struct {
		name string
		env  map[string]string
	}{
		{"LOG_LEVEL_debug", map[string]string{"LOG_LEVEL": "debug"}},
		{"LOG_LEVEL_warn", map[string]string{"LOG_LEVEL": "warn"}},
		{"LOG_LEVEL_error", map[string]string{"LOG_LEVEL": "error"}},
		{"LOG_LEVEL_fatal", map[string]string{"LOG_LEVEL": "fatal"}},
		{"LOG_LEVEL_invalid", map[string]string{"LOG_LEVEL": "invalid_value"}},
		{"VERIFY_MODE_enable", map[string]string{"VERIFY_MODE": "enable"}},
		{"LOG_MODE_few", map[string]string{"LOG_MODE": "few"}},
		{"LOG_MODE_minimum", map[string]string{"LOG_MODE": "minimum"}},
		{"LOG_MODE_other", map[string]string{"LOG_MODE": "other_value"}},
	}

	for _, tc := range cases { //nolint:paralleltest
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^Test_Logger_Init_Subprocess$", "-test.count=1") //nolint:gosec
			cmd.Env = append(os.Environ(), "LOGGER_INIT_SUBPROCESS=1")
			for k, v := range tc.env {
				cmd.Env = append(cmd.Env, k+"="+v)
			}
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("subprocess failed with env %v: %v\noutput: %s", tc.env, err, out)
			}
		})
	}
}
