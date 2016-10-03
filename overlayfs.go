package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
)

type OverlayFS struct {
	Directory string
	upperdir  string
	workdir   string
}

func (o *OverlayFS) Overlap() error {
	file, err := os.Open(o.Directory)
	if err != nil {
		return fmt.Errorf("Can't open directory: %s", err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("Can't open directory: %s", err)
	}
	stats := fileInfo.Sys().(*syscall.Stat_t)

	o.upperdir, _ = ioutil.TempDir("", "shadowud")
	o.workdir, _ = ioutil.TempDir("", "shadowwd")

	if err := syscall.Mount("overlay", o.Directory, "overlay", uintptr(0), fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", o.Directory, o.upperdir, o.workdir)); err != nil {
		return fmt.Errorf("Can't mount directory: %s", err)
	}
	os.Chown(o.Directory, int(stats.Uid), int(stats.Gid))

	return nil
}

func (o *OverlayFS) Discard() error {
	defer os.RemoveAll(o.upperdir)
	defer os.RemoveAll(o.workdir)
	return syscall.Unmount(o.Directory, 0)
}
