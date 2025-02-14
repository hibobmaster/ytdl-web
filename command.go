package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

var (
	YTCmd      string
	FFprobeCmd string
)

func RunCommand(ctx context.Context, command string, flags ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, flags...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			fmt.Printf("runcommand exit error, stderr '%v'\n", string(ee.Stderr))
		}
	}
	return out, err
}

func RunCommandCh(ctx context.Context, outCh chan<- string, command string, flags ...string) error {
	r, w := io.Pipe()
	defer w.Close()
	defer r.Close()
	cmd := exec.CommandContext(ctx, command, flags...)
	// set process group so that children can be killed
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = w
	//cmd.Stderr = os.Stderr // or set this to w as well
	cmd.Stderr = w
	go func() {
		stdout := bufio.NewReader(r)
		for {
			line, err := stdout.ReadString('\n')
			if err != nil {
				return
			}
			outCh <- line
		}
	}()
	err := cmd.Run()
	if err == nil {
		// kill process group including children
		syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
	return err
}
