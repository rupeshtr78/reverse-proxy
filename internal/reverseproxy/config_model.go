package reverseproxy

import (
	"fmt"
	"os"
)

// Config represents the configuration for the reverse proxy.
// The Routes field contains a list of Route configurations.
type Config struct {
	Routes []Route `yaml:"routes"`
}
type Route struct {
	Name       string `yaml:"name omitempty=false"`
	ListenPort int    `yaml:"listenport omitempty=false"`
	Protocol   string `yaml:"protocol omitempty=false"`
	Pattern    string `yaml:"pattern omitempty=false"`
	Target     Target `yaml:"target omitempty=false"`
}

type Target struct {
	Name     string `yaml:"name omitempty=false"`
	Protocol string `yaml:"protocol omitempty=false"`
	Host     string `yaml:"host omitempty=false"`
	Port     int    `yaml:"port omitempty=false"`
	CertFile string `yaml:"certfile omitempty=false"`
	KeyFile  string `yaml:"keyfile omitempty=false"`
}

func (config *Config) ValidateConfig() error {
	if len(config.Routes) == 0 {
		return fmt.Errorf("no routes defined in the configuration")
	}
	for _, route := range config.Routes {
		if err := validateRoute(route); err != nil {
			return err
		}
	}

	return nil
}

func validateRoute(route Route) error {
	if route.ListenPort <= 0 || route.ListenPort > 65535 {
		return fmt.Errorf("invalid listenport for route %s", route.Name)
	}

	if route.Protocol != "http" && route.Protocol != "https" {
		return fmt.Errorf("invalid protocol for route %s", route.Name)
	}

	if route.Protocol == "https" {
		if err := validateCertPath(route.Target.CertFile); err != nil {
			return err
		}
		if err := validateCertPath(route.Target.KeyFile); err != nil {
			return err
		}
	}

	return nil

}

func validateCertPath(certPath string) error {
	// Add your certificate validation logic here
	// check if the file exists
	stat, err := os.Stat(certPath)
	if err != nil {
		return fmt.Errorf("certificate file %s does not exist", certPath)
	}
	// check if the file is a regular file
	if !stat.Mode().IsRegular() {
		return fmt.Errorf("certificate file %s is not a regular file", certPath)
	}
	// check if the file is readable
	if _, err := os.Open(certPath); err != nil {
		return fmt.Errorf("certificate file %s is not readable", certPath)
	}
	return nil
}
