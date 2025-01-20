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
	Routes Route `yaml:"routes"`
}

type Route struct {
	Name       string   `yaml:"name,omitempty"`
	ListenHost string   `yaml:"listenhost,omitempty"`
	ListenPort int      `yaml:"listenport,omitempty"`
	Protocol   string   `yaml:"protocol,omitempty"`
	Pattern    string   `yaml:"pattern,omitempty"`
	CertFile   string   `yaml:"certfile,omitempty"`
	KeyFile    string   `yaml:"keyfile,omitempty"`
	Targets    []Target `yaml:"targets,omitempty"`
}

type Target struct {
	Name       string `yaml:"name,omitempty"`
	PathPrefix string `yaml:"pathprefix,omitempty"`
	Protocol   string `yaml:"protocol,omitempty"`
	Host       string `yaml:"host,omitempty"`
	Port       int    `yaml:"port,omitempty"`
	CertFile   string `yaml:"certfile,omitempty"`
	KeyFile    string `yaml:"keyfile,omitempty"`
	CaCert     string `yaml:"cacert,omitempty"`
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
	if len(config.Routes.Targets) == 0 {
		return fmt.Errorf("no routes defined")
	}

	log.Debug("Number of Loaded targets ", len(config.Routes.Targets))
	for _, target := range config.Routes.Targets {
		log.Debug("Target Name: ", target.Name)
		log.Debug("Target PathPrefix: ", target.PathPrefix)
	}

	// Check for unique pathprefix values
	pathprefixes := make(map[string]bool)
	for _, target := range config.Routes.Targets {
		if target.PathPrefix == "" {
			return fmt.Errorf("pathprefix is required for target %s", target.Name)
		}
		if pathprefixes[target.PathPrefix] {
			return fmt.Errorf("duplicate pathprefix: %s", target.PathPrefix)
		}
		pathprefixes[target.PathPrefix] = true

		// Validate host and port
		if target.Host == "" {
			return fmt.Errorf("host is required for target %s", target.Name)
		}
		if target.Port <= 0 {
			return fmt.Errorf("invalid port for target %s", target.Name)
		}
	}

	return validateRoute(config.Routes)
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

	for _, target := range route.Targets {
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
