package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/shallowclouds/scp-util/util"
	"github.com/sirupsen/logrus"
)

const (
	BinPath = "/usr/bin/ssh"
)

type RemoteServer struct {
	Host     string
	IP       string
	Port     int
	Username string
	Password string // TODO: use password with sshpass
}

func (rs *RemoteServer) ProxyCommand() string {
	return fmt.Sprintf("%s -W %%h:%%p %s@%s -p %d", BinPath, rs.Username, rs.IP, rs.Port)
}

func (rs *RemoteServer) String() string {
	return fmt.Sprintf("%s: %s@%s:%d", rs.Host, rs.Username, rs.IP, rs.Port)
}

// ExecOnce executes a command on the remote server and returns the results if the command finished.
func (rs *RemoteServer) ExecOnce(ctx context.Context, proxy *RemoteServer, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...) // only for log.
	if proxy != nil {
		logrus.Debugf("star to exec command on %s via %s: %s", rs.String(), proxy.String(), cmd.String())
	} else {
		logrus.Debugf("star to exec command on %s: %s", rs.String(), cmd.String())
	}

	var all []string

	if proxy != nil {
		all = append(all, fmt.Sprintf("-oProxyCommand=%s", proxy.ProxyCommand()))
	}

	all = append(all,
		"-tt",
		"-oStrictHostKeyChecking=no",
		fmt.Sprintf("%s@%s", rs.Username, rs.IP),
		"-p",
		strconv.FormatInt(int64(rs.Port), 10),
		fmt.Sprintf("%s %s", command, strings.Join(args, " ")),
	)

	cmd = exec.CommandContext(ctx, BinPath, all...)

	logrus.Debugf("executing: %s", cmd.String())

	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	defer util.LogTimeCost("exec remote command")()

	output, err := cmd.Output()
	if err != nil {
		output, _ := ioutil.ReadAll(stderr)
		return "", errors.WithMessagef(err, "read output err: %s", string(output))
	}

	return string(output), nil
}
