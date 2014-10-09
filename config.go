package flex

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config holds settings to be used by a client or daemon.
type Config struct {
	// TestOption is used only for testing purposes.
	TestOption string `yaml:"test-option,omitempty"`
}

var configPath = "$HOME/.flex/config.yaml"

// LoadConfig reads the configuration from $HOME/.flex/config.yaml.
func LoadConfig() (*Config, error) {
	data, err := ioutil.ReadFile(os.ExpandEnv(configPath))
	if os.IsNotExist(err) {
		// A missing file is equivalent to the default configuration.
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}

	var c Config
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("cannot parse configuration: %v", err)
	}
	return &c, nil
}

// SaveConfig writes the provided configuration to $HOME/.flex/config.yaml.
func SaveConfig(c *Config) error {
	fname := os.ExpandEnv(configPath)

	// Ignore errors on these two calls. Create will report any problems.
	os.Remove(fname + ".new")
	os.Mkdir(filepath.Dir(fname), 0700)
	f, err := os.Create(fname + ".new")
	if err != nil {
		return fmt.Errorf("cannot create config file: %v", err)
	}

	// If there are any errors, do not leave it around.
	defer f.Close()
	defer os.Remove(fname + ".new")

	data, err := yaml.Marshal(c)
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("cannot write configuration: %v", err)
	}

	f.Close()
	err = os.Rename(fname + ".new", fname)
	if err != nil {
		return fmt.Errorf("cannot rename temporary config file: %v", err)
	}
	return nil
}
