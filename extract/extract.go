package extract

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/nanoteck137/pyrin/ast"
)

type Context struct {
	types map[string]reflect.Type

	nameUsed map[string]int
	names    map[string]string
}

func NewContext() *Context {
	return &Context{
		types:    map[string]reflect.Type{},
		nameUsed: map[string]int{},
		names:    map[string]string{},
	}
}

func (c *Context) isTypeRegisterd(t reflect.Type) bool {
	fullName := t.PkgPath() + "-" + t.Name()
	_, exists := c.names[fullName]
	return exists
}

func (c *Context) registerType(t reflect.Type) {
	name := t.Name()
	fullName := t.PkgPath() + "-" + name

	used, exists := c.nameUsed[name]
	if !exists {
		c.nameUsed[name] = 1
		c.names[fullName] = name
	} else {
		c.nameUsed[name] = used + 1

		newName := name + strconv.Itoa(used+1)
		c.names[fullName] = newName

		name = newName
	}

	c.types[name] = t
}

func (c *Context) translateName(name, pkg string) (string, error) {
	fullName := pkg + "-" + name

	n, exists := c.names[fullName]
	if !exists {
		return "", fmt.Errorf("Name not registered: name: %s pkg: %s", name, pkg)
	}

	return n, nil
}

func (c *Context) checkType(t reflect.Type) error {
	switch t.Kind() {
	case reflect.Struct:
		return c.checkStruct(t)
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Pointer, reflect.Slice:
		return c.checkType(t.Elem())
	}

	return nil
}

func (c *Context) checkStruct(t reflect.Type) error {
	if t.Kind() != reflect.Struct {
		return errors.New("Type needs to be struct")
	}

	if c.isTypeRegisterd(t) {
		return nil
	}

	c.registerType(t)

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)

		if !sf.IsExported() {
			continue
		}

		err := c.checkType(sf.Type)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Context) ExtractTypes(value any) error {
	t := reflect.TypeOf(value)
	return c.checkStruct(t)
}

func (c *Context) getType(t reflect.Type) ast.Typespec {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &ast.IdentTypespec{Ident: "int"}
	case reflect.Bool:
		return &ast.IdentTypespec{Ident: "int"}
	case reflect.String:
		return &ast.IdentTypespec{Ident: "string"}
	case reflect.Float32, reflect.Float64:
		// TODO(patrik): Wrong type
		return &ast.IdentTypespec{Ident: "int"}
	case reflect.Struct:
		name, err := c.translateName(t.Name(), t.PkgPath())
		if err != nil {
			log.Fatal(err)
		}

		return &ast.IdentTypespec{Ident: name}
	case reflect.Slice:
		el := c.getType(t.Elem())
		return &ast.ArrayTypespec{
			Element: el,
		}
	case reflect.Pointer:
		base := c.getType(t.Elem())
		return &ast.PtrTypespec{
			Base: base,
		}

	default:
		log.Fatal("Unknown type ", t.Name(), " ", t.Kind())
	}

	return nil
}

func (c *Context) ConvertToDecls() ([]ast.Decl, error) {
	var decls []ast.Decl

	for k, t := range c.types {
		extend := ""

		var fields []*ast.Field

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			if !f.IsExported() {
				continue
			}

			if f.Anonymous {
				if extend != "" {
					return nil, errors.New("Multiple embedded structs")
				}

				n, err := c.translateName(f.Type.Name(), f.Type.PkgPath())
				if err != nil {
					return nil, err
				}

				extend = n

				continue
			}

			j := f.Tag.Get("json")

			parts := strings.Split(j, ",")

			jname := parts[0]
			joptions := parts[1:]

			hasOmit := false

			for _, v := range joptions {
				if v == "omitempty" {
					hasOmit = true
					break
				}
			}

			ts := c.getType(f.Type)

			name := f.Name
			if jname != "" {
				name = jname
			}

			fields = append(fields, &ast.Field{
				Name: name,
				Type: ts,
				Omit: hasOmit,
			})
		}

		decls = append(decls, &ast.StructDecl{
			Name:   k,
			Extend: extend,
			Fields: fields,
		})
	}

	return decls, nil
}
