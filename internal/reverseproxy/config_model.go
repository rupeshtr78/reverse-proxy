package reverseproxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Config represents the configuration for the reverse proxy.
// The Routes field contains a list of Route configurations.
type Config struct {
	Routes []Route `yaml:"routes"`
}
type Route struct {
	Name       string   `yaml:"name omitempty=false"`
	ListenHost string   `yaml:"listenhost omitempty=false"`
	ListenPort int      `yaml:"listenport omitempty=false"`
	Protocol   string   `yaml:"protocol omitempty=false"`
	Pattern    string   `yaml:"pattern omitempty=false"`
	CertFile   string   `yaml:"certfile omitempty=false"`
	KeyFile    string   `yaml:"keyfile omitempty=false"`
	Target     []Target `yaml:"target omitempty=false"` // @TODO list of targets
}

type Target struct {
	Name       string `yaml:"name omitempty=false"`
	PathPrefix string `yaml:"pathprefix omitempty=false"`
	Protocol   string `yaml:"protocol omitempty=false"`
	Host       string `yaml:"host omitempty=false"`
	Port       int    `yaml:"port omitempty=false"`
	CertFile   string `yaml:"certfile omitempty=false"`
	KeyFile    string `yaml:"keyfile omitempty=false"`
	CaCert     string `yaml:"cacert omitempty=false"`
}

func (target *Target) GetTlsTransport() (*tls.Config, error) {
	if target.Protocol != "https" {
		return nil, nil
	}

	_, err := os.Stat(target.CertFile)
	if err != nil {
		log.Error("Error reading certificate file", err)
		return nil, err
	}

	_, err = os.Stat(target.KeyFile)
	if err != nil {
		log.Error("Error reading key file", err)
		return nil, err
	}

	_, err = os.Stat(target.CaCert)
	if err != nil {
		log.Error("Error reading CA certificate file", err)
		return nil, err
	}

	tlsPair, err := tls.LoadX509KeyPair(target.CertFile, target.KeyFile)
	if err != nil {
		log.Error("Error loading certificate files", err)
		return nil, err
	}

	caCert, err := os.ReadFile(target.CaCert)
	if err != nil {
		log.Error("Error reading CA certificate file", err)
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsPair},
		RootCAs:      caCertPool,
	}

	return tlsConfig, nil
}

// ValidateConfig validates the configuration for the reverse proxy.
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

// validateRoute validates a single route configuration.
func validateRoute(route Route) error {
	if route.ListenPort <= 0 || route.ListenPort > 65535 {
		return fmt.Errorf("invalid listenport for route %s", route.Name)
	}

	if route.Protocol != "http" && route.Protocol != "https" {
		return fmt.Errorf("invalid protocol for route %s", route.Name)
	}

	// if route.Protocol == "https" {
	// 	if err := validateCertPath(route.Target.CertFile); err != nil {
	// 		return err
	// 	}
	// 	if err := validateCertPath(route.Target.KeyFile); err != nil {
	// 		return err
	// 	}
	// }

	for _, target := range route.Target {
		if target.Port <= 0 || target.Port > 65535 {
			return fmt.Errorf("invalid port for target %s", target.Name)
		}

		if target.Protocol != "http" && target.Protocol != "https" {
			return fmt.Errorf("invalid protocol for target %s", target.Name)
		}

		if target.Protocol == "https" {
			if err := validateCertPath(target.CertFile); err != nil {
				return err
			}
			if err := validateCertPath(target.KeyFile); err != nil {
				return err
			}
		}
	}

	return nil

}

// validateCertPath validates the path to a certificate file.
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
