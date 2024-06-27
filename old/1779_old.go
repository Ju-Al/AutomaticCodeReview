package openapi

func TypeSchema(api *design.APIExpr, t design.DataType) *Schema {
import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"goa.design/goa/design"
	httpdesign "goa.design/goa/http/design"
)

type (
	// Schema represents an instance of a JSON schema.
	// See http://json-schema.org/documentation.html
	Schema struct {
		Schema string `json:"$schema,omitempty" yaml:"$schema,omitempty"`
		// Core schema
		ID           string             `json:"id,omitempty" yaml:"id,omitempty"`
		Title        string             `json:"title,omitempty" yaml:"title,omitempty"`
		Type         Type               `json:"type,omitempty" yaml:"type,omitempty"`
		Items        *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
		Properties   map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
		Definitions  map[string]*Schema `json:"definitions,omitempty" yaml:"definitions,omitempty"`
		Description  string             `json:"description,omitempty" yaml:"description,omitempty"`
		DefaultValue interface{}        `json:"default,omitempty" yaml:"default,omitempty"`
		Example      interface{}        `json:"example,omitempty" yaml:"example,omitempty"`

		// Hyper schema
		Media     *Media  `json:"media,omitempty" yaml:"media,omitempty"`
		ReadOnly  bool    `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
		PathStart string  `json:"pathStart,omitempty" yaml:"pathStart,omitempty"`
		Links     []*Link `json:"links,omitempty" yaml:"links,omitempty"`
		Ref       string  `json:"$ref,omitempty" yaml:"$ref,omitempty"`

		// Validation
		Enum                 []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
		Format               string        `json:"format,omitempty" yaml:"format,omitempty"`
		Pattern              string        `json:"pattern,omitempty" yaml:"pattern,omitempty"`
		Minimum              *float64      `json:"minimum,omitempty" yaml:"minimum,omitempty"`
		Maximum              *float64      `json:"maximum,omitempty" yaml:"maximum,omitempty"`
		MinLength            *int          `json:"minLength,omitempty" yaml:"minLength,omitempty"`
		MaxLength            *int          `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
		MinItems             *int          `json:"minItems,omitempty" yaml:"minItems,omitempty"`
		MaxItems             *int          `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
		Required             []string      `json:"required,omitempty" yaml:"required,omitempty"`
		AdditionalProperties bool          `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`

		// Union
		AnyOf []*Schema `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	}

	// Type is the JSON type enum.
	Type string

	// Media represents a "media" field in a JSON hyper schema.
	Media struct {
		BinaryEncoding string `json:"binaryEncoding,omitempty" yaml:"binaryEncoding,omitempty"`
		Type           string `json:"type,omitempty" yaml:"type,omitempty"`
	}

	// Link represents a "link" field in a JSON hyper schema.
	Link struct {
		Title        string  `json:"title,omitempty" yaml:"title,omitempty"`
		Description  string  `json:"description,omitempty" yaml:"description,omitempty"`
		Rel          string  `json:"rel,omitempty" yaml:"rel,omitempty"`
		Href         string  `json:"href,omitempty" yaml:"href,omitempty"`
		Method       string  `json:"method,omitempty" yaml:"method,omitempty"`
		Schema       *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
		TargetSchema *Schema `json:"targetSchema,omitempty" yaml:"targetSchema,omitempty"`
		ResultType   string  `json:"mediaType,omitempty" yaml:"mediaType,omitempty"`
		EncType      string  `json:"encType,omitempty" yaml:"encType,omitempty"`
	}
)

const (
	// Array represents a JSON array.
	Array Type = "array"
	// Boolean represents a JSON boolean.
	Boolean = "boolean"
	// Integer represents a JSON number without a fraction or exponent part.
	Integer = "integer"
	// Number represents any JSON number. Number includes integer.
	Number = "number"
	// Null represents the JSON null value.
	Null = "null"
	// Object represents a JSON object.
	Object = "object"
	// String represents a JSON string.
	String = "string"
	// File is an extension used by Swagger to represent a file download.
	File = "file"
)

// SchemaRef is the JSON Hyper-schema standard href.
const SchemaRef = "http://json-schema.org/draft-04/hyper-schema"

var (
	// Definitions contains the generated JSON schema definitions
	Definitions map[string]*Schema
)

// Initialize the global variables
func init() {
	Definitions = make(map[string]*Schema)
}

// NewSchema instantiates a new JSON schema.
func NewSchema() *Schema {
	js := Schema{
		Properties:  make(map[string]*Schema),
		Definitions: make(map[string]*Schema),
	}
	return &js
}

// JSON serializes the schema into JSON.
// It makes sure the "$schema" standard field is set if needed prior to
// delegating to the standard // JSON marshaler.
func (s *Schema) JSON() ([]byte, error) {
	if s.Ref == "" {
		s.Schema = SchemaRef
	}
	return json.Marshal(s)
}

// APISchema produces the API JSON hyper schema.
func APISchema(api *design.APIExpr, r *httpdesign.RootExpr) *Schema {
	for _, res := range r.HTTPServices {
		GenerateServiceDefinition(api, res)
	}
	href := api.Servers[0].URL
	links := []*Link{
		{
			Href: href,
			Rel:  "self",
		},
		{
			Href:   "/schema",
			Method: "GET",
			Rel:    "self",
			TargetSchema: &Schema{
				Schema:               SchemaRef,
				AdditionalProperties: true,
			},
		},
	}
	s := Schema{
		ID:          fmt.Sprintf("%s/schema", href),
		Title:       api.Title,
		Description: api.Description,
		Type:        Object,
		Definitions: Definitions,
		Properties:  propertiesFromDefs(Definitions, "#/definitions/"),
		Links:       links,
	}
	return &s
}

// GenerateServiceDefinition produces the JSON schema corresponding to the given
// service. It stores the results in cachedSchema.
func GenerateServiceDefinition(api *design.APIExpr, res *httpdesign.ServiceExpr) {
	s := NewSchema()
	s.Description = res.Description()
	s.Type = Object
	s.Title = res.Name()
	Definitions[res.Name()] = s
	for _, a := range res.HTTPEndpoints {
		var requestSchema *Schema
		if a.MethodExpr.Payload != nil {
			requestSchema = TypeSchema(api, a.MethodExpr.Payload.Type, a.MethodExpr.Payload.Validation)
			requestSchema.Description = a.Name() + " payload"
		}
		if a.Params() != nil {
			params := a.MappedParams()
			// We don't want to keep the path params, these are
			// defined inline in the href
			for _, r := range a.Routes {
				for _, p := range r.Params() {
					design.AsObject(params.Type).Delete(p)
				}
			}
		}
		var targetSchema *Schema
		var identifier string
		for _, resp := range a.Responses {
			dt := resp.Body.Type
			val := resp.Body.Validation
			if mt := dt.(*design.ResultTypeExpr); mt != nil {
				if identifier == "" {
					identifier = mt.Identifier
				} else {
					identifier = ""
				}
				if targetSchema == nil {
					targetSchema = TypeSchema(api, mt, val)
				} else if targetSchema.AnyOf == nil {
					firstSchema := targetSchema
					targetSchema = NewSchema()
					targetSchema.AnyOf = []*Schema{firstSchema, TypeSchema(api, mt, val)}
				} else {
					targetSchema.AnyOf = append(targetSchema.AnyOf, TypeSchema(api, mt, val))
				}
			}
		}
		for i, r := range a.Routes {
			for j, href := range toSchemaHrefs(r) {
				link := Link{
					Title:        a.Name(),
					Rel:          a.Name(),
					Href:         href,
					Method:       r.Method,
					Schema:       requestSchema,
					TargetSchema: targetSchema,
					ResultType:   identifier,
				}
				if i == 0 && j == 0 {
					if ca := a.Service.CanonicalEndpoint(); ca != nil {
						if ca.Name() == a.Name() {
							link.Rel = "self"
						}
					}
				}
				s.Links = append(s.Links, &link)
			}
		}
	}
}

// ResultTypeRef produces the JSON reference to the media type definition with
// the given view.
func ResultTypeRef(api *design.APIExpr, mt *design.ResultTypeExpr, view string) string {
	projected, err := design.Project(mt, view)
	if err != nil {
		panic(fmt.Sprintf("failed to project media type %#v: %s", mt.Identifier, err)) // bug
	}
	if _, ok := Definitions[projected.TypeName]; !ok {
		GenerateResultTypeDefinition(api, projected, "default")
	}
	ref := fmt.Sprintf("#/definitions/%s", projected.TypeName)
	return ref
}

// TypeRef produces the JSON reference to the type definition.
func TypeRef(api *design.APIExpr, ut *design.UserTypeExpr) string {
	if _, ok := Definitions[ut.TypeName]; !ok {
		GenerateTypeDefinition(api, ut)
	}
	return fmt.Sprintf("#/definitions/%s", ut.TypeName)
}

// GenerateResultTypeDefinition produces the JSON schema corresponding to the
// given media type and given view.
func GenerateResultTypeDefinition(api *design.APIExpr, mt *design.ResultTypeExpr, view string) {
	if _, ok := Definitions[mt.TypeName]; ok {
		return
	}
	s := NewSchema()
	s.Title = fmt.Sprintf("Mediatype identifier: %s", mt.Identifier)
	Definitions[mt.TypeName] = s
	buildResultTypeSchema(api, mt, view, s)
}

// GenerateTypeDefinition produces the JSON schema corresponding to the given
// type.
func GenerateTypeDefinition(api *design.APIExpr, ut *design.UserTypeExpr) {
	if _, ok := Definitions[ut.TypeName]; ok {
		return
	}
	s := NewSchema()
	s.Title = ut.TypeName
	Definitions[ut.TypeName] = s
	buildAttributeSchema(api, s, ut.AttributeExpr)
}

// TypeSchema produces the JSON schema corresponding to the given data type.
func TypeSchema(api *design.APIExpr, t design.DataType, val *design.ValidationExpr) *Schema {
	s := NewSchema()
	switch actual := t.(type) {
	case design.Primitive:
		if name := actual.Name(); name != "any" {
			s.Type = Type(actual.Name())
		}
		switch actual.Kind() {
		case design.IntKind, design.Int64Kind,
			design.UIntKind, design.UInt64Kind:
			s.Type = Type("integer")
			s.Format = "int64"
		case design.Int32Kind, design.UInt32Kind:
			s.Type = Type("integer")
			s.Format = "int32"
		case design.Float32Kind:
			s.Type = Type("number")
			s.Format = "float"
		case design.Float64Kind:
			s.Type = Type("number")
			s.Format = "double"
		case design.BytesKind:
			s.Type = Type("string")
			s.Format = "byte"
		}
	case *design.Array:
		s.Type = Array
		s.Items = NewSchema()
		buildAttributeSchema(api, s.Items, actual.ElemType)
	case *design.Object:
		s.Type = Object
		for _, nat := range *actual {
			prop := NewSchema()
			buildAttributeSchema(api, prop, nat.Attribute)
			s.Properties[nat.Name] = prop
		}
	case *design.Map:
		s.Type = Object
		s.AdditionalProperties = true
	case *design.UserTypeExpr:
		s.Ref = TypeRef(api, actual)
	case *design.ResultTypeExpr:
		// Use "default" view by default
		s.Ref = ResultTypeRef(api, actual, design.DefaultView)
	}

	if val == nil {
		return s
	}
	s.Enum = val.Values
	s.Format = string(val.Format)
	s.Pattern = val.Pattern
	if val.Minimum != nil {
		s.Minimum = val.Minimum
	}
	if val.Maximum != nil {
		s.Maximum = val.Maximum
	}
	if val.MinLength != nil {
		if _, ok := t.(*design.Array); ok {
			s.MinItems = val.MinLength
		} else {
			s.MinLength = val.MinLength
		}
	}
	if val.MaxLength != nil {
		if _, ok := t.(*design.Array); ok {
			s.MaxItems = val.MaxLength
		} else {
			s.MaxLength = val.MaxLength
		}
	}
	s.Required = val.Required

	return s
}

type mergeItems []struct {
	a, b   interface{}
	needed bool
}

func (s *Schema) createMergeItems(other *Schema) mergeItems {
	minInt := func(a, b *int) bool { return (a == nil && b != nil) || (a != nil && b != nil && *a > *b) }
	maxInt := func(a, b *int) bool { return (a == nil && b != nil) || (a != nil && b != nil && *a < *b) }
	minFloat64 := func(a, b *float64) bool { return (a == nil && b != nil) || (a != nil && b != nil && *a > *b) }
	maxFloat64 := func(a, b *float64) bool { return (a == nil && b != nil) || (a != nil && b != nil && *a < *b) }

	return mergeItems{
		{&s.ID, other.ID, s.ID == ""},
		{&s.Type, other.Type, s.Type == ""},
		{&s.Ref, other.Ref, s.Ref == ""},
		{&s.Items, other.Items, s.Items == nil},
		{&s.DefaultValue, other.DefaultValue, s.DefaultValue == nil},
		{&s.Title, other.Title, s.Title == ""},
		{&s.Media, other.Media, s.Media == nil},
		{&s.ReadOnly, other.ReadOnly, s.ReadOnly == false},
		{&s.PathStart, other.PathStart, s.PathStart == ""},
		{&s.Enum, other.Enum, s.Enum == nil},
		{&s.Format, other.Format, s.Format == ""},
		{&s.Pattern, other.Pattern, s.Pattern == ""},
		{&s.AdditionalProperties, other.AdditionalProperties, s.AdditionalProperties == false},
		{&s.Minimum, other.Minimum, minFloat64(s.Minimum, other.Minimum)},
		{&s.Maximum, other.Maximum, maxFloat64(s.Maximum, other.Maximum)},
		{&s.MinLength, other.MinLength, minInt(s.MinLength, other.MinLength)},
		{&s.MaxLength, other.MaxLength, maxInt(s.MaxLength, other.MaxLength)},
		{&s.MinItems, other.MinItems, minInt(s.MinItems, other.MinItems)},
		{&s.MaxItems, other.MaxItems, maxInt(s.MaxItems, other.MaxItems)},
	}
}

// Merge does a two level deep merge of other into s.
func (s *Schema) Merge(other *Schema) {
	items := s.createMergeItems(other)
	for _, v := range items {
		if v.needed && v.b != nil {
			reflect.Indirect(reflect.ValueOf(v.a)).Set(reflect.ValueOf(v.b))
		}
	}

	for n, p := range other.Properties {
		if _, ok := s.Properties[n]; !ok {
			if s.Properties == nil {
				s.Properties = make(map[string]*Schema)
			}
			s.Properties[n] = p
		}
	}

	for n, d := range other.Definitions {
		if _, ok := s.Definitions[n]; !ok {
			s.Definitions[n] = d
		}
	}

	for _, l := range other.Links {
		s.Links = append(s.Links, l)
	}

	for _, r := range other.Required {
		s.Required = append(s.Required, r)
	}
}

// Dup creates a shallow clone of the given schema.
func (s *Schema) Dup() *Schema {
	js := Schema{
		ID:                   s.ID,
		Description:          s.Description,
		Schema:               s.Schema,
		Type:                 s.Type,
		DefaultValue:         s.DefaultValue,
		Title:                s.Title,
		Media:                s.Media,
		ReadOnly:             s.ReadOnly,
		PathStart:            s.PathStart,
		Links:                s.Links,
		Ref:                  s.Ref,
		Enum:                 s.Enum,
		Format:               s.Format,
		Pattern:              s.Pattern,
		Minimum:              s.Minimum,
		Maximum:              s.Maximum,
		MinLength:            s.MinLength,
		MaxLength:            s.MaxLength,
		MinItems:             s.MinItems,
		MaxItems:             s.MaxItems,
		Required:             s.Required,
		AdditionalProperties: s.AdditionalProperties,
	}
	for n, p := range s.Properties {
		js.Properties[n] = p.Dup()
	}
	if s.Items != nil {
		js.Items = s.Items.Dup()
	}
	for n, d := range s.Definitions {
		js.Definitions[n] = d.Dup()
	}
	return &js
}

// buildAttributeSchema initializes the given JSON schema that corresponds to
// the given attribute.
func buildAttributeSchema(api *design.APIExpr, s *Schema, at *design.AttributeExpr) *Schema {
	s.Merge(TypeSchema(api, at.Type, at.Validation))
	if s.Ref != "" {
		// Ref is exclusive with other fields
		return s
	}
	s.DefaultValue = toStringMap(at.DefaultValue)
	s.Description = at.Description
	s.Example = at.Example(api.Random())
	return s
}

// toStringMap converts map[interface{}]interface{} to a map[string]interface{}
// when possible.
func toStringMap(val interface{}) interface{} {
	switch actual := val.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range actual {
			m[toString(k)] = toStringMap(v)
		}
		return m
	case []interface{}:
		mapSlice := make([]interface{}, len(actual))
		for i, e := range actual {
			mapSlice[i] = toStringMap(e)
		}
		return mapSlice
	default:
		return actual
	}
}

// toString returns the string representation of the given type.
func toString(val interface{}) string {
	switch actual := val.(type) {
	case string:
		return actual
	case int:
		return strconv.Itoa(actual)
	case float64:
		return strconv.FormatFloat(actual, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(actual)
	default:
		panic("unexpected key type")
	}
}

// toSchemaHrefs produces hrefs that replace the path wildcards with JSON
// schema references when appropriate.
func toSchemaHrefs(r *httpdesign.RouteExpr) []string {
	paths := r.FullPaths()
	res := make([]string, len(paths))
	for i, path := range paths {
		params := httpdesign.ExtractRouteWildcards(path)
		args := make([]interface{}, len(params))
		for j, p := range params {
			args[j] = fmt.Sprintf("/{%s}", p)
		}
		tmpl := httpdesign.WildcardRegex.ReplaceAllLiteralString(path, "%s")
		res[i] = fmt.Sprintf(tmpl, args...)
	}
	return res
}

// propertiesFromDefs creates a Properties map referencing the given definitions
// under the given path.
func propertiesFromDefs(definitions map[string]*Schema, path string) map[string]*Schema {
	res := make(map[string]*Schema, len(definitions))
	for n := range definitions {
		if n == "identity" {
			continue
		}
		s := NewSchema()
		s.Ref = path + n
		res[n] = s
	}
	return res
}

// buildResultTypeSchema initializes s as the JSON schema representing mt for the
// given view.
func buildResultTypeSchema(api *design.APIExpr, mt *design.ResultTypeExpr, view string, s *Schema) {
	s.Media = &Media{Type: mt.Identifier}
	projected, err := design.Project(mt, view)
	if err != nil {
		panic(fmt.Sprintf("failed to project media type %#v: %s", mt.Identifier, err)) // bug
	}
	buildAttributeSchema(api, s, projected.AttributeExpr)
}
