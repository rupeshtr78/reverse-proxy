package reverseproxy

// Config represents the configuration for the reverse proxy.
// The Routes field contains a list of Route configurations.
type Config struct {
	Routes []Route `yaml:"routes"`
}
type Route struct {
	Name       string `yaml:"name omitempty=false"`
	ListenPort int    `yaml:"listenport omitempty=false"`
	Protocol   string `yaml:"protocol omitempty=false"`
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
