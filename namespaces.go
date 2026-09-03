package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func run(cfg config, target []string) {
	fmt.Printf("Running %v as PID %d (host namespace)\n", target, os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, target...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "MYGODOCKER_ROOTFS="+cfg.rootfs)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	must(cmd.Start())

	cgroupPath, err := setupCgroup("mygodocker", cmd.Process.Pid, cfg.mem, cfg.cpu)
	must(err)
	fmt.Println("cgroup:", cgroupPath)

	must(cmd.Wait())
	// cleanup , if cgroup goes out of process 
	_ = os.Remove(cgroupPath)
}

func child(rootfs string, target []string) {
	fmt.Printf("Running %v as PID %d (new namespace)\n", target, os.Getpid())
	must(syscall.Sethostname([]byte("container")))

	must(pivotRoot(rootfs))
	must(syscall.Mount("proc", "/proc", "proc", 0, ""))

	binary, err := exec.LookPath(target[0])
	must(err)
	must(syscall.Exec(binary, target, os.Environ()))
}


func pivotRoot(newRoot string) error {
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return err
	} 
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return err
	}
	oldRootDir := filepath.Join(newRoot, ".old_root")
	if err := os.MkdirAll(oldRootDir, 0700); err != nil {
		return err
	} 
	if err := syscall.PivotRoot(newRoot, oldRootDir); err != nil {
		return err
	}
	if err := syscall.Chdir("/"); err != nil {
		return err
	}
	oldRootDir = "/.old_root"
	if err := syscall.Unmount(oldRootDir, syscall.MNT_DETACH); err != nil {
		return err
	}
	return os.RemoveAll(oldRootDir)
} 