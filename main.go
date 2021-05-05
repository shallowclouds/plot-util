package main

import (
	"flag"
	"time"

	"github.com/shallowclouds/scp-util/plot"
	"github.com/shallowclouds/scp-util/ssh"
	"github.com/shallowclouds/scp-util/util"
	"github.com/sirupsen/logrus"
)

const (
	version = "v0"
)

type Server struct {
	Server *ssh.RemoteServer
	Conf   *plot.Host
}

var (
	harvesterServer *ssh.RemoteServer
	proxy           *ssh.RemoteServer
	farmersServers  map[string]*ssh.RemoteServer
	hProxy          *ssh.RemoteServer

	harvester *Server
	farmers   map[string]*Server
)

func getJobStatus() {
	logrus.Infof("%10s%6s%6s", "farmer", "phase", "time")
	for h, farmer := range farmers {
		jobs, err := plot.GetJobStatus(farmer.Server, proxy, farmer.Conf.ChiaDir)
		if err != nil {
			logrus.WithError(err).Error("failed to get query job status for %s", h)
		}

		for _, job := range jobs {
			logrus.Infof("%10s%6s%6s", h, job.Phase, job.TimeCost)
		}
	}
}

func main() {
	logrus.Infof("version: %s", version)
	debug := flag.Bool("debug", false, "")
	fetch := flag.Bool("fetch", false, "")
	loop := flag.Bool("loop", false, "")
	status := flag.Bool("status", false, "")

	flag.Parse()

	if debug != nil && *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

start:

	conf := plot.MustReadConfig("conf/hosts.yaml")
	proxy, harvesterServer, farmersServers = plot.MustInitServers(conf)

	harvester = &Server{
		Server: harvesterServer,
		Conf:   &conf.Harvester,
	}

	farmers = make(map[string]*Server)
	for _, f := range conf.Farmers {
		f := f
		s := &Server{
			Server: nil,
			Conf:   &f,
		}
		for _, ser := range farmersServers {
			if ser.Host == f.Name {
				s.Server = ser
			}
		}
		farmers[s.Server.Host] = s
	}

	logrus.Infof("harvester: %s", harvester.Conf.Name)
	logrus.Infof("proxy: %s", proxy.Host)
	logrus.Infof("farmers: %d", len(conf.Farmers))

	if status != nil && *status {
		getJobStatus()
		return
	}

	// var hProxy *ssh.RemoteServer
	if conf.HarvesterProxy != nil {
		hProxy = &ssh.RemoteServer{
			Host:     conf.HarvesterProxy.Name,
			IP:       conf.HarvesterProxy.IP,
			Port:     conf.HarvesterProxy.Port,
			Username: conf.HarvesterProxy.Username,
			Password: "",
		}
	}

	hPlots, err := plot.GetFarmerStatus(harvesterServer, hProxy, conf.Harvester.DstDir)
	if err != nil {
		panic(err)
	}

	hsMap := util.ArrayToMap(hPlots.Plots)
	logrus.Infof("total %d plots on harvester", len(hPlots.Plots))
	for _, p := range hPlots.Plots {
		logrus.Infof("%s", p)
	}

	newPlots := map[string][]string{}

	for _, f := range farmers {
		_, _ = plot.GetJobStatus(f.Server, proxy, f.Conf.DstDir+"-backup")
		plots, err := plot.GetFarmerStatus(f.Server, proxy, f.Conf.DstDir)
		if err != nil {
			logrus.WithError(err).Errorf("failed to get plot status for %s", f.Conf.Name)
			continue
		}
		logrus.Infof("farmer %s has total %d plots", f.Conf.Name, len(plots.Plots))
		var nPlots []string
		for _, k := range plots.Plots {
			if _, ok := hsMap[k]; !ok {
				nPlots = append(nPlots, k)
				continue
			}
		}
		if len(nPlots) != 0 {
			newPlots[f.Conf.Name] = nPlots
		}
	}

	var latest string
	var server *Server
	for f, plots := range newPlots {
		logrus.Infof("%s has %d new slots:", f, len(plots))
		for _, p := range plots {
			latest = p
			server = farmers[f]
			logrus.Debug(p)
		}
	}

	if fetch != nil && *fetch {
		if latest != "" && server != nil {
			logrus.Infof("try to pull plot %s from %s", latest, server.Conf.Name)
			plot.FetchPlot(
				server.Conf.DstDir,
				latest, harvester.Conf.DstDir,
				"", true,
				proxy,
				server.Server,
			)
		}

		if loop != nil && *loop {
			logrus.Infof("fetching next plot in 60s")
			time.Sleep(time.Second * 60)
			goto start
		}
	}
}
