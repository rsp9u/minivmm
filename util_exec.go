package minivmm

import (
	"fmt"
	"io"
	"os/exec"
)

// Execs executes commands array. If any command occures an error, it will return std error.
func Execs(cmds [][]string) error {
	return execs(cmds, false)
}

// ExecsIgnoreErr executes commands array. If any command occures an error, it will be ignored.
func ExecsIgnoreErr(cmds [][]string) {
	execs(cmds, true)
}

func execs(cmds [][]string, ignoreErr bool) error {
	for _, cmd := range cmds {
		c := exec.Command(cmd[0], cmd[1:]...)
		stderr, _ := c.StderrPipe()
		if err := c.Start(); err != nil && !ignoreErr {
			return fmt.Errorf("%v: failed to start command %v", err, cmd)
		}
		msg, _ := io.ReadAll(stderr)
		if err := c.Wait(); err != nil && !ignoreErr {
			return fmt.Errorf("%v, %v: %s", err, cmd, msg)
		}
	}
	return nil
}

// ExecsStdout executes commands array and retunrs an array of stdout.
func ExecsStdout(cmds [][]string) ([]string, error) {
	msgs := []string{}
	for _, cmd := range cmds {
		c := exec.Command(cmd[0], cmd[1:]...)
		stdout, _ := c.StdoutPipe()
		stderr, _ := c.StderrPipe()
		if err := c.Start(); err != nil {
			return msgs, fmt.Errorf("%v: failed to start command %v", err, cmd)
		}
		msgStdout, _ := io.ReadAll(stdout)
		msgStderr, _ := io.ReadAll(stderr)
		if err := c.Wait(); err != nil {
			return msgs, fmt.Errorf("%v, %v: %s", err, cmd, msgStderr)
		}
		msgs = append(msgs, string(msgStdout))
	}
	return msgs, nil
}
