package main

// TODO: check that we're running with mount priv

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"io/ioutil"

	"github.com/rosmo/dkvolume"
)

const (
	pluginId = "snapshot"
)

var (
	socketAddress = filepath.Join("/run/docker/plugins/", strings.Join([]string{pluginId, ".sock"}, ""))
	defaultDir    = filepath.Join(dkvolume.DefaultDockerRootDirectory, strings.Join([]string{"_", pluginId}, ""))
	dmDir         = flag.String("dmDir", "/dev/", "Device mapper root directory")
	dmSize	      = flag.String("dmSize", "100%ORIGIN", "Snapshot size")
	root	      = flag.String("root", defaultDir, "Directory where temporary mounts are done")
)

type snapshotDriver struct {
	root string
}

func (g snapshotDriver) Create(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Create %v\n", r)
	return dkvolume.Response{}
}

func (g snapshotDriver) Remove(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Remove %v\n", r)
	return dkvolume.Response{}
}

func (g snapshotDriver) Path(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Path %v\n", r)
	return dkvolume.Response{Mountpoint: filepath.Join(g.root, r.Name)}
}

func (g snapshotDriver) Mount(r dkvolume.Request) dkvolume.Response {
       	v := strings.Split(r.Name, "/")
	mapperDirectory := v[0]
	source := v[1]

	if err := os.MkdirAll(g.root + "/" + r.ID + "/" + mapperDirectory, 0755); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	mountDir, err := ioutil.TempDir(g.root + "/" + r.ID + "/" + mapperDirectory, source)
        if err != nil {
		return dkvolume.Response{Err: err.Error()}
	}
	fmt.Printf("Temporary directory %s\n", mountDir)

	if err := os.MkdirAll(mountDir, 0755); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	snapshotName := filepath.Base(mountDir)
	snapshotSource := *dmDir + mapperDirectory + "/" + source
	fmt.Printf("Creating snapshot %s from %s\n", snapshotName, snapshotSource)

	if err := run("lvcreate", "-s", "-l", *dmSize, "-n", snapshotName, snapshotSource); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	fmt.Printf("Mount %s at %s\n", snapshotSource, mountDir)

	snapshotDevice := *dmDir + mapperDirectory + "/" + snapshotName
	if err := run("mount", snapshotDevice, mountDir); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	return dkvolume.Response{Mountpoint: mountDir}
}

func (g snapshotDriver) Unmount(r dkvolume.Request) dkvolume.Response {
     	fmt.Printf("Unmount request %v\n\n\n", r)

	mountDir := g.root + "/" + r.ID
	matches, err := filepath.Glob(mountDir + "/*/*")
	if err != nil {
	   	return dkvolume.Response{Err: err.Error()}
	}
	if len(matches) == 0 {
	   	return dkvolume.Response{Err: "Unable to find mounted directory"}
	}
	
	fmt.Printf("Unmounting %s\n", matches[0])
	if err := run("umount", matches[0]); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

       	v := strings.Split(matches[0], "/")
	snapshotPath := *dmDir + v[len(v)-2] + "/" + v[len(v)-1]
	fmt.Printf("Removing snapshot from %s\n", snapshotPath)

	if err := run("lvremove", "-f", snapshotPath); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	
	if err := os.Remove(matches[0]); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}		

	if err := os.Remove(filepath.Dir(matches[0])); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}		

	if err := os.Remove(mountDir); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}		
	
	return dkvolume.Response{}
}

func main() {
     	flag.Parse()
	d := snapshotDriver{*root}
	h := dkvolume.NewHandler(d)
	fmt.Printf("listening on %s\n", socketAddress)
	fmt.Printf("Device mapper directory: %s\n", *dmDir)
	fmt.Println(h.ServeUnix("root", socketAddress))
}

var (
	verbose = true
)

func run(exe string, args ...string) error {
	cmd := exec.Command(exe, args...)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Printf("executing: %v %v\n", exe, strings.Join(args, " "))
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
