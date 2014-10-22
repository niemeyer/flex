package flex

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/lxc/go-lxc.v2"
	"gopkg.in/tomb.v2"
)

const lxcpath = "/usr/local/var/lib/lxc"

// A Daemon can respond to requests from a flex client.
type Daemon struct {
	tomb   tomb.Tomb
	config Config
	l      net.Listener
	mux    *http.ServeMux
}

// varPath returns the provided path elements joined by a slash and
// appended to the end of $FLEX_DIR, which defaults to /var/lib/flex.
func varPath(path ...string) string {
	varDir := os.Getenv("FLEX_DIR")
	if varDir == "" {
		return "/var/lib/flex"
	}
	items := []string{varDir}
	items = append(items, path...)
	return filepath.Join(items...)
}

// StartDaemon starts the flex daemon with the provided configuration.
func StartDaemon(config *Config) (*Daemon, error) {
	d := &Daemon{config: *config}
	d.mux = http.NewServeMux()
	d.mux.HandleFunc("/ping", d.servePing)
	d.mux.HandleFunc("/list", d.serveList)
	d.mux.HandleFunc("/create", d.serveCreate)

	d.mux.HandleFunc("/start", buildByNameServe("start", func(c *lxc.Container) error { return c.Start() }))
	d.mux.HandleFunc("/stop", buildByNameServe("stop", func(c *lxc.Container) error { return c.Stop() }))
	d.mux.HandleFunc("/reboot", buildByNameServe("reboot", func(c *lxc.Container) error { return c.Reboot() }))
	d.mux.HandleFunc("/destroy", buildByNameServe("destroy", func(c *lxc.Container) error { return c.Destroy() }))

	addr, err := net.ResolveUnixAddr("unix", varPath("unix.socket"))
	if err != nil {
		return nil, fmt.Errorf("cannot resolve unix socket address: %v", err)
	}
	l, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot listen on unix socket: %v", err)
	}
	d.l = l
	d.tomb.Go(func() error { return http.Serve(d.l, d.mux) })
	return d, nil
}

var errStop = fmt.Errorf("requested stop")

// Stop stops the flex daemon.
func (d *Daemon) Stop() error {
	d.tomb.Kill(errStop)
	d.l.Close()
	err := d.tomb.Wait()
	if err == errStop {
		return nil
	}
	return err
}

func (d *Daemon) servePing(w http.ResponseWriter, r *http.Request) {
	Debugf("responding to ping")
	w.Write([]byte("pong"))
}

func (d *Daemon) serveList(w http.ResponseWriter, r *http.Request) {
	Debugf("responding to list")
	c := lxc.DefinedContainers(lxcpath)
	for i := range c {
		fmt.Fprintf(w, "%d: %s (%s)\n", i, c[i].Name(), c[i].State())
	}

}

func (d *Daemon) serveCreate(w http.ResponseWriter, r *http.Request) {
	Debugf("responding to create")

	name := r.FormValue("name")
	if name == "" {
		fmt.Fprintf(w, "failed parsing name")
		return
	}

	distro := r.FormValue("distro")
	if distro == "" {
		fmt.Fprintf(w, "failed parsing distro")
		return
	}

	release := r.FormValue("release")
	if release == "" {
		fmt.Fprintf(w, "failed parsing release")
		return
	}

	arch := r.FormValue("arch")
	if arch == "" {
		fmt.Fprintf(w, "failed parsing arch")
		return
	}

	opts := lxc.TemplateOptions{
		Template: "download",
		Distro:   distro,
		Release:  release,
		Arch:     arch,
	}

	c, err := lxc.NewContainer(name, lxcpath)
	if err != nil {
		return
	}

	err = c.Create(opts)
	if err != nil {
		fmt.Fprintf(w, "success!")
	} else {
		fmt.Fprintf(w, "fail!")
	}
}

type byname func(*lxc.Container) error

func buildByNameServe(function string, f byname) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		Debugf(fmt.Sprintf("responding to %s", function))

		name := r.FormValue("name")
		if name == "" {
			fmt.Fprintf(w, "failed parsing name")
			return
		}

		c, err := lxc.NewContainer(name, lxcpath)
		if err != nil {
			fmt.Fprintf(w, "failed getting container")
			return
		}

		err = f(c)
		if err != nil {
			fmt.Fprintf(w, "operation failed")
			return
		}
	}
}
