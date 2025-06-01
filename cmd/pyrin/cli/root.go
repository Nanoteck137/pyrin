package cli

import (
	"fmt"
	"log"

	"github.com/nanoteck137/pyrin/trail"
	"github.com/spf13/cobra"
)

var logger = trail.NewLogger(&trail.Options{
	Debug: true,
})

var AppName = "pyrin"
var Version = "no-version"
var Commit = "no-commit"

var rootCmd = &cobra.Command{
	Use:     AppName,
	Version: Version,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func versionTemplate() string {
	return fmt.Sprintf(
		"%s: %s (%s)\n",
		AppName, Version, Commit)
}

func init() {
	rootCmd.SetVersionTemplate(versionTemplate())
}
