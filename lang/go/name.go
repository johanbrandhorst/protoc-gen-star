package pgsgo

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/lyft/protoc-gen-star"
)

func (c context) Name(node pgs.Node) pgs.Name {
	// Message or Enum
	type ChildEntity interface {
		Name() pgs.Name
		Parent() pgs.ParentEntity
	}

	switch en := node.(type) {
	case pgs.Package: // the package name for the first file (should be consistent)
		return c.PackageName(en)
	case pgs.File: // the package name for this file
		return c.PackageName(en)
	case ChildEntity: // Message or Enum types, which may be nested
		n := pggUpperCamelCase(en.Name())
		if p, ok := en.Parent().(pgs.Message); ok {
			n = pgs.Name(joinNames(c.Name(p), n))
		}
		return n
	case pgs.Field: // field names cannot conflict with other generated methods
		return replaceProtected(pggUpperCamelCase(en.Name()))
	case pgs.OneOf: // oneof field names cannot conflict with other generated methods
		return replaceProtected(pggUpperCamelCase(en.Name()))
	case pgs.EnumValue: // EnumValue are prefixed with the enum name
		return pgs.Name(joinNames(c.Name(en.Enum()), en.Name()))
	case pgs.Service: // always return the server name
		return c.ServerName(en)
	case pgs.Entity: // any other entity should be just upper-camel-cased
		return pggUpperCamelCase(en.Name())
	default:
		panic("unreachable")
	}
}

func (c context) OneofOption(field pgs.Field) pgs.Name {
	return pgs.Name(joinNames(c.Name(field.Message()), c.Name(field)))
}

func (c context) ServerName(s pgs.Service) pgs.Name {
	n := pggUpperCamelCase(s.Name())
	return pgs.Name(fmt.Sprintf("%sServer", n))
}

func (c context) ClientName(s pgs.Service) pgs.Name {
	n := pggUpperCamelCase(s.Name())
	return pgs.Name(fmt.Sprintf("%sClient", n))
}

// pggUpperCamelCase converts Name n to the protoc-gen-go defined upper
// camelcase. The rules are slightly different from pgs.UpperCamelCase in that
// leading underscores are converted to 'X', mid-string underscores followed by
// lowercase letters are removed and the letter is capitalized, all other
// punctuation is preserved. This method should be used when deriving names of
// protoc-gen-go generated code (ie, message/service struct names and field
// names).
//
// See: https://godoc.org/github.com/golang/protobuf/protoc-gen-go/generator#CamelCase
func pggUpperCamelCase(n pgs.Name) pgs.Name {
	return pgs.Name(generator.CamelCase(n.String()))
}

var protectedNames = map[pgs.Name]pgs.Name{
	"Reset":               "Reset_",
	"String":              "String_",
	"ProtoMessage":        "ProtoMessage_",
	"Marshal":             "Marshal_",
	"Unmarshal":           "Unmarshal_",
	"ExtensionRangeArray": "ExtensionRangeArray_",
	"ExtensionMap":        "ExtensionMap_",
	"Descriptor":          "Descriptor_",
}

func replaceProtected(n pgs.Name) pgs.Name {
	if use, protected := protectedNames[n]; protected {
		return use
	}
	return n
}

func joinNames(a, b pgs.Name) pgs.Name {
	return pgs.Name(fmt.Sprintf("%s_%s", a, b))
}