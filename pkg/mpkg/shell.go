/*
Copyright © 2021 Isteevan Shetoo <isteevan.shetoo@is-info.fr>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package mpkg

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Shell struct {
	Commands []string
}

func NewShell() *Shell {
	cmd := &Shell{}
	cmd.Commands = append(cmd.Commands, "set -xe; ")
	return cmd
}

func (c *Shell) AddArgs(args string) {
	c.Commands = append(c.Commands, args+"; ")
}

func (c *Shell) Exec(dir string, args []string) error {
	var stdoutBuf, stderrBuf bytes.Buffer
	_args := append(c.Commands, args...)
	_exec := strings.Join(_args, " ")
	cmd := exec.Command("/bin/sh", "-c", _exec)
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	cmd.Dir = dir
	return cmd.Run()
}
