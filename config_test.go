package flex_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"os"

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
	cfg, err := flex.ReadConfig(nil)
	c.Assert(err, IsNil)
	c.Assert(cfg.TestOption, Equals, "")
}

func (s *ConfigSuite) TestReadConfigDefault(c *C) {
	err := ioutil.WriteFile(s.confPath, []byte("test-option: value"), 0644)
	c.Assert(err, IsNil)
	cfg, err := flex.ReadConfig(nil)
	c.Assert(err, IsNil)
	c.Assert(cfg.TestOption, Equals, "value")
}

func (s *ConfigSuite) TestReadConfigReader(c *C) {
	cfg, err := flex.ReadConfig(bytes.NewBufferString("test-option: value"))
	c.Assert(err, IsNil)
	c.Assert(cfg.TestOption, Equals, "value")
}

func (s *ConfigSuite) TestWriteConfigDefault(c *C) {
	err := flex.WriteConfig(&flex.Config{TestOption: "value"}, nil)
	c.Assert(err, IsNil)
	data, err := ioutil.ReadFile(s.confPath)
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "test-option: value\n")
}

func (s *ConfigSuite) TestWriteConfigWriter(c *C) {
	var buf bytes.Buffer
	err := flex.WriteConfig(&flex.Config{TestOption: "value"}, &buf)
	c.Assert(err, IsNil)
	c.Assert(buf.String(), Equals, "test-option: value\n")

	_, err = os.Stat(s.confPath)
	if !os.IsNotExist(err) {
		c.Fatalf("config file should not exist, got stat err: %v", err)
	}
}
