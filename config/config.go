package config

import (
	"io/ioutil"

	"bitbucket.org/titan098/go-dns/logging"
	"github.com/pelletier/go-toml"
)

type Domain struct {
	Prefix string `toml:"prefix"`
	Mask   int    `toml:"mask"`
}

type Soa struct {
	TTL     int    `toml:"ttl"`
	Refresh int    `toml:"refresh"`
	Retry   int    `toml:"retry"`
	Expire  int    `toml:"expire"`
	Minimum int    `toml:"minimum"`
	Mname   string `toml:"mname"`
	Rname   string `toml:"rname"`
}

type Ns struct {
	Servers []string `toml:"servers"`
}

type Config struct {
	Soa     Soa               `toml:"soa"`
	Ns      Ns                `toml:"ns"`
	Domains map[string]Domain `toml:"domains"`
}

var log = logging.SetupLogging("config")
var config *Config

func Load(filename string) error {
	log.Infof("loading config: %s", filename)
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Errorf("error loading config: %s", err.Error())
		return err
	}

	config := &Config{}
	err = toml.Unmarshal(configFile, config)
	if err != nil {
		log.Errorf("error loading config: %s", err.Error())
		return err
	}
	return nil
}
