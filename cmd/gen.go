package cmd

import (
	"log"
	"os"

	"github.com/nanoteck137/pyrin/gen/zod"
	"github.com/nanoteck137/pyrin/parser"
	"github.com/nanoteck137/pyrin/resolve"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use: "gen",
}

var genZodCmd = &cobra.Command{
	Use:  "zod <GO_FILE>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		output, _ := cmd.Flags().GetString("output")

		file, err := os.ReadFile(input)
		if err != nil {
			log.Fatal(err)
		}

		resolver := resolve.New()
		decls := parser.Parse(string(file))

		for _, decl := range decls {
			resolver.AddSymbolDecl(decl)
		}

		err = resolver.ResolveAll()
		if err != nil {
			log.Fatal(err)
		}

		generator := zod.New(zod.Options{
			Output: output,
		})

		generator.Generate(resolver)
	},
}

func init() {
	genZodCmd.Flags().StringP("output", "o", "./src/types/types.ts", "Output file")
	genCmd.AddCommand(genZodCmd)

	rootCmd.AddCommand(genCmd)
}
