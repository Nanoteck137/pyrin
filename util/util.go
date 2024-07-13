package util

import (
	"fmt"

	"github.com/nanoteck137/pyrin/resolve"
)

func TypeToString(ty resolve.Type) (string, error) {
	switch ty := ty.(type) {
	case *resolve.TypeString:
		return "string", nil
	case *resolve.TypeInt:
		return "int", nil
	case *resolve.TypeBoolean:
		return "bool", nil
	case *resolve.TypeArray:
		s, err := TypeToString(ty.ElementType)
		if err != nil {
			return "", err
		}
		return "[]" + s, nil
	case *resolve.TypePtr:
		s, err := TypeToString(ty.BaseType)
		if err != nil {
			return "", err
		}
		return "*" + s, nil
	case *resolve.TypeStruct:
		return ty.Name, nil
	case *resolve.TypeSameStruct:
		return ty.Type.Name, nil
	default:
		return "", fmt.Errorf("Unknown resolved type: %T", ty)
	}
}
