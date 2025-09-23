package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nanoteck137/pyrin/spark"
	"github.com/nanoteck137/pyrin/spark/dart"
	"github.com/nanoteck137/pyrin/spark/golang"
	"github.com/nanoteck137/pyrin/spark/typescript"
	"github.com/spf13/cobra"
)

func readServerDef(p string) (*spark.ServerDef, *spark.Resolver, error) {
	d, err := os.ReadFile(p)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	var serverDef spark.ServerDef
	err = json.Unmarshal(d, &serverDef)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal server def: %w", err)
	}

	resolver, err := spark.CreateResolverFromServerDef(&serverDef)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resolver: %w", err)
	}

	return &serverDef, resolver, nil
}

func parseMapping(a []string) map[string]string {
	res := make(map[string]string)

	for _, i := range a {
		splits := strings.Split(i, "=")
		if len(splits) == 2 {
			left := splits[0]
			right := splits[1]

			res[left] = right
		}
	}

	return res
}

var genCmd = &cobra.Command{
	Use: "gen",
}

var genTsCmd = &cobra.Command{
	Use:   "typescript <SERVER_DEF_FILE>",
	Short: "Generate Typescript code",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		output, _ := cmd.Flags().GetString("output")

		mappings, _ := cmd.Flags().GetStringArray("map")
		nameMapping := parseMapping(mappings)

		serverDef, resolver, err := readServerDef(input)
		if err != nil {
			logger.Fatal("failed to retrive server def", "err", err)
		}

		gen := typescript.TypescriptGenerator{
			NameMapping: nameMapping,
		}

		err = gen.Generate(serverDef, resolver, output)
		if err != nil {
			logger.Fatal("failed to generate typescript", "err", err)
		}
	},
}

var genGoCmd = &cobra.Command{
	Use:   "go <SERVER_DEF_FILE>",
	Short: "Generate Golang code",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		output, _ := cmd.Flags().GetString("output")

		mappings, _ := cmd.Flags().GetStringArray("map")
		nameMapping := parseMapping(mappings)

		serverDef, resolver, err := readServerDef(input)
		if err != nil {
			logger.Fatal("failed to retrive server def", "err", err)
		}

		gen := golang.GolangGenerator{
			NameMapping: nameMapping,
		}

		err = gen.Generate(serverDef, resolver, output)
		if err != nil {
			logger.Fatal("failed to generate golang", "err", err)
		}
	},
}

var genDartCmd = &cobra.Command{
	Use:   "dart <SERVER_DEF_FILE>",
	Short: "Generate Dart code",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		output, _ := cmd.Flags().GetString("output")

		mappings, _ := cmd.Flags().GetStringArray("map")
		nameMapping := parseMapping(mappings)

		serverDef, resolver, err := readServerDef(input)
		if err != nil {
			logger.Fatal("failed to retrive server def", "err", err)
		}

		gen := dart.DartGenerator{
			NameMapping: nameMapping,
		}

		err = gen.Generate(serverDef, resolver, output)
		if err != nil {
			logger.Fatal("failed to generate dart", "err", err)
		}
	},
}

func init() {
	genTsCmd.Flags().StringArrayP("map", "m", []string{}, "Map name")
	genTsCmd.Flags().StringP("output", "o", "./src/api", "Output directory")

	genGoCmd.Flags().StringArrayP("map", "m", []string{}, "Map name")
	genGoCmd.Flags().StringP("output", "o", "./api", "Output directory")

	genDartCmd.Flags().StringArrayP("map", "m", []string{}, "Map name")
	genDartCmd.Flags().StringP("output", "o", "./lib/api", "Output directory")

	genCmd.AddCommand(genTsCmd)
	genCmd.AddCommand(genGoCmd)
	genCmd.AddCommand(genDartCmd)

	rootCmd.AddCommand(genCmd)
}
