package test

import (
	"fmt"
	"log/slog"
	"os"
	"reverseproxy/pkg/logger"

	"github.com/spf13/viper"
)

func ViperPoc() {
	logger := logger.NewLogger(os.Stdout, "viper-poc", slog.LevelInfo)

	// Setup Viper
	viper.SetConfigName("config")   // name of config file (without extension)
	viper.SetConfigType("yaml")     // YAML format
	viper.AddConfigPath("configs/") // look for config in the config directory

	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("Error reading config file", err)
		return
	}

	keys := viper.AllKeys()
	for _, key := range keys {
		fmt.Println(key)
	}

	fmt.Printf("viper.Get(\"routes\"): %v\n", viper.Get("routes"))
	fmt.Printf("viper.ConfigFileUsed(): %v\n", viper.ConfigFileUsed())
	fmt.Printf("viper.AllSettings(): %v\n", viper.AllSettings())

	config := &Config{}

	err = viper.Unmarshal(config)
	if err != nil {
		panic(err)
	}

	s := config.Routes[0]
	fmt.Printf("s.Name: %v\n", s.Name)
	fmt.Printf("s.ListenPort: %v\n", s.ListenPort)

	t := s.Targets[0]
	fmt.Printf("t.Name: %v\n", t.Name)
	fmt.Printf("t.Protocol: %v\n", t.Protocol)
	fmt.Printf("t.Host: %v\n", t.Port)

}

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
