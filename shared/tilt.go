package shared

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	defaultTimeoutSeconds = 600
)

// starts from dir and tries finding the controller by stepping outside
// until root is reached.
func lookForController(name string, dir string) (string, error) {
	separatorIndex := strings.LastIndex(dir, "/")
	for separatorIndex > 0 {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return filepath.Join(dir, name), nil
		}

		separatorIndex = strings.LastIndex(dir, string(os.PathSeparator))
		dir = dir[0:separatorIndex]
	}

	return "", fmt.Errorf("failed to find controller %s", name)
}

// RunTiltForControllers executes tilt for a list of controllers.
func RunTiltForControllers(controllers ...string) env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		tiltFile := ""
		tctx, cancel := context.WithTimeout(ctx, defaultTimeoutSeconds*time.Second)

		defer cancel()

		_, dir, _, _ := runtime.Caller(0)

		for _, controller := range controllers {
			path, err := lookForController(controller, dir)
			if err != nil {
				return ctx, fmt.Errorf("controller with name %q not found", controller)
			}

			tiltFile += fmt.Sprintf("include('%s/Tiltfile')\n", path)
		}

		temp, err := os.MkdirTemp("", "tilt-ci")
		if err != nil {
			return ctx, fmt.Errorf("failed to create temp folder: %w", err)
		}

		defer os.RemoveAll(temp)

		var tiltFilePermMod os.FileMode = 0o600
		if err := os.WriteFile(filepath.Join(temp, "Tiltfile"), []byte(tiltFile), tiltFilePermMod); err != nil {
			return ctx, fmt.Errorf("failed to create tilt file %w", err)
		}

		cmd := exec.CommandContext(tctx, "tilt", "ci")
		cmd.Dir = temp

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("output from tilt: ", string(output))

			return ctx, err
		}

		return ctx, nil
	}
}
