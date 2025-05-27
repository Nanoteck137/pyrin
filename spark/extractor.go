package spark

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type StructRegistry struct {
	types map[string]reflect.Type

	nameUsed map[string]int
	names    map[reflect.Type]string
}

func NewStructRegistry() *StructRegistry {
	return &StructRegistry{
		types:    map[string]reflect.Type{},
		nameUsed: map[string]int{},
		names:    map[reflect.Type]string{},
	}
}

func (c *StructRegistry) isTypeRegisterd(t reflect.Type) bool {
	// fullName := t.PkgPath() + "-" + t.Name()
	_, exists := c.names[t]
	return exists
}

func (c *StructRegistry) registerType(t reflect.Type) {
	name := t.Name()
	// fullName := t.PkgPath() + "-" + name

	used, exists := c.nameUsed[name]
	if !exists {
		c.nameUsed[name] = 1
		c.names[t] = name
	} else {
		c.nameUsed[name] = used + 1

		newName := name + strconv.Itoa(used+1)
		c.names[t] = newName

		name = newName
	}

	c.types[name] = t
}

func (c *StructRegistry) TranslateName(t reflect.Type) (string, error) {
	n, exists := c.names[t]
	if !exists {
		return "", fmt.Errorf("name not registered")
	}

	return n, nil
}

func (c *StructRegistry) checkType(t reflect.Type) error {
	switch t.Kind() {
	case reflect.Struct:
		return c.check(t)
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Pointer, reflect.Slice:
		return c.checkType(t.Elem())
	}

	return nil
}

func (c *StructRegistry) check(t reflect.Type) error {
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

func (c *StructRegistry) Register(value any) error {
	if value == nil {
		return nil
	}

	// TODO(patrik): Add checks
	t := reflect.TypeOf(value)
	return c.check(t)
}

func (c *StructRegistry) getType(t reflect.Type) Typespec {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &IdentTypespec{Ident: "int"}
	case reflect.Bool:
		return &IdentTypespec{Ident: "bool"}
	case reflect.String:
		return &IdentTypespec{Ident: "string"}
	case reflect.Float32, reflect.Float64:
		return &IdentTypespec{Ident: "float"}
	case reflect.Struct:
		name, err := c.TranslateName(t)
		if err != nil {
			// TODO(patrik): Fix
			log.Fatal(err)
		}

		return &IdentTypespec{Ident: name}
	case reflect.Slice:
		el := c.getType(t.Elem())
		return &ArrayTypespec{
			Element: el,
		}
	case reflect.Pointer:
		base := c.getType(t.Elem())
		return &PtrTypespec{
			Base: base,
		}

	default:
		// TODO(patrik): Fix
		log.Fatal("Unknown type ", t.Name(), " ", t.Kind())
	}

	return nil
}

func (c *StructRegistry) GetStructDecls() ([]StructDecl, error) {
	var res []StructDecl

	for k, t := range c.types {
		extend := ""

		var fields []*FieldDecl

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			if !f.IsExported() {
				continue
			}

			if f.Anonymous {
				if extend != "" {
					return nil, errors.New("Multiple embedded structs")
				}

				n, err := c.TranslateName(f.Type)
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

			omitEmpty := false

			for _, v := range joptions {
				if v == "omitempty" {
					omitEmpty = true
					break
				}
			}

			ts := c.getType(f.Type)

			name := f.Name
			if jname != "" {
				name = jname
			}

			fields = append(fields, &FieldDecl{
				Name: name,
				Type: ts,
				OmitEmpty: omitEmpty,
			})
		}

		res = append(res, StructDecl{
			Name:   k,
			Extend: extend,
			Fields: fields,
		})
	}

	return res, nil
}
