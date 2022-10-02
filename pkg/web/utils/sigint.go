package utils

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func HandleSignalInterrupt(msg string, args ...interface{}) {
	log.Printf(msg, args...)
	log.Println("Please press ctrl+c to exit.")
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()
}

func HandleSignalInterruptFunc(fn func(), msg string, args ...interface{}) {
	log.Printf(msg, args...)
	log.Println("Please press ctrl+c to exit.")
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fn()
		os.Exit(1)
	}()
}
