//go:build windows

package runner

import "os/exec"

func configureCommandProcessGroupCancel(cmd *exec.Cmd) {}

func killCommandProcessGroup(cmd *exec.Cmd) {}
