package flex_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/niemeyer/flex"
)

var _ = Suite(&ConfigSuite{})

type ConfigSuite struct {
	realHome string
	tempHome string
	confPath string
}

func (s *ConfigSuite) SetUpTest(c *C) {
	s.realHome = os.Getenv("HOME")
	s.tempHome = c.MkDir()
	s.confPath = filepath.Join(s.tempHome, ".flex", "config.yaml")
	os.Setenv("HOME", s.tempHome)

	os.Mkdir(filepath.Dir(s.confPath), 0700)
}

func (s *ConfigSuite) TearDownTest(c *C) {
	os.Setenv("HOME", s.realHome)
}

func (s *ConfigSuite) TestReadConfigMissing(c *C) {
	cfg, err := flex.LoadConfig()
	c.Assert(err, IsNil)
	c.Assert(cfg.TestOption, Equals, "")
}

func (s *ConfigSuite) TestLoadConfig(c *C) {
	err := ioutil.WriteFile(s.confPath, []byte("test-option: value"), 0644)
	c.Assert(err, IsNil)
	cfg, err := flex.LoadConfig()
	c.Assert(err, IsNil)
	c.Assert(cfg.TestOption, Equals, "value")
}

func (s *ConfigSuite) TestSaveConfig(c *C) {
	err := flex.SaveConfig(&flex.Config{TestOption: "value"})
	c.Assert(err, IsNil)
	data, err := ioutil.ReadFile(s.confPath)
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "test-option: value\n")
}
