package dsnotify

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	authCfgFilename     string
	dsnotifyCfgFilename string
	readyMessage        string
	config              *DSNotifyConfig
)

func init() {
	flag.StringVar(&authCfgFilename, "auth", "./auth.yaml", "Auth config filename")
	flag.StringVar(&dsnotifyCfgFilename, "config", "./dsnotify.yaml", "Server config filename")
	flag.StringVar(&readyMessage, "ready", "Discord Stream Notify", "Ready message")
	flag.Parse()
	config = configure()
}

type DSNotifyConfig struct {
	Auth   *AuthConfig
	Guilds map[string]*GuildConfig
}

type AuthConfig struct {
	Token  string `yaml:"token"`
	Bearer string `yaml:"oauth2"`
}

type GuildConfig struct {
	Debug         bool   `yaml:"debug"`
	Enabled       bool   `yaml:"enabled"`
	NotifyChannel string `yaml:"channel"`
	NotifyRole    string `yaml:"role"`
}

func configure() *DSNotifyConfig {
	auth := &AuthConfig{}
	dsnotify := make(map[string]*GuildConfig)

	configs := []struct {
		file string
		conf interface{}
	}{
		{
			file: authCfgFilename,
			conf: auth,
		},
		{
			file: dsnotifyCfgFilename,
			conf: dsnotify,
		},
	}

	for _, c := range configs {
		bytes, err := os.ReadFile(c.file)
		if err != nil {
			log.Fatalf("[ERROR] Config read %s: %v", c.file, err)
		}

		if err := yaml.Unmarshal(bytes, c.conf); err != nil {
			log.Fatalf("[ERROR] Config unmarshal %s: %v", c.file, err)
		}
	}

	return &DSNotifyConfig{
		Auth:   auth,
		Guilds: dsnotify,
	}
}
