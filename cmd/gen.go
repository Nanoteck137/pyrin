package cmd

import (
	"log"
	"os"
	"os/exec"

	"github.com/nanoteck137/pyrin/gen/gogen"
	"github.com/nanoteck137/pyrin/gen/jsgen"
	"github.com/nanoteck137/pyrin/parser"
	"github.com/nanoteck137/pyrin/resolve"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use: "gen",
}

var genGoCmd = &cobra.Command{
	Use:  "go <PYRIN_FILE>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		pkg, _ := cmd.Flags().GetString("package")
		output, _ := cmd.Flags().GetString("output")
		formatCode, _ := cmd.Flags().GetBool("format")

		file, err := os.Open(input)
		if err != nil {
			log.Fatal(err)
		}

		parser := parser.New(file)
		resolver := resolve.New()

		decls := parser.Parse()

		for _, decl := range decls {
			resolver.AddSymbolDecl(decl)
		}

		resolver.ResolveAll()

		generator := gogen.New(gogen.Options{
			PackageName: pkg,
			Output:      output,
		})

		generator.Generate(resolver)

		if formatCode {
			cmd := exec.Command("gofmt", "-w", output)
			err := cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

var genJsCmd = &cobra.Command{
	Use:  "js <PYRIN_FILE>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		output, _ := cmd.Flags().GetString("output")

		file, err := os.Open(input)
		if err != nil {
			log.Fatal(err)
		}

		parser := parser.New(file)
		resolver := resolve.New()

		decls := parser.Parse()

		for _, decl := range decls {
			resolver.AddSymbolDecl(decl)
		}

		resolver.ResolveAll()

		generator := jsgen.New(jsgen.Options{
			Output: output,
		})

		generator.Generate(resolver)
	},
}

func init() {
	genGoCmd.Flags().StringP("package", "p", "types", "Name of the package declaration")
	genGoCmd.Flags().StringP("output", "o", "./types/types.go", "Output file")
	genGoCmd.Flags().BoolP("format", "f", false, "Use 'gofmt' to format output")

	genJsCmd.Flags().StringP("output", "o", "./src/types/types.ts", "Output file")

	genCmd.AddCommand(genGoCmd)
	genCmd.AddCommand(genJsCmd)

	rootCmd.AddCommand(genCmd)
}
