package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
)

const (
	CONFIG_ENV_VAR = "FIN_CONFIG_PATH"
)

func RawBank() error {
	rawDir := path.Join(Config().ProjectPath, "rawbank")
	nightwatch := path.Join(rawDir, "node_modules/.bin/nightwatch")
	configEnv := fmt.Sprintf("%s=%s", CONFIG_ENV_VAR, *CONFIG_FILE)

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, nightwatch)

	cmd.Dir = rawDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), configEnv)

	return cmd.Run()
}
