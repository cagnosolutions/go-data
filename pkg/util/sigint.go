package util

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func ShutdownHook(fn func()) {
	log.Println("Please press ctrl+c to exit.")
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	go func() {
		sig := <-c
		log.Printf("Received signal: %q (%d)\n", sig, sig)
		if fn != nil {
			fn()
		}
		os.Exit(1)
	}()
}
