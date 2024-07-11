package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/kr/pretty"
	"github.com/spf13/cobra"
)

type Endpoint struct {
	Name     string
	Method   string
	Endpoint string
	Data     string
	Body     string
}

type ServerSetup struct {
	Endpoints []Endpoint
}

type CodeWriter struct {
	writer io.Writer
	indent int
}

func (w *CodeWriter) Indent() {
	w.indent += 1
}

func (w *CodeWriter) Unindent() {
	if w.indent == 0 {
		return
	}

	w.indent -= 1
}

func (w *CodeWriter) WriteIndent() error {
	for i := 0; i < w.indent; i++ {
		_, err := w.writer.Write(([]byte)("  "))
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *CodeWriter) Writef(format string, a ...any) error {
	_, err := fmt.Fprintf(w.writer, format, a...)
	return err
}

func (w *CodeWriter) IWritef(format string, a ...any) error {
	err := w.WriteIndent()
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w.writer, format, a...)

	return err
}

func (w *CodeWriter) GenerateCodeForEndpoint(e *Endpoint) error {
	// getPlaylists() {
	//   return this.request("/api/v1/playlists", "GET", api.GetPlaylists);
	// }

	// return this.request("/api/v1/playlists", "POST", api.PostPlaylist, body);

	var args []string

	parts := strings.Split(e.Endpoint, "/")
	pretty.Println(parts)

	endpointHasArgs := false

	for i, p := range parts {
		if len(p) == 0 {
			continue
		}

		if p[0] == ':' {
			name := p[1:]
			args = append(args, name)

			parts[i] = fmt.Sprintf("${%s}", name)

			endpointHasArgs = true
		}
	}

	pretty.Println(parts)

	newEndpoint := strings.Join(parts, "/")

	w.IWritef("%s(", strcase.ToLowerCamel(e.Name))

	for i, arg := range args {
		if i == 0 {
			w.Writef("%s: string", arg)
		} else {
			w.Writef(", %s: string", arg)
		}
	}

	if e.Body != "" {
		if len(args) > 0 {
			w.Writef(", ")
		}

		w.Writef("body: api.%s", e.Body)
	}

	w.Writef(") {\n")

	w.Indent()
	w.IWritef("return this.request(")

	if endpointHasArgs {
		w.Writef("`%s`", newEndpoint)
	} else {
		w.Writef("\"%s\"", newEndpoint)
	}

	w.Writef(", \"%s\"", e.Method)

	if e.Data != "" {
		w.Writef(", api.%s", e.Data)
	} else {
		w.Writef(", z.undefined()")
	}

	if e.Body != "" {
		w.Writef(", body")
	}

	w.Writef(")\n")
	// \"%s\", \"%s\", api.%s)\n", newEndpoint, e.Method, e.Data)
	w.Unindent()

	w.IWritef("}\n")

	return nil
}

var testCmd = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
		server := ServerSetup{
			Endpoints: []Endpoint{
				{
					Name:     "GetTracks",
					Method:   http.MethodGet,
					Endpoint: "/api/v1/tracks",
					Data:     "GetTracks",
					Body:     "",
				},
				{
					Name:     "GetTrackById",
					Method:   http.MethodGet,
					Endpoint: "/api/v1/tracks/:id",
					Data:     "GetTrackById",
					Body:     "",
				},
				{
					Name:     "Signin",
					Method:   http.MethodPost,
					Endpoint: "/api/v1/auth/signin",
					Data:     "PostAuthSignin",
					Body:     "PostAuthSigninBody",
				},
				{
					Name:     "addItemsToPlaylist",
					Method:   http.MethodPost,
					Endpoint: "/api/v1/playlists/:playlistId/items",
					Data:     "",
					Body:     "PostPlaylistItemsByIdBody",
				},
			},
		}

		pretty.Println(server)

		buf := &bytes.Buffer{}
		w := CodeWriter{
			writer: buf,
		}


		// export class ApiClient extends BaseApiClient {
		//   constructor(baseUrl: string) {
		//     super(baseUrl);
		//   }
		// }

		w.IWritef("import { z } from \"zod\";\n")
		w.IWritef("import * as api from \"./types\";\n")
		w.IWritef("import BaseApiClient from \"./base-client\";\n")
		w.IWritef("\n")

		w.IWritef("export class ApiClient extends BaseApiClient {\n")
		w.Indent()

		w.IWritef("constructor(baseUrl: string) {\n")
		w.Indent()
		w.IWritef("super(baseUrl);\n")
		w.Unindent()
		w.IWritef("}\n")

		for _, endpoint := range server.Endpoints {
			w.IWritef("\n")

			err := w.GenerateCodeForEndpoint(&endpoint)
			if err != nil {
				log.Fatal(err)
			}
		}

		w.Unindent()
		w.IWritef("}\n")

		fmt.Printf("%v\n", buf.String())
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
