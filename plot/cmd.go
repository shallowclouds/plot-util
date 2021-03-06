package plot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/shallowclouds/scp-util/ssh"
	"github.com/shallowclouds/scp-util/util"
	"github.com/sirupsen/logrus"
)

func FetchPlot(remoteDir, file, dstDir, tmpDir string, direct bool, proxy, remote *ssh.RemoteServer) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*10)
	defer cancel()

	var args []string
	if proxy != nil {
		// args = []string{
		// 	fmt.Sprintf("-oProxyCommand=%s", proxy.ProxyCommand()),
		// 	"-P",
		// 	strconv.FormatInt(int64(remote.Port), 10),
		// 	fmt.Sprintf("%s@%s:%s/%s", remote.Username, remote.IP, remoteDir, file),
		// 	fmt.Sprintf("%s/%s.tmp", dstDir, file),
		// }

		// Use rsync instead.
		args = []string{
			"-Pa",
			"--append",
			"--inplace",
			"--no-whole-file",
			"--progress",
			"-e",
			fmt.Sprintf("ssh -oProxyCommand='%s' -p %d", proxy.ProxyCommand(), remote.Port),
			fmt.Sprintf("%s@%s:%s/%s", remote.Username, remote.IP, remoteDir, file),
			fmt.Sprintf("%s/%s.tmp", dstDir, file),
		}
	}

	// cmd := exec.CommandContext(ctx, "scp", args...)
	cmd := exec.CommandContext(ctx, "rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	logrus.Debugf("execting copy command: %s", cmd.String())

	defer util.LogTimeCost(fmt.Sprintf("fetch slot %s", file))()
	if err := cmd.Start(); err != nil {
		logrus.WithError(err).Error("failed to start fetch plot command")
		return
	}

	if err := cmd.Wait(); err != nil {
		logrus.WithError(err).Error("failed to wait fetch plot command")
		return
	}

	if err := os.Rename(fmt.Sprintf("%s/%s.tmp", dstDir, file), fmt.Sprintf("%s/%s", dstDir, file)); err != nil {
		logrus.WithError(err).Error("failed to rename file")
		return
	}
}
