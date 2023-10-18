package cmd

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var cfgFile string
var cfg Config

type Config struct {
	NucleiConfig NucleiConfig `yaml:"nuclei_config,omitempty"`
	LogLevel     log.Level    `yaml:"log_level,omitempty"`
	Target       string       `yaml:"target,omitempty"`
	TempDir      string       `yaml:"temp_dir,omitempty"`
}

type NucleiConfig struct {
	NucleiPath    string   `yaml:"path,omitempty"`
	NucleiConfig  string   `yaml:"config,omitempty"`
	NucleiOptions []string `yaml:"options,omitempty"`
}

func initConfig() {
	log.SetLevel(log.InfoLevel)
	file, err := os.ReadFile(cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.UnmarshalStrict(file, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.LogLevel == 0 {
		cfg.LogLevel = log.InfoLevel
	}
	if cfg.TempDir == "" {
		cfg.TempDir = "./tmp-nuclei-output/"
	}
	if err := os.MkdirAll(cfg.TempDir, 0755); err != nil {
		log.Fatal(err)
	}
	log.SetLevel(cfg.LogLevel)

	if strings.HasPrefix(cfg.NucleiConfig.NucleiPath, "~/") {
		cfg.NucleiConfig.NucleiPath = strings.Replace(cfg.NucleiConfig.NucleiPath, "~", os.Getenv("HOME"), 1)
	}
	log.Debugf("Config: %+v", cfg)
}
