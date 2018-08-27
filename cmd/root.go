package cmd

import (
	"fmt"
	"os"

	"github.com/franzwilhelm/uio-exam-helper/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "uio-exam-helper",
	Short: "A CLI for UiO subjects",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "Debug mode (log db and debug output)")
	if rootCmd.PersistentFlags().Lookup("debug").Changed {
		db.Default.LogMode(true)
		log.SetLevel(log.DebugLevel)
	}
}
