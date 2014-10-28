package flex

import (
	"bufio"
	"fmt"
	"os"
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

func checkmap(fname string) (uint, uint, error) {
	f, err := os.Open(fname)
	var min uint
	var idrange uint
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Println("scannling line: %s", scanner.Text())
	}
	min = 100000
	idrange = 65536
	return min, idrange, nil
}

func (m *Idmap) InitUidmap() error {
	umin, urange, err := checkmap("/etc/subuid")
	if err != nil {
		return err
	}
	gmin, grange, err := checkmap("/etc/subgid")
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
