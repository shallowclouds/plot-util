package plot

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shallowclouds/scp-util/ls"
	"github.com/shallowclouds/scp-util/ssh"
	"github.com/sirupsen/logrus"
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

type JobStatus struct {
	PlotID   string
	K        string
	TmpDir   string
	DstDir   string
	TimeCost string
	Phase    string
	TmpSize  string
	Pid      string
	Stat     string
	Mem      string
	UserTime string
	SysTime  string
	IOTime   string
}

const (
	columnPlotID uint8 = iota
	columnK
	columnTmpDir
	columnDstDir
	columnTimeCost
	columnPhase
	columnTmpSize
	columnPid
	columnStat
	columnMem
	columnUserTime
	columnSysTime
	columnIOTime
)

var (
	validStatusLineRegexp = regexp.MustCompile(`([[:alnum:]]{8})\s+([0-9]+)\s+([a-zA-Z\-0-9\/]+)\s+([a-zA-Z\-0-9\/]+)\s+([0-9]+:[0-9]+)\s+([0-9]+:[0-9]+)\s+(\w+)\s+([0-9]+)\s+([a-zA-Z]+)\s+([0-9.GMKB]+)\s+([\w:]+)\s+([\w:]+)\s+([\w:]+)`)
)

// GetJobStatus gets the plotting job status from plotman.
func GetJobStatus(rs *ssh.RemoteServer, ps *ssh.RemoteServer, dir string) ([]*JobStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	data, err := rs.ExecOnce(ctx, ps, filepath.Join(dir, "venv/bin/plotman"), "status")
	if err != nil {
		return nil, errors.WithMessage(err, "exec job status command err")
	}

	lines := strings.Split(data, "\n")
	jobs := make([]*JobStatus, 0, len(lines))
	for _, line := range lines {
		subs := validStatusLineRegexp.FindStringSubmatch(line)
		if subs == nil || len(subs) != 14 {
			continue
		}

		subs = subs[1:]
		jobs = append(jobs, &JobStatus{
			PlotID:   subs[columnPlotID],
			K:        subs[columnK],
			TmpDir:   subs[columnTmpDir],
			DstDir:   subs[columnDstDir],
			TimeCost: subs[columnTimeCost],
			Phase:    subs[columnPhase],
			TmpSize:  subs[columnTmpSize],
			Pid:      subs[columnPid],
			Stat:     subs[columnStat],
			Mem:      subs[columnMem],
			UserTime: subs[columnUserTime],
			SysTime:  subs[columnSysTime],
			IOTime:   subs[columnIOTime],
		})
	}

	logrus.Debugf("got %d running jobs on %s", len(jobs), rs.Host)

	return jobs, nil
}
