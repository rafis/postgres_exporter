package target

import (
	"os"
	"path/filepath"
	"time"
)

// Path reports if any of the sources have been modified more recently
// than the destination.  Path does not descend into directories, it literally
// just checks the modtime of each thing you pass to it.
func Path(dst string, sources ...string) (bool, error) {
	stat, err := os.Stat(dst)
	if err != nil {
		return false, err
	}
	srcTime := stat.ModTime()
	dt, err := loadTargets(sources)
	if err != nil {
		return false, err
	}
	t := dt.modTime()
	if t.After(srcTime) {
		return true, nil
	}
	return false, nil
}

// Dir reports whether any of the sources have been modified
// more recently than the destination.  If a source or destination is
// a directory, modtimes of files under those directories are compared
// instead.
func Dir(dst string, sources ...string) (bool, error) {
	stat, err := os.Stat(dst)
	if err != nil {
		return false, err
	}
	srcTime := stat.ModTime()
	if stat.IsDir() {
		srcTime, err = calDirModTimeRecursive(stat)
		if err != nil {
			return false, err
		}
	}
	dt, err := loadTargets(sources)
	if err != nil {
		return false, err
	}
	t, err := dt.modTimeDir()
	if err != nil {
		return false, err
	}
	if t.After(srcTime) {
		return true, nil
	}
	return false, nil
}

func calDirModTimeRecursive(dir os.FileInfo) (time.Time, error) {
	t := dir.ModTime()
	ferr := filepath.Walk(dir.Name(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.ModTime().After(t) {
			t = info.ModTime()
		}
		return nil
	})
	if ferr != nil {
		return time.Time{}, ferr
	}
	return t, nil
}

type depTargets struct {
	src    []os.FileInfo
	hasdir bool
	latest time.Time
}

func loadTargets(targets []string) (*depTargets, error) {
	d := &depTargets{}
	for _, v := range targets {
		stat, err := os.Stat(v)
		if err != nil {
			return nil, err
		}
		if stat.IsDir() {
			d.hasdir = true
		}
		d.src = append(d.src, stat)
		if stat.ModTime().After(d.latest) {
			d.latest = stat.ModTime()
		}
	}
	return d, nil
}

func (d *depTargets) modTime() time.Time {
	return d.latest
}

func (d *depTargets) modTimeDir() (time.Time, error) {
	if !d.hasdir {
		return d.latest, nil
	}
	var err error
	for _, i := range d.src {
		t := i.ModTime()
		if i.IsDir() {
			t, err = calDirModTimeRecursive(i)
			if err != nil {
				return time.Time{}, err
			}
		}
		if t.After(d.latest) {
			d.latest = t
		}
	}
	return d.latest, nil
}
