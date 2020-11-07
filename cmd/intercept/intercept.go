package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/projectdiscovery/collaborator/biid"
	"github.com/projectdiscovery/gologger"
)

type Options struct {
	InterceptBIIDTimeout int
}

func main() {
	var options Options
	flag.IntVar(&options.InterceptBIIDTimeout, "intercept-biid-timeout", 600, "Automatic BIID intercept Timeout")

	// Setup close handler
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Println("\r- Ctrl+C pressed in Terminal")
			os.Exit(0)
		}()
	}()

	if os.Getuid() != 0 {
		gologger.Printf("Command may fail as the program is not running as root and unable to access raw sockets")
	}
	gologger.Printf("Attempting to intercept BIID")
	// otherwise attempt to retrieve it
	interceptedBiid, err := biid.Intercept(time.Duration(options.InterceptBIIDTimeout) * time.Second)
	if err != nil {
		gologger.Fatalf("%s", err)
	}
	if interceptedBiid == "" {
		gologger.Fatalf("BIID not found")
	}
	gologger.Printf("BIID found: %s", interceptedBiid)
}
