package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "Run":
		run()
	case "child":
		child()
	default:
		panic("unknown command:" + os.Args[1])
	}
}

func run() {
	fmt.Printf("Running %v with process id %d", os.Args[2:], os.Getpid())
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("running %v with process id %d (new namspace)", os.Args[2:], os.Getpid())
	must(syscall.Sethostname([]byte("container-owner")))
	binary, err := exec.LookPath(os.Args[2])
	must(err)
	must(syscall.Exec(binary, os.Args[2:], os.Environ()))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
