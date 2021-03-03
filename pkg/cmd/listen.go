package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/garethjevans/captain-hook/pkg/hook"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ListenCmd defines the cmd.
type ListenCmd struct {
	Cmd  *cobra.Command
	Args []string

	ForwardURL string
}

// NewListenCmd defines a new cmd.
func NewListenCmd() *cobra.Command {
	c := &ListenCmd{}
	cmd := &cobra.Command{
		Use:     "listen",
		Short:   "captain-hook listen",
		Long:    ``,
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			c.Cmd = cmd
			c.Args = args
			err := c.Run()
			if err != nil {
				logrus.Errorf("unhandled error - %s", err)
				logrus.Fatal("unable to run command")
			}
		},
	}

	c.Cmd = cmd

	cmd.Flags().StringVarP(&c.ForwardURL, "forward-url", "f", "http://jenkins/",
		"URL to forward webhooks to")

	return cmd
}

// Run update help.
func (c *ListenCmd) Run() error {
	wait := 15 * time.Second
	options, err := hook.NewHook()
	if err != nil {
		return err
	}

	go func() {
		if err = options.Start(); err != nil {
			logrus.Fatal(err)
		}
	}()

	// create a new router
	r := mux.NewRouter()

	// add routes
	options.Handle(r)

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logrus.Fatal(err)
		}
	}()

	channel := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(channel, os.Interrupt)

	// Block until we receive our signal.
	<-channel

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	err = srv.Shutdown(ctx)
	if err != nil {
		return err
	}
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	logrus.Info("shutting down")

	return nil
}
