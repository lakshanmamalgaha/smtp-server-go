package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/mail"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mhale/smtpd"
)

func mailHandler(origin net.Addr, from string, to []string, data []byte) error {
	_, err := mail.ReadMessage(bytes.NewReader(data))
	log.Print("Email received from " + from)
	if err != nil {
		log.Print(err)
		log.Printf("\ndata: \"%s\"", data)
	}
	return nil
}

func authHandler(remoteAddr net.Addr, mechanism string, username []byte, password []byte, shared []byte) (bool, error) {
	return string(username) == "username" && string(password) == "password", nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	srv := &smtpd.Server{
		Addr:         "0.0.0.0:8025",
		Handler:      mailHandler,
		Hostname:     "",
		AuthHandler:  authHandler,
		AuthMechs:    map[string]bool{"LOGIN": true, "PLAIN": true, "CRAM-MD5": true},
		TLSRequired:  false,
		AuthRequired: true}

	go func() {
		fmt.Println("Server started on address 0.0.0.0:8025")
		err := srv.ListenAndServe()
		log.Print(err)
	}()
	c := make(chan os.Signal, 1)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C), SIGKILL, SIGQUIT or SIGTERM (Ctrl+/)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	// Block until we receive our sig
	sig := <-c
	log.Print("Server stopped ", fmt.Sprintf("Received Signal: %s", sig))

	// Start destructing the process
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Print(err)
	}

}
