package extract

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type Type struct {
	Name     string
	FullName string
	Type     reflect.Type
}

type Context struct {
	types map[string]Type

	nameUsed map[string]int
	names    map[string]string
}

func NewContext() *Context {
	return &Context{
		types:    map[string]Type{},
		nameUsed: map[string]int{},
		names:    map[string]string{},
	}
}

func printIndent(indent int) {
	if indent == 0 {
		return
	}

	for i := 0; i < indent; i++ {
		fmt.Print("  ")
	}
}

func (c *Context) RegisterName(name, pkg string) string {
	fullName := pkg + "-" + name

	used, exists := c.nameUsed[name]
	if !exists {
		c.nameUsed[name] = 1
		c.names[fullName] = name

		return name
	} else {
		c.nameUsed[name] = used + 1

		newName := name + strconv.Itoa(used+1)
		c.names[fullName] = newName

		return newName
	}
}

func (c *Context) translateName(name, pkg string) (string, error) {
	fullName := pkg + "-" + name

	n, exists := c.names[fullName]
	if !exists {
		return "", fmt.Errorf("Name not registered: name: %s pkg: %s", name, pkg)
	}

	return n, nil
}

func (c *Context) checkType(t reflect.Type, indent int) error {
	switch t.Kind() {
	case reflect.Struct:
		return c.CheckStruct(t, indent+1)
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Pointer, reflect.Slice:
		return c.checkType(t.Elem(), indent)
	}

	return nil
}

func (c *Context) CheckStruct(t reflect.Type, indent int) error {
	if t.Kind() != reflect.Struct {
		return errors.New("Type needs to be struct")
	}

	printIndent(indent)
	fmt.Println("Name: ", t.Name())

	fullName := t.PkgPath() + "-" + t.Name()
	_, exists := c.names[fullName]
	if exists {
		return nil
	}

	n := c.RegisterName(t.Name(), t.PkgPath())
	c.types[n] = Type{
		Name:     t.Name(),
		FullName: t.PkgPath() + "-" + t.Name(),
		Type:     t,
	}

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)

		err := c.checkType(sf.Type, indent)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Context) ExtractType(value any) error {
	t := reflect.TypeOf(value)
	return c.CheckStruct(t, 0)
}
