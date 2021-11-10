package util

import (
	"context"
	"go.elastic.co/apm"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func WaitForShutdown() {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	_, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log.Println("Shutting down")
	apm.DefaultTracer.Flush(make(chan struct{}))
	os.Exit(0)
}

func SplitUniqueNonEmptyEntries(str string, delim rune) []string {
	items := []string{}
	tokens := strings.FieldsFunc(str, func(r rune) bool {
		return r == delim
	})
	for _, t := range tokens {
		items = appendIfMissing(items, t)
	}
	return items
}
func appendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if strings.EqualFold(ele, i) {
			return slice
		}
	}
	return append(slice, i)
}
