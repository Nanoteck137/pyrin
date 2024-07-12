package cmd

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/nanoteck137/pyrin/client"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
		server := client.Server{
			Types: []client.MetadataType{
				{
					Name:   "Track",
					Extend: "",
					Fields: []client.TypeField{
						{
							Name: "name",
							Type: "string",
							Omit: false,
						},
						{
							Name: "num",
							Type: "int",
							Omit: false,
						},
					},
				},
				{
					Name:   "GetTracks",
					Extend: "",
					Fields: []client.TypeField{
						{
							Name: "tracks",
							Type: "[]Track",
							Omit: false,
						},
					},
				},
				{
					Name:   "GetTrackById",
					Extend: "Track",
					Fields: []client.TypeField{
						{
							Name: "num",
							Type: "string",
							Omit: false,
						},
					},
				},
			},
			Endpoints: []client.Endpoint{
				{
					Name:         "GetTracks",
					Method:       http.MethodGet,
					Path:         "/api/v1/tracks",
					ResponseType: "GetTracks",
					BodyType:     "",
				},
				{
					Name:         "GetTrackById",
					Method:       http.MethodGet,
					Path:         "/api/v1/tracks/:id",
					ResponseType: "GetTrackById",
					BodyType:     "",
				},
			},
		}

		d, err := json.MarshalIndent(server, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		out := "work"
		err = os.MkdirAll(out, 0755)
		if err != nil {
			log.Fatal(err)
		}

		p := path.Join(out, "test.json")
		err = os.WriteFile(p, d, 0644)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
