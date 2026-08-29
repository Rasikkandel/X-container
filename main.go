package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("unknown command: " + os.Args[1])
	}
}

func run() {
	fmt.Printf("Running %v as PID %d (host namespace)\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v as PID %d (new namespace)\n", os.Args[2:], os.Getpid())
	must(syscall.Sethostname([]byte("container-lab")))
	rootfs := "/home/rasik-kandel/rootfs" // change acc to the where the tyo bin, lib , lib64 haru xah 

	must(pivotRoot(rootfs))

	must(syscall.Mount("proc", "/proc", "proc", 0, ""))
	binary, err := exec.LookPath(os.Args[2]) 
	must(err)

	must(syscall.Exec(binary, os.Args[2:], os.Environ()))
}

func pivotRoot(newRoot string) error {
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount private: %w", err)
	}
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("bind mount rootfs: %w", err)
	}
	oldRootDir := filepath.Join(newRoot, ".old_root")
	if err := os.MkdirAll(oldRootDir, 0700); err != nil {
		return err
	}
	if err := syscall.PivotRoot(newRoot, oldRootDir); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}
	if err := syscall.Chdir("/"); err != nil {
		return err
	}
	oldRootDir = "/.old_root"
	if err := syscall.Unmount(oldRootDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount old root: %w", err)
	}
	return os.RemoveAll(oldRootDir)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}