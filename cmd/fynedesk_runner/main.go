package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

const runCmd = "fynedesk"

func main() {
	for {
		exe := exec.Command(runCmd)
		exe.Env = append(os.Environ(), "FYNE_DESK_RUNNER=1")
		exe.Stdout, exe.Stderr = openLogWriter()
		err := exe.Run()
		if err == nil {
			return
		}

		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			log.Println("Could not execute", runCmd, "command")
			return
		}

		if exit, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			status := exit.ExitStatus()
			if status == 0 {
				log.Println("Exiting Error 0")
				return
			} else if status == 512 { // X server unavailable
				log.Println("X server went away")
				return
			} else {
				log.Println("Restart from status", status)
			}
		}
	}
}
