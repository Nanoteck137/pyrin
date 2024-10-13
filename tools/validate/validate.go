package validate

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type InternalError struct {
	Err error
}

func (e InternalError) Error() string {
	return "internal error: " + e.Err.Error()
}

func NewInternalError(err error) InternalError {
	return InternalError{
		Err: err,
	}
}

type Errors map[string]error

func (e Errors) Error() string {
	if len(e) == 0 {
		return ""
	}

	keys := make([]string, len(e))
	i := 0
	for key := range e {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	var s strings.Builder
	for i, key := range keys {
		if i > 0 {
			s.WriteString("; ")
		}

		if errs, ok := e[key].(Errors); ok {
			_, _ = fmt.Fprintf(&s, "%v: (%v)", key, errs)
		} else {
			_, _ = fmt.Fprintf(&s, "%v: %v", key, e[key].Error())
		}
	}

	s.WriteString(".")
	return s.String()
}

type Rule interface {
	Validate(value interface{}) error
}

type Field struct {
	p     any
	rules []Rule
}

type Validator interface {
	Struct(p any, fields ...*Field) error
	Field(p any, rules ...Rule) *Field
}

type Validatable interface {
	Validate(validator Validator) error
}

var _ Validator = (*NormalValidator)(nil)

type NormalValidator struct{}

func (*NormalValidator) Field(p any, rules ...Rule) *Field {
	return &Field{
		p:     p,
		rules: rules,
	}
}

func findStructField(structValue reflect.Value, fieldValue reflect.Value) *reflect.StructField {
	ptr := fieldValue.Pointer()
	for i := structValue.NumField() - 1; i >= 0; i-- {
		sf := structValue.Type().Field(i)
		if ptr == structValue.Field(i).UnsafeAddr() {
			// do additional type comparison because it's possible that the address of
			// an embedded struct is the same as the first field of the embedded struct
			if sf.Type == fieldValue.Elem().Type() {
				return &sf
			}
		}
		if sf.Anonymous {
			// delve into anonymous struct to look for the field
			fi := structValue.Field(i)
			if sf.Type.Kind() == reflect.Ptr {
				fi = fi.Elem()
			}
			if fi.Kind() == reflect.Struct {
				if f := findStructField(fi, fieldValue); f != nil {
					return f
				}
			}
		}
	}
	return nil
}

func getErrorFieldName(f *reflect.StructField) string {
	tag := f.Tag.Get("json")

	if tag != "" && tag != "-" {
		splits := strings.SplitN(tag, ",", 2)
		if splits[0] != "" {
			return splits[0]
		}
	}

	return f.Name
}

func validateMap(validator Validator, reflectValue reflect.Value) error {
	errs := Errors{}
	for _, key := range reflectValue.MapKeys() {
		value := reflectValue.MapIndex(key).Interface()
		if value != nil {
			v := value.(Validatable)
			err := v.Validate(validator)
			if err != nil {
				name := fmt.Sprintf("%v", key.Interface())
				errs[name] = err
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func validateSlice(validator Validator, reflectValue reflect.Value) error {
	errs := Errors{}
	for i := 0; i < reflectValue.Len(); i++ {
		value := reflectValue.Index(i).Interface()
		if value != nil {
			v := value.(Validatable)
			err := v.Validate(validator)
			if err != nil {
				errs[strconv.Itoa(i)] = err
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

var validatableType = reflect.TypeOf((*Validatable)(nil)).Elem()

func Validate(validator Validator, value any, rules ...Rule) error {
	for _, rule := range rules {
		// if s, ok := rule.(skipRule); ok && s.skip {
		// 	return nil
		// }

		if err := rule.Validate(value); err != nil {
			return err
		}
	}

	reflectValue := reflect.ValueOf(value)
	if (reflectValue.Kind() == reflect.Ptr || reflectValue.Kind() == reflect.Interface) && reflectValue.IsNil() {
		return nil
	}

	if v, ok := value.(Validatable); ok {
		return v.Validate(validator)
	}

	switch reflectValue.Kind() {
	case reflect.Map:
		if reflectValue.Type().Elem().Implements(validatableType) {
			return validateMap(validator, reflectValue)
		}
	case reflect.Slice, reflect.Array:
		if reflectValue.Type().Elem().Implements(validatableType) {
			return validateSlice(validator, reflectValue)
		}
	case reflect.Ptr, reflect.Interface:
		return Validate(validator, reflectValue.Elem().Interface())
	}

	return nil
}

func (v *NormalValidator) Struct(p any, fields ...*Field) error {
	value := reflect.ValueOf(p)
	if value.Kind() != reflect.Ptr || !value.IsNil() && value.Elem().Kind() != reflect.Struct {
		// must be a pointer to a struct
		return errors.New("Struct not pointer")
	}

	if value.IsNil() {
		// treat a nil struct pointer as valid
		return nil
	}

	value = value.Elem()

	errs := Errors{}
	for _, field := range fields {
		fv := reflect.ValueOf(field.p)
		if fv.Kind() != reflect.Ptr {
			return errors.New("Field not pointer")
		}

		structField := findStructField(value, fv)
		if structField == nil {
			return errors.New("Field not in struct")
		}

		err := Validate(v, fv.Elem().Interface(), field.rules...)
		if err != nil {
			if ie, ok := err.(InternalError); ok {
				return ie.Err
			}

			if structField.Anonymous {
				// merge errors from anonymous struct field
				if es, ok := err.(Errors); ok {
					for name, value := range es {
						errs[name] = value
					}
					continue
				}
			}

			errs[getErrorFieldName(structField)] = err
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
