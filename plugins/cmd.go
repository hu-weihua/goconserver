package plugins

import (
	"fmt"
	"github.com/chenglch/goconserver/common"
	"github.com/kr/pty"
	"os"
	"os/exec"
	"strings"
)

const (
	DRIVER_CMD = "cmd"
)

type CommondConsole struct {
	node    string // session name
	cmd     string
	params  []string
	command *exec.Cmd
	pty     *os.File
}

func init() {
	DRIVER_INIT_MAP[DRIVER_CMD] = NewCommondConsole
	DRIVER_VALIDATE_MAP[DRIVER_CMD] = func(name string, params map[string]string) error {
		if _, ok := params["cmd"]; !ok {
			return common.NewErr(common.INVALID_PARAMETER, fmt.Sprintf("%s: Please specify the parameter cmd", name))
		}
		return nil
	}
}

func NewCommondConsole(node string, params map[string]string) (ConsolePlugin, error) {
	cmd := params["cmd"]
	args := strings.Split(cmd, " ")
	return &CommondConsole{node: node, cmd: args[0], params: args[1:]}, nil
}

func (c *CommondConsole) Start() (*BaseSession, error) {
	var err error
	c.command = exec.Command(c.cmd, c.params...)
	c.pty, err = pty.Start(c.command)
	if err != nil {
		plog.ErrorNode(c.node, err.Error())
		return nil, err
	}
	tty := common.Tty{}
	ttyWidth, ttyHeight, err := tty.GetSize(os.Stdin)
	if err != nil {
		plog.DebugNode(c.node, "Could not get tty size, use 80,80 as default")
		ttyHeight = 80
		ttyWidth = 80
	}
	if err = tty.SetSize(c.pty, ttyWidth, ttyHeight); err != nil {
		plog.ErrorNode(c.node, err.Error())
		return nil, err
	}
	return &BaseSession{In: c.pty, Out: c.pty, Session: c}, nil
}

func (c *CommondConsole) Close() error {
	c.pty.Close()
	return c.command.Process.Kill()
}

func (c *CommondConsole) Wait() error {
	return c.command.Wait()
}
