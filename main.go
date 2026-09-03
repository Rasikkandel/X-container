package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
) 

type config struct {
	mem    string
	cpu    float64
	rootfs string 
} 

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "run":
		cmdRun(os.Args[2:])
	case "child":
		cmdChild(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Println("usage: mygo-docker run [--mem=100m] [--cpu=0.5] [--rootfs=path] <command> [args...]")
	os.Exit(1)
}

func cmdRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	mem := fs.String("mem", "", "memory limit, e.g. 100m")
	cpu := fs.Float64("cpu", 0, "cpu limit in cores, e.g. 0.5")
	rootfs := fs.String("rootfs", defaultRootfs(), "path to container rootfs")
	fs.Parse(args)

	target := fs.Args()
	if len(target) == 0 {
		usage()
	}

	// flagset returns the pointer of the flags so, dereferencing it 
	run(config{mem: *mem, cpu: *cpu, rootfs: *rootfs}, target)
}

func defaultRootfs() string {
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		if u, err := user.Lookup(sudoUser); err == nil {
			return filepath.Join(u.HomeDir, "rootfs")
		}
	}
	return filepath.Join(os.Getenv("HOME"), "rootfs")
}

func cmdChild(args []string) {
	rootfs := os.Getenv("MYGODOCKER_ROOTFS")
	child(rootfs, args)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}