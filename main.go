package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

func checkFile(f string) {
	stat, err := os.Stat(f)
	if err != nil {
		fmt.Printf("os.Stat(..) = %v\n", err)
		return
	}
	fmt.Printf("os.Stat(%q).Size = %v\n", f, stat.Size)
}

func main() {
	file, err := compileProgramCwd("myproj")
	fmt.Printf("file = %v\n", file)
	fmt.Printf("err = %v\n", err)
	checkFile(file)
	fmt.Printf("os.Remove(file) = %v\n", os.Remove(file))
	checkFile(file)
}
