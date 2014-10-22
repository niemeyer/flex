package flex_test

import (
	"os"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/niemeyer/flex"
)

// Hook up gocheck in the standard go test framework.
func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&FlexSuite{})

type FlexSuite struct {
	realHome string
	tempHome string
	confPath string
	flexDir  string

	client *flex.Client
	daemon *flex.Daemon
}

func (s *FlexSuite) SetUpTest(c *C) {
	flex.SetLogger(c)
	flex.SetDebug(true)

	s.realHome = os.Getenv("HOME")
	s.tempHome = c.MkDir()
	s.confPath = filepath.Join(s.tempHome, ".flex", "config.yaml")
	os.Setenv("HOME", s.tempHome)

	s.flexDir = c.MkDir()
	os.Setenv("FLEX_DIR", s.flexDir)

	os.Mkdir(filepath.Dir(s.confPath), 0700)

	var config flex.Config
	daemon, err := flex.StartDaemon(&config)
	c.Assert(err, IsNil)
	client, err := flex.NewClient(&config)
	c.Assert(err, IsNil)
	s.client = client
	s.daemon = daemon
}

func (s *FlexSuite) TearDownTest(c *C) {
	os.Setenv("HOME", s.realHome)
	os.Setenv("FLEX_DIR", "")
}

func (s *FlexSuite) TestPing(c *C) {
	// NewClient should have pinged already.
	c.Assert(c.GetTestLog(), Matches, "(?s).*responding to ping.*")
}
