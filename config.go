package flex

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

// Config holds settings to be used by a client or daemon.
type Config struct {
	// TestOption is used only for testing purposes.
	TestOption string `yaml:"test-option,omitempty"`
}

var configPath = "$HOME/.flex/config.yaml"

// ReadConfig reads settings from the provided reader.
// If r is nil, the default configuration file is used.
func ReadConfig(r io.Reader) (*Config, error) {
	if r == nil {
		f, err := os.Open(os.ExpandEnv(configPath))
		if os.IsNotExist(err) {
			// A missing file is equivalent to the default configuration.
			return &Config{}, nil
		}
		if err != nil {
			return nil, fmt.Errorf("cannot open config file: %v", err)
		}
		defer f.Close()

		r = f
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read configuration: %v", err)
	}
	var c Config
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("cannot parse configuration: %v", err)
	}
	return &c, nil
}

// WriteConfig writes settings to the provider writer.
// If w is nil, the default configuration file is used.
func WriteConfig(c *Config, w io.Writer) error {
	var wname string
	var wfile *os.File
	if w == nil {
		wname = os.ExpandEnv(configPath)
		// Ignore errors on these two calls. Create will report any problems.
		os.Remove(wname + ".new")
		os.Mkdir(filepath.Dir(wname), 0700)
		f, err := os.Create(wname + ".new")
		if err != nil {
			return fmt.Errorf("cannot create config file: %v", err)
		}

		// If there are any errors, do not leave it around.
		defer f.Close()
		defer os.Remove(wname + ".new")

		wfile = f
		w = f
	}

	data, err := yaml.Marshal(c)
	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("cannot write configuration: %v", err)
	}

	if wfile != nil {
		wfile.Close()
		err := os.Rename(wname + ".new", wname)
		if err != nil {
			return fmt.Errorf("cannot rename temporary config file: %v", err)
		}
	}
	return nil
}
