package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var printCmd = &cobra.Command{
	Use: "print <SERVER_DEF_FILE>",
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		_, resolver, err := readServerDef(input)
		if err != nil {
			logger.Fatal("failed to retrive server def", "err", err)
		}

		fmt.Println("Symbols")
		for _, sym := range resolver.ResolvedSymbols {
			fmt.Printf("%s\n", sym.Name)
			for _, field := range sym.Decl.Fields {
				fullName := sym.Name + "." + field.Name

				fmt.Printf(" - %s : %s\n", fullName, field.Name)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(printCmd)
}
