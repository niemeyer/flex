package flex

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/lxc/go-lxc.v2"
	"gopkg.in/tomb.v2"

	"github.com/kr/pty"
)

// A Daemon can respond to requests from a flex client.
type Daemon struct {
	tomb     tomb.Tomb
	config   Config
	l        net.Listener
	idmap    *Idmap
	lxcpath  string
	mux      *http.ServeMux
}

// varPath returns the provided path elements joined by a slash and
// appended to the end of $FLEX_DIR, which defaults to /var/lib/flex.
func varPath(path ...string) string {
	varDir := os.Getenv("FLEX_DIR")
	if varDir == "" {
		varDir = "/var/lib/flex"
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
	d.mux.HandleFunc("/attach", d.serveAttach)

	m := new(Idmap)
	err := m.InitUidmap()
	if err != nil {
		return nil, err
	}

	d.mux.HandleFunc("/start", buildByNameServe("start", func(c *lxc.Container) error { return c.Start() }, d))
	d.mux.HandleFunc("/stop", buildByNameServe("stop", func(c *lxc.Container) error { return c.Stop() }, d))
	d.mux.HandleFunc("/reboot", buildByNameServe("reboot", func(c *lxc.Container) error { return c.Reboot() }, d))
	d.mux.HandleFunc("/destroy", buildByNameServe("destroy", func(c *lxc.Container) error { return c.Destroy() }, d))

	d.lxcpath = varPath("lxc")
	err = os.MkdirAll(varPath("/"), 0755)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(d.lxcpath, 0755)
	if err != nil {
		return nil, err
	}

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
	c := lxc.DefinedContainers(d.lxcpath)
	for i := range c {
		fmt.Fprintf(w, "%d: %s (%s)\n", i, c[i].Name(), c[i].State())
	}

}

func (d *Daemon) serveAttach(w http.ResponseWriter, r *http.Request) {
	Debugf("responding to attach")

	name := r.FormValue("name")
	if name == "" {
		fmt.Fprintf(w, "failed parsing name")
		return
	}

	command := r.FormValue("command")
	if command == "" {
		fmt.Fprintf(w, "failed parsing command")
		return
	}

	secret := r.FormValue("secret")
	if secret == "" {
		fmt.Fprintf(w, "failed parsing secret")
		return
	}

	var err error
	addr := ":0"
	// tcp6 doesn't seem to work with Dial("tcp", ) at the client
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		fmt.Fprintf(w, "failed listening")
		return
	}
	fmt.Fprintf(w, "Port: ", l.Addr().String())

	go func (l net.Listener, name string, command string, secret string) {
		conn, err := l.Accept()
		l.Close()
		if err != nil {
			Debugf(err.Error())
			return
		}
		defer conn.Close()
		b := make([]byte, 100)
		n, err := conn.Read(b)
		if err != nil {
			Debugf("bad read: %s", err.Error())
			return
		}
		if n != len(secret) {
			Debugf("read %d characters, secret is %d", n, len(secret))
			return
		}
		if !strings.EqualFold(string(b), secret) {
			Debugf("strings not equal")
			// Why do they never match?  TODO fix
			// return
			// FIXME(niemeyer): It does not match because the
			// provided buffer has length 100, so string(b) will
			// also have lenght 100. It should be string(b[:n]) instead.
			// Why is casing being folded here? Shouldn't that be just
			//     string(b[:n]) == secret
			// ?
		}
		Debugf("Attaching")

		c, err := lxc.NewContainer(name, d.lxcpath)
		if err != nil {
			Debugf("%s", err.Error())
		}

		pty, tty, err := pty.Open()

		if err != nil {
			Debugf("Failed opening getting a tty: %q", err.Error())
			return
		}

		defer pty.Close()
		defer tty.Close()

		/*
		 * The pty will be passed to the container's Attach.  The two
		 * below goroutines will copy output from the socket to the
		 * pty.stdin, and from pty.std{out,err} to the socket
		 * If the RunCommand exits, we want ourselves (the gofunc) and
		 * the copy-goroutines to exit.  If the connection closes, we
		 * also want to exit
		 */
		// FIXME(niemeyer): tomb is not doing anything useful in this case.
		// It cannot externally kill the goroutines without them collaborating
		// to make that possible. Please see the blog post for details:
		// 
		// http://blog.labix.org/2011/10/09/death-of-goroutines-under-control
		var tomb tomb.Tomb
		tomb.Go(func() error {
			_, err := io.Copy(pty, conn)
			if err != nil {
				return err
			}
			return nil
		})
		tomb.Go(func() error {
			_, err := io.Copy(conn, pty)
			if err != nil {
				return err
			}
			return nil
		})

		options := lxc.DefaultAttachOptions

		options.StdinFd = tty.Fd()
		options.StdoutFd = tty.Fd()
		options.StderrFd = tty.Fd()

		options.ClearEnv = true

		_, err = c.RunCommand([]string{command}, options)
		if err != nil {
			Debugf("RunCommand error: %s", err.Error())
			return
		}

		Debugf("RunCommand exited, stopping console")
		tomb.Kill(errStop)
	} (l, name, command, secret)
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

	c, err := lxc.NewContainer(name, d.lxcpath)
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

func buildByNameServe(function string, f byname, d *Daemon) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		Debugf(fmt.Sprintf("responding to %s", function))

		name := r.FormValue("name")
		if name == "" {
			fmt.Fprintf(w, "failed parsing name")
			return
		}

		c, err := lxc.NewContainer(name, d.lxcpath)
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
