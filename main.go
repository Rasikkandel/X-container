package main 

import (
	"fmt" 
	"os" 
	"os/exec" 
	"syscall"
)  


func main() {
	switch os.Args[1] {
	case "Run" : 
		run() 
	default : 
		panic("unknown command:"+os.Args[1])
	} 
} 

func run() {
	fmt.Printf("Running %v with process id %d",os.Args[2:],os.Getpid()) 
	cmd := exec.Command(os.Args[2],os.Args[3:]...) 
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	} 
	must(cmd.Run()) 
}

func must(err error) {
	if err != nil {
		panic(err)
	} 
} 