package main

import (
	"os"
	"time"
)

func fileLog(file *os.File, message string) {
	_, err = file.Write([]byte(time.Now().Format(time.RFC3339Nano) + message + "\n"))
	if err != nil {
		// Exit the program
		// As fileLog.Fatal might not be accessible we directly exit the program here with a single-use exit code
		//TODO document exit codes
		os.Exit(500)
	}
}

func openLogFile() *os.File {
	logFile, err := os.OpenFile(time.Now().Format(time.RFC3339)+"_out.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		os.Exit(499)
	}
	return logFile
}
