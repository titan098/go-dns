package config

import (
	"fmt"
	"io/ioutil"

	"bitbucket.org/titan098/go-dns/logging"
	"github.com/pelletier/go-toml"
)

var log = logging.SetupLogging("config")

// Domain is the structure used to represent domain configurations in the config file
type Domain struct {
	Domain        string `toml:"domain"`
	ReverseDomain string `toml:"reverse_domain"`
	Prefix        string `toml:"prefix"`
	Mask          int    `toml:"mask"`
	ResponseType  string `toml:"response_type"`
}

// Soa contains the parameters for the SOA record returned for this domain
type Soa struct {
	TTL     int    `toml:"ttl"`
	Refresh int    `toml:"refresh"`
	Retry   int    `toml:"retry"`
	Expire  int    `toml:"expire"`
	Minimum int    `toml:"minimum"`
	Mname   string `toml:"mname"`
	Rname   string `toml:"rname"`
}

// Ns is a list of Nameservers returned when an NS record is returned
type Ns struct {
	Servers []string `toml:"servers"`
}

// DNS is a the toplevel collection of times returned when the config is parsed
type DNS struct {
	Port     int    `toml:"port"`
	Protocol string `toml:"protocol"`
	Domain   Domain `toml:"domain"`
	Soa      Soa    `toml:"soa"`
	Ns       Ns     `toml:"ns"`
}

// Config is the main configuration object
type Config struct {
	DNS       DNS               `toml:"dns"`
	SubDomain map[string]Domain `toml:"subdomain"`
	Static    map[string]Domain `toml:"static"`
}

var config *Config

// String returns formatted string for an SOA record
func (soa *Soa) String(domain string) string {
	return fmt.Sprintf("%s SOA %s %s 1 %d %d %d %d", domain, soa.Mname, soa.Rname, soa.Refresh, soa.Retry, soa.Expire, soa.TTL)
}

// GetConfig returns the Config structure for this application
func GetConfig() *Config {
	return config
}

// Load will load the configuration from a passed filename
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
