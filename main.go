package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/executable"
)

func compileProgramCwd(buildID string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "unable to get current working directory")
	}

	goFileSearchPattern := filepath.Join(cwd, "*.go")
	if matches, err := filepath.Glob(goFileSearchPattern); err != nil || len(matches) == 0 {
		return "", errors.Errorf("Failed to find go files for 'go build' matching %s", goFileSearchPattern)
	}

	f, err := os.CreateTemp("", fmt.Sprintf("pulumi-go.%s.*", buildID))
	if err != nil {
		return "", errors.Wrap(err, "unable to create go program temp file")
	}

	if err := f.Close(); err != nil {
		return "", errors.Wrap(err, "unable to close go program temp file")
	}
	outfile := f.Name()

	gobin, err := executable.FindExecutable("go")
	if err != nil {
		return "", errors.Wrap(err, "unable to find 'go' executable")
	}
	buildCmd := exec.Command(gobin, "build", "-o", outfile, cwd)
	buildCmd.Stdout, buildCmd.Stderr = os.Stdout, os.Stderr

	if err := buildCmd.Run(); err != nil {
		return "", errors.Wrap(err, "unable to run `go build`")
	}

	return outfile, nil
}

func execProgramCmd(cmd *exec.Cmd, env []string) error {
	cmd.Env = env
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err := cmd.Run()
	if err == nil {
		// Go program completed successfully
		return nil
	}

	// error handling
	exiterr, ok := err.(*exec.ExitError)
	if !ok {
		return errors.Wrapf(exiterr, "command errored unexpectedly")
	}

	// retrieve the status code
	status, ok := exiterr.Sys().(syscall.WaitStatus)
	if !ok {
		return errors.Wrapf(exiterr, "program exited unexpectedly")
	}

	// If the program ran, but exited with a non-zero error code. This will happen often, since user
	// errors will trigger this.  So, the error message should look as nice as possible.
	return errors.Errorf("program exited with non-zero exit code: %d", status.ExitStatus())
}

func checkFile(f string) {
	stat, err := os.Stat(f)
	if err != nil {
		fmt.Printf("os.Stat(..) = %v\n", err)
		return
	}
	fmt.Printf("os.Stat(%q).Size = %v\n", f, stat.Size)
}

func main() {
	if os.Getenv("MODE") == "test" {
		fmt.Printf("TEST\n")
		return
	}

	file, err := compileProgramCwd("myproj")
	fmt.Printf("file = %v\n", file)
	fmt.Printf("err = %v\n", err)
	checkFile(file)

	env := os.Environ()
	env = append(env, "MODE=test")

	fmt.Printf("execProgramCmd() = %v\n",
		execProgramCmd(exec.Command(file), env))

	fmt.Printf("os.Remove(file) = %v\n", os.Remove(file))
	checkFile(file)
}
