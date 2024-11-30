package tsg

import (
	"fmt"
	"io"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/nanoteck137/pyrin/spec"
	"github.com/nanoteck137/pyrin/tools/resolve"
	"github.com/nanoteck137/pyrin/utils"
)

func GenerateType(w io.Writer, ty resolve.Type) {
	switch t := ty.(type) {
	case *resolve.TypeString:
		fmt.Fprint(w, "z.string()")
	case *resolve.TypeInt:
		fmt.Fprint(w, "z.number()")
	case *resolve.TypeBoolean:
		fmt.Fprint(w, "z.boolean()")
	case *resolve.TypeArray:
		fmt.Fprint(w, "z.array(")
		GenerateType(w, t.ElementType)
		fmt.Fprint(w, ")")
	case *resolve.TypePtr:
		GenerateType(w, t.BaseType)
		fmt.Fprint(w, ".nullable()")
	case *resolve.TypeStruct:
		fmt.Fprint(w, t.Name)
	}
}

func GenerateField(w io.Writer, field *resolve.Field) {
	fmt.Fprint(w, "  ", field.Name, ": ")
	GenerateType(w, field.Type)
	if field.Optional {
		fmt.Fprint(w, ".optional()")
	}
	fmt.Fprintln(w, ",")
}

func GenerateTypeCode(w io.Writer, resolver *resolve.Resolver) error {
	fmt.Fprintln(w, "// DO NOT EDIT THIS: This file was generated by the Pyrin Typescript Generator")
	fmt.Fprintln(w, `import { z } from "zod";`)
	fmt.Fprintln(w)

	for _, s := range resolver.ResolvedStructs {
		switch ty := s.Type.(type) {
		case *resolve.TypeStruct:
			fmt.Fprintf(w, "export const %s = z.object({\n", s.Name)
			for _, f := range ty.Fields {
				GenerateField(w, &f)
			}
			fmt.Fprintln(w, "});")
			fmt.Fprintf(w, "export type %s = z.infer<typeof %s>;\n", s.Name, s.Name)
			fmt.Fprintln(w)
		case *resolve.TypeSameStruct:
			fmt.Fprintf(w, "export const %s = %s;\n", s.Name, ty.Type.Name)
			fmt.Fprintf(w, "export type %s = z.infer<typeof %s>;\n", s.Name, s.Name)
			fmt.Fprintln(w)
		}
	}

	return nil
}

func genNormalEndpoint(w *utils.CodeWriter, e *spec.Endpoint) error {
	var args []string
	parts := strings.Split(e.Path, "/")
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

	newEndpoint := strings.Join(parts, "/")

	w.IWritef("%s(", strcase.ToLowerCamel(e.Name))

	for _, arg := range args {
		w.Writef("%s: string, ", arg)
	}

	if e.BodyType != "" {
		w.Writef("body: api.%s, ", e.BodyType)
	}

	w.Writef("options?: ExtraOptions")

	w.Writef(") {\n")

	w.Indent()

	w.IWritef("return this.request(")

	if endpointHasArgs {
		w.Writef("`%s`", newEndpoint)
	} else {
		w.Writef("\"%s\"", newEndpoint)
	}

	w.Writef(", \"%s\"", e.Method)

	if e.ResponseType != "" {
		w.Writef(", api.%s", e.ResponseType)
	} else {
		w.Writef(", z.undefined()")
	}

	w.Writef(", z.any()")

	if e.BodyType != "" {
		w.Writef(", body")
	} else {
		w.Writef(", undefined")
	}

	w.Writef(", options")

	w.Writef(")\n")
	w.Unindent()

	w.IWritef("}\n")

	return nil
}

func genFormDataEndpoint(w *utils.CodeWriter, e *spec.Endpoint) error {
	var args []string
	parts := strings.Split(e.Path, "/")
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

	newEndpoint := strings.Join(parts, "/")

	w.IWritef("%s(", strcase.ToLowerCamel(e.Name))

	for _, arg := range args {
		w.Writef("%s: string, ", arg)
	}

	w.Writef("formData: FormData, ")
	w.Writef("options?: ExtraOptions")

	w.Writef(") {\n")

	w.Indent()

	w.IWritef("return this.requestWithFormData(")

	if endpointHasArgs {
		w.Writef("`%s`", newEndpoint)
	} else {
		w.Writef("\"%s\"", newEndpoint)
	}

	w.Writef(", \"%s\"", e.Method)

	if e.ResponseType != "" {
		w.Writef(", api.%s", e.ResponseType)
	} else {
		w.Writef(", z.undefined()")
	}

	w.Writef(", z.undefined()")

	w.Writef(", formData")

	w.Writef(", options")

	w.Writef(")\n")
	w.Unindent()

	w.IWritef("}\n")

	return nil
}

func GenerateClientCode(w io.Writer, server *spec.Server) error {
	cw := utils.CodeWriter{
		Writer:    w,
		IndentStr: "  ",
	}

	cw.IWritef("import { z } from \"zod\";\n")
	cw.IWritef("import * as api from \"./types\";\n")
	cw.IWritef("import { BaseApiClient, type ExtraOptions } from \"./base-client\";\n")
	cw.IWritef("\n")

	for _, endpoint := range server.Endpoints {
		name := strcase.ToScreamingSnake(endpoint.Name) + "_URL"
		cw.IWritef("export const %s = \"%s\"\n", name, endpoint.Path)
	}

	cw.IWritef("\n")

	cw.IWritef("export class ApiClient extends BaseApiClient {\n")
	cw.Indent()

	cw.IWritef("constructor(baseUrl: string) {\n")
	cw.Indent()
	cw.IWritef("super(baseUrl);\n")
	cw.Unindent()
	cw.IWritef("}\n")

	for _, endpoint := range server.Endpoints {
		cw.IWritef("\n")

		if endpoint.RequireFormData {
			err := genFormDataEndpoint(&cw, &endpoint)
			if err != nil {
				return err
			}
		} else {
			err := genNormalEndpoint(&cw, &endpoint)
			if err != nil {
				return err
			}
		}
	}

	cw.Unindent()
	cw.IWritef("}\n")

	return nil
}
