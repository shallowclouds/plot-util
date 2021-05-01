package plot

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shallowclouds/scp-util/ls"
	"github.com/shallowclouds/scp-util/ssh"
)

var (
	NameRegexp = regexp.MustCompile(`plot-k32-\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-[[:alnum:]]{64}\.plot(.tmp)?`)
)

type FarmerStatus struct {
	Plots    []string
	TmpPlots []string
}

func GetFarmerStatus(rs *ssh.RemoteServer, ps *ssh.RemoteServer, dir string) (*FarmerStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	lsArgs, parser := ls.FullCommand(dir)
	data, err := rs.ExecOnce(ctx, ps, "ls", lsArgs...)
	if err != nil {
		return nil, errors.WithMessage(err, "get remote dir status err")
	}

	files, err := parser(data)
	if err != nil {
		return nil, errors.WithMessagef(err, "parse ls output err")
	}

	status := new(FarmerStatus)
	for _, file := range files {
		if !NameRegexp.MatchString(file.Name) {
			continue
		}
		if strings.HasSuffix(file.Name, ".tmp") {
			status.TmpPlots = append(status.TmpPlots, file.Name)
			continue
		}
		status.Plots = append(status.Plots, file.Name)
	}

	return status, nil
}
