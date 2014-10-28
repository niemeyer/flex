package flex

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
)

/*
 * We'll flesh this out to be lists of ranges
 * We will want a list of available ranges (all ranges
 * which flex may use) and taken range (parts of the
 * available ranges which are already in use by containers)
 *
 * We also may want some way of deciding which containers may
 * or perhaps must not share ranges
 *
 * For now, we simply have a single range, shared by all
 * containers
 */
type Idmap struct {
	uidmin, uidrange uint
	gidmin, gidrange uint
}

func checkmap(fname string, username string) (uint, uint, error) {
	f, err := os.Open(fname)
	var min uint
	var idrange uint
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	min = 0
	idrange = 0
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ":")
		if strings.EqualFold(s[0], username) {
			bigmin, err := strconv.ParseUint(s[1], 10, 32)
			if err != nil {
				continue
			}
			bigidrange, err := strconv.ParseUint(s[2], 10, 32)
			if err != nil {
				continue
			}
			min = uint(bigmin)
			idrange = uint(bigidrange)
		}
	}
	if idrange == 0 {
		return 0, 0, fmt.Errorf("User %q has no subuids.", username)
	}
	return min, idrange, nil
}

func (m *Idmap) InitUidmap() error {
	me, err := user.Current()
	if err != nil {
		return err
	}

	umin, urange, err := checkmap("/etc/subuid", me.Username)
	if err != nil {
		return err
	}
	gmin, grange, err := checkmap("/etc/subgid", me.Username)
	if err != nil {
		return err
	}
	m = new(Idmap)
	m.uidmin = umin
	m.uidrange = urange
	m.gidmin = gmin
	m.gidrange = grange
	return nil
}
