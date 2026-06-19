package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Place your code here,
	// P.S. Do not rush to throw context down, think think if it is useful with blocking operation?

	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "timeout")
	flag.Parse()
	if flag.NArg() < 2 {
		println("Usage: go-telnet --timeout=10s host port")
		return
	}
	host := flag.Arg(0)
	port := flag.Arg(1)
	address := net.JoinHostPort(host, port)
	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)
	err := client.Connect()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer client.Close()
	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()

	errCh := make(chan error, 2)

	go func() {
		errCh <- client.Send()
	}()

	go func() {
		errCh <- client.Receive()
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if errors.Is(err, io.EOF) {
			fmt.Fprintln(os.Stderr, "...EOF")
		} else if err != nil {
			fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
		}
	}
}
