package executor

import (
	"io"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type executor struct {
}

type Executor interface {
	NewCommand(cmd string, args ...string) Command
}

func New() Executor {
	return &executor{}
}

func (e *executor) NewCommand(cmd string, args ...string) Command {
	return &command{
		cmd: exec.Command(cmd, args...),
	}
}

type command struct {
	cmd *exec.Cmd
}

type Command interface {
	WithDir(dir string) Command
	WithCombinedOutput(out io.Writer) Command
	Run() error
}

func (c *command) WithDir(dir string) Command {
	c.cmd.Dir = dir
	return c
}

func (c *command) WithCombinedOutput(out io.Writer) Command {
	c.cmd.Stdout = out
	c.cmd.Stderr = out
	return c
}

func (c *command) Run() error {
	log.Debugf("Executing: %s %v", c.cmd.Path, c.cmd.Args)
	return c.cmd.Run()
}
