package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func getOrElse(env, def string) string {
	value := os.Getenv(env)
	if value == "" {
		return def
	}
	return value
}

// Overlaper has the methods to overlap a directory and discard the changes
type Overlaper interface {
	Overlap() error
	Discard() error
}

func waitForClosure(timeout time.Duration) chan struct{} {
	waitTimeout := timeout
	if timeout == 0 {
		waitTimeout = time.Hour * 24
	}
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	out := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(waitTimeout):
				log.Printf("Timeout reached\n")
				if timeout > 0 {
					out <- struct{}{}
				}
			case s := <-signalChan:
				log.Printf("Received signal: %s\n", s)
				out <- struct{}{}
			}
		}
	}()
	return out
}

func main() {
	var timeout string
	var path string
	var leave bool

	flag.StringVar(&timeout, "t", getOrElse("OVERLAP_DURATION", "0s"), "Time to overlay (0 = infinity)")
	flag.BoolVar(&leave, "l", false, "Leave the directory overlapped and exit")
	flag.Parse()

	if args := flag.Args(); len(args) != 1 {
		log.Fatalf("Only one argument must be provided")
	} else {
		path = args[0]
	}

	durationTimeout, err := time.ParseDuration(timeout)
	if err != nil {
		log.Fatalf("Duration provided isn't valid: %s\n", err)
	}

	o := OverlayFS{Directory: path}
	if err = o.Overlap(); err != nil {
		log.Fatalf("Error overlapping directory: %s\n", err)
	}
	log.Printf("Directory %s overlapped correctly\n", path)

	if !leave {
		out := waitForClosure(durationTimeout)
		for {
			select {
			case <-out:
				if err := o.Discard(); err != nil {
					log.Printf("Error discarding directory: %s\n", err)
					go func() {
						retryTime := time.Second * 5
						log.Printf("Retrying in %s\n", retryTime)
						time.Sleep(retryTime)
						out <- struct{}{}
					}()
				} else {
					log.Printf("Changes to directory %s discarded\n", path)
					return
				}
			}
		}
	}
}
