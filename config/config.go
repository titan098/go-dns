package config

import (
	"fmt"
	"io/ioutil"

	"bitbucket.org/titan098/go-dns/logging"
	"github.com/pelletier/go-toml"
)

var log = logging.SetupLogging("config")

type Domain struct {
	Domain string
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

type DNS struct {
	Port     int    `toml:"port"`
	Protocol string `toml:"protocol"`
	Soa      Soa    `toml:"soa"`
	Ns       Ns     `toml:"ns"`
}

type Config struct {
	DNS     DNS               `toml:"dns"`
	Domains map[string]Domain `toml:"domains"`
}

var config *Config

func (soa *Soa) String(domain string) string {
	return fmt.Sprintf("%s SOA %s %s 1 %d %d %d %d", domain, soa.Mname, soa.Rname, soa.Refresh, soa.Retry, soa.Expire, soa.TTL)
}

func GetConfig() *Config {
	return config
}

func Load(filename string) (*Config, error) {
	if config != nil {
		return config, nil
	}

	log.Infof("loading config: %s", filename)
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Errorf("error loading config: %s", err.Error())
		return nil, err
	}

	config = &Config{}
	err = toml.Unmarshal(configFile, config)
	if err != nil {
		log.Errorf("error loading config: %s", err.Error())
		return nil, err
	}
	return config, nil
}
