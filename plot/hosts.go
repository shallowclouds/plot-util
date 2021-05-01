package plot

import (
	"github.com/jinzhu/configor"
	"github.com/shallowclouds/scp-util/ssh"
)

type Host struct {
	Name     string `yaml:"Name"`
	IP       string `yaml:"IP"`
	Username string `yaml:"Username"`
	Port     int    `yaml:"Port"`
	TmpDir   string `yaml:"TmpDir"`
	DstDir   string `yaml:"DstDir"`
}

type Config struct {
	Harvester Host   `yaml:"Harvester"`
	HarvesterProxy *Host `yaml:"HarvesterProxy"`
	Farmers   []Host `yaml:"Farmers"`
	Proxy   Host `yaml:"Proxy"`
}

func MustReadConfig(path string) *Config {
	config := new(Config)
	if err := configor.Load(config, path); err != nil {
		panic(err)
	}

	return config
}

func MustInitServers(conf *Config) (*ssh.RemoteServer, *ssh.RemoteServer, map[string]*ssh.RemoteServer) {
	ps := &ssh.RemoteServer{
		Host: conf.Proxy.Name,
		IP: conf.Proxy.IP,
		Port: conf.Proxy.Port,
		Username: conf.Proxy.Username,
	}
	
	hs := &ssh.RemoteServer{
		Host:     conf.Harvester.Name,
		IP:       conf.Harvester.IP,
		Port:     conf.Harvester.Port,
		Username: conf.Harvester.Username,
		Password: "",
	}

	farmers := make(map[string]*ssh.RemoteServer, len(conf.Farmers))
	for _, s := range conf.Farmers {
		f := &ssh.RemoteServer{
			Host:     s.Name,
			IP:       s.IP,
			Port:     s.Port,
			Username: s.Username,
			Password: "",
		}
		farmers[s.Name] = f
	}

	return ps, hs, farmers
}
