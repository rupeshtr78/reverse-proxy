package reverseproxy

type Config struct {
	Routes []Route `yaml:"routes"`
}
type Route struct {
	Name          string   `yaml:"name omitempty=false"`
	ListenPort    int      `yaml:"listen_port omitempty=false"`
	Protocol      string   `yaml:"protocol omitempty=false"`
	ProxyProtocol string   `yaml:"proxy_protocol omitempty=false"`
	Targets       []Target `yaml:"targets omitempty=false"`
}

type Target struct {
	Name     string `yaml:"name omitempty=false"`
	Protocol string `yaml:"protocol omitempty=false"`
	Host     string `yaml:"host omitempty=false"`
	Port     int    `yaml:"port omitempty=false"`
}
