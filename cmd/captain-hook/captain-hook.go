package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/jenkins-infra/captain-hook/pkg/cmd"
	"github.com/jenkins-infra/captain-hook/pkg/version"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Version is dynamically set by the toolchain or overridden by the Makefile.
var Version = version.Version

var Verbose bool

// BuildDate is dynamically set at build time in the Makefile.
var BuildDate = version.BuildDate

var versionOutput = ""

func init() {
	// log in json format
	logrus.SetFormatter(&logrus.JSONFormatter{})

	if strings.Contains(Version, "dev") {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
	Version = strings.TrimPrefix(Version, "v")
	if BuildDate == "" {
		RootCmd.Version = Version
	} else {
		RootCmd.Version = fmt.Sprintf("%s (%s)", Version, BuildDate)
	}
	versionOutput = fmt.Sprintf("captain-hook version %s", RootCmd.Version)
	RootCmd.AddCommand(versionCmd)
	RootCmd.SetVersionTemplate(versionOutput)

	RootCmd.PersistentFlags().Bool("help", false, "Show help for command")
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "debug", "v", false, "Debug Output")

	RootCmd.Flags().Bool("version", false, "Show version")

	RootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if err == pflag.ErrHelp {
			return err
		}
		return &FlagError{Err: err}
	})

	RootCmd.AddCommand(cmd.NewListenCmd())

	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if Verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}
	}
}

// FlagError is the kind of error raised in flag processing.
type FlagError struct {
	Err error
}

// Error.
func (fe FlagError) Error() string {
	return fe.Err.Error()
}

// Unwrap FlagError.
func (fe FlagError) Unwrap() error {
	return fe.Err
}

// RootCmd is the entry point of command-line execution.
var RootCmd = &cobra.Command{
	Use:   "captain-hook",
	Short: "Store and Forward Webhooks",
	Long:  `a HA store & forward webhook handler for Jenkins webhook events.`,

	SilenceErrors: false,
	SilenceUsage:  false,
}

var versionCmd = &cobra.Command{
	Use:    "version",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(versionOutput)
	},
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
