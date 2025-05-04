package cli

import (
	"log"
	"os"

	"github.com/nanoteck137/pyrin/tools/gen"
	"github.com/spf13/cobra"
)

var genTsCmd = &cobra.Command{
	Use:   "ts <SERVER_SPEC>",
	Short: "Generate Typescript code",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		output, _ := cmd.Flags().GetString("output")

		err := os.MkdirAll(output, 0755)
		if err != nil {
			log.Fatal(err)
		}

		spec, err := gen.ReadSpec(input)
		if err != nil {
			log.Fatal(err)
		}

		err = gen.GenerateTypescript(spec, output)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var genGoCmd = &cobra.Command{
	Use:   "go <SERVER_SPEC>",
	Short: "Generate Golang code",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		output, _ := cmd.Flags().GetString("output")

		err := os.MkdirAll(output, 0755)
		if err != nil {
			log.Fatal(err)
		}

		spec, err := gen.ReadSpec(input)
		if err != nil {
			log.Fatal(err)
		}

		err = gen.GenerateGolang(spec, output)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var genDartCmd = &cobra.Command{
	Use:   "dart <SERVER_SPEC>",
	Short: "Generate Dart code",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		output, _ := cmd.Flags().GetString("output")

		err := os.MkdirAll(output, 0755)
		if err != nil {
			log.Fatal(err)
		}

		spec, err := gen.ReadSpec(input)
		if err != nil {
			log.Fatal(err)
		}

		err = gen.GenerateDart(spec, output)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	genTsCmd.Flags().StringP("output", "o", "./src/api", "Output directory")
	genGoCmd.Flags().StringP("output", "o", "./api", "Output directory")
	genDartCmd.Flags().StringP("output", "o", "./lib/api", "Output directory")

	rootCmd.AddCommand(genTsCmd)
	rootCmd.AddCommand(genGoCmd)
	rootCmd.AddCommand(genDartCmd)
}
