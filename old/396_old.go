package python

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

const (
	lang             = "py"
	defaultOutputDir = "gen-py"
	tab              = "    "
	tabtab           = tab + tab
	tabtabtab        = tab + tab + tab
	tabtabtabtab     = tab + tab + tab + tab
	tabtabtabtabtab  = tab + tab + tab + tab + tab
)


// Generator implements the LanguageGenerator interface for Python.
type Generator struct {
	*generator.BaseGenerator
	outputDir string
	typesFile *os.File
}

// NewGenerator creates a new Python LanguageGenerator.
func NewGenerator(options map[string]string) generator.LanguageGenerator {
	gen := &Generator{&generator.BaseGenerator{Options: options}, "", nil}
	switch getAsyncOpt(options) {
	case tornado:
		return &TornadoGenerator{gen}
	case asyncio:
		return &AsyncIOGenerator{gen}
	}
	return gen
}

// SetupGenerator performs any setup logic before generation.
func (g *Generator) SetupGenerator(outputDir string) error {
	g.outputDir = outputDir

	dir := g.outputDir
	for filepath.Dir(dir) != "." {
		file, err := g.GenerateFile("__init__", dir, generator.ObjectFile)
		file.Close()
		if err != nil {
			return err
		}

		dir = filepath.Dir(dir)
	}

	// create types file
	typesFile, err := g.GenerateFile("ttypes", outputDir, generator.ObjectFile)
	if err != nil {
		return err
	}
	if err = g.GenerateDocStringComment(typesFile); err != nil {
		return err
	}
	if _, err = typesFile.WriteString("\n\n"); err != nil {
		return err
	}
	if err = g.GenerateTypesImports(typesFile, false); err != nil {
		return err
	}
	if _, err = typesFile.WriteString("\n\n"); err != nil {
		return err
	}
	g.typesFile = typesFile

	return nil
}

// TeardownGenerator is run after generation.
func (g *Generator) TeardownGenerator() error {
	return g.typesFile.Close()
}

// GenerateConstantsContents generates constants.
func (g *Generator) GenerateConstantsContents(constants []*parser.Constant) error {
	if len(constants) == 0 {
		return nil
	}

	contents := "\n\n"
	contents += "from thrift.Thrift import TType, TMessageType, TException, TApplicationException\n"
	contents += "from .ttypes import *\n\n"

	for _, constant := range constants {
		value := g.generateConstantValue(constant.Type, constant.Value, "")
		contents += fmt.Sprintf("%s = %s\n", constant.Name, value)
	}

	file, err := g.GenerateFile("constants", g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.GenerateDocStringComment(file); err != nil {
		return err
	}
	_, err = file.WriteString(contents)
	return err
}

func (g *Generator) generateConstantValue(t *parser.Type, value interface{}, ind string) string {
	if value == nil {
		return "None"
	}

	underlyingType := g.Frugal.UnderlyingType(t)
	// If the value being referenced is of type Identifier, it's referencing
	// another constant. Need to recurse to get that value.
	identifier, ok := value.(parser.Identifier)
	// TODO consolidate this between generators
	if ok {
		name := string(identifier)

		// split based on '.', if present, it should be from an include
		pieces := strings.Split(name, ".")
		switch len(pieces) {
		case 1:
			// From this file
			for _, constant := range g.Frugal.Thrift.Constants {
				if name == constant.Name {
					return g.generateConstantValue(t, constant.Value, ind)
				}
			}
		case 2:
			// Either from an include, or part of an enum
			for _, enum := range g.Frugal.Thrift.Enums {
				if pieces[0] == enum.Name {
					for _, value := range enum.Values {
						if pieces[1] == value.Name {
							return fmt.Sprintf("%v", value.Value)
						}
					}
					panic(fmt.Sprintf("referenced value '%s' of enum '%s' doesn't exist", pieces[1], pieces[0]))
				}
			}

			// If not part of an enum , it's from an include
			include, ok := g.Frugal.ParsedIncludes[pieces[0]]
			if !ok {
				panic(fmt.Sprintf("referenced include '%s' in constant '%s' not present", pieces[0], name))
			}
			for _, constant := range include.Thrift.Constants {
				if pieces[1] == constant.Name {
					return g.generateConstantValue(t, constant.Value, ind)
				}
			}
		case 3:
			// enum from an include
			include, ok := g.Frugal.ParsedIncludes[pieces[0]]
			if !ok {
				panic(fmt.Sprintf("referenced include '%s' in constant '%s' not present", pieces[0], name))
			}
			for _, enum := range include.Thrift.Enums {
				if pieces[1] == enum.Name {
					for _, value := range enum.Values {
						if pieces[2] == value.Name {
							return fmt.Sprintf("%v", value.Value)
						}
					}
					panic(fmt.Sprintf("referenced value '%s' of enum '%s' doesn't exist", pieces[1], pieces[0]))
				}
			}
		default:
			panic("reference constant doesn't exist: " + name)
		}
	}

	if parser.IsThriftPrimitive(underlyingType) || parser.IsThriftContainer(underlyingType) {
		switch underlyingType.Name {
		case "bool":
			return strings.Title(fmt.Sprintf("%v", value))
		case "i8", "byte", "i16", "i32", "i64", "double":
			return fmt.Sprintf("%v", value)
		case "string", "binary":
			return fmt.Sprintf("\"%s\"", value)
		case "list", "set":
			contents := ""
			if underlyingType.Name == "set" {
				contents += "set("
			}
			contents += "[
"
			for _, v := range value.([]interface{}) {
				val := g.generateConstantValue(underlyingType.ValueType, v, ind+tab)
				contents += fmt.Sprintf(ind+tab+"%s,
", val)
			}
			contents += ind + "]"
			if underlyingType.Name == "set" {
				contents += ")"
			}
			return contents
		case "map":
			contents := "{
"
			for _, pair := range value.([]parser.KeyValue) {
				key := g.generateConstantValue(underlyingType.KeyType, pair.Key, ind+tab)
				val := g.generateConstantValue(underlyingType.ValueType, pair.Value, ind+tab)
				contents += fmt.Sprintf(ind+tab+"%s: %s,
", key, val)
			}
			contents += ind + "}"
			return contents
		}
	} else if g.Frugal.IsEnum(underlyingType) {
		return fmt.Sprintf("%d", value)
	} else if g.Frugal.IsStruct(underlyingType) {
		var s *parser.Struct
		for _, potential := range g.Frugal.Thrift.Structs {
			if underlyingType.Name == potential.Name {
				s = potential
				break
			}
		}

		contents := ""

		contents += fmt.Sprintf("%s(**{
", g.qualifiedTypeName(underlyingType))
		for _, pair := range value.([]parser.KeyValue) {
			name := pair.Key.(string)
			for _, field := range s.Fields {
				if name == field.Name {
					val := g.generateConstantValue(field.Type, pair.Value, ind+tab)
					contents += fmt.Sprintf(tab+ind+"\"%s\": %s,
", name, val)
				}
			}
		}
		contents += ind + "})"
		return contents
	}

	panic("no entry for type " + underlyingType.Name)
}

// GenerateTypeDef generates the given typedef.
func (g *Generator) GenerateTypeDef(*parser.TypeDef) error {
	// No typedefs in python
	return nil
}

// GenerateEnum generates the given enum.
func (g *Generator) GenerateEnum(enum *parser.Enum) error {
	contents := ""
	contents += fmt.Sprintf("class %s:
", enum.Name)
	if enum.Comment != nil {
		contents += g.generateDocString(enum.Comment, tab)
	}
	for _, value := range enum.Values {
		contents += fmt.Sprintf(tab+"%s = %d
", value.Name, value.Value)
	}
	contents += "\n"

	contents += tab + "_VALUES_TO_NAMES = {
"
	for _, value := range enum.Values {
		contents += fmt.Sprintf(tabtab+"%d: \"%s\",
", value.Value, value.Name)
	}
	contents += tab + "}

"

	contents += tab + "_NAMES_TO_VALUES = {
"
	for _, value := range enum.Values {
		contents += fmt.Sprintf(tabtab+"\"%s\": %d,
", value.Name, value.Value)
	}
	contents += tab + "}

"

	_, err := g.typesFile.WriteString(contents)
	return err
}

// GenerateStruct generates the given struct.
func (g *Generator) GenerateStruct(s *parser.Struct) error {
	_, err := g.typesFile.WriteString(g.generateStruct(s))
	return err
}

// GenerateUnion generates the given union.
func (g *Generator) GenerateUnion(union *parser.Struct) error {
	// TODO 2.0 consider adding validation only one field is set,
	// similar to other languages
	_, err := g.typesFile.WriteString(g.generateStruct(union))
	return err
}

// GenerateException generates the given exception.
func (g *Generator) GenerateException(exception *parser.Struct) error {
	_, err := g.typesFile.WriteString(g.generateStruct(exception))
	return err
}

// GenerateServiceArgsResults generates the args and results objects for the
// given service.
func (g *Generator) GenerateServiceArgsResults(serviceName string, outputDir string, structs []*parser.Struct) error {
	file, err := g.GenerateFile(serviceName, g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.GenerateDocStringComment(file); err != nil {
		return err
	}
	if _, err = file.WriteString("\n\n"); err != nil {
		return err
	}
	if err = g.GenerateTypesImports(file, true); err != nil {
		return err
	}
	if _, err = file.WriteString("\n\n"); err != nil {
		return err
	}

	contents := ""
	for _, s := range structs {
		contents += g.generateStruct(s)
	}

	_, err = file.WriteString(contents)
	return err
}

// generateStruct generates a python representation of a thrift struct
func (g *Generator) generateStruct(s *parser.Struct) string {
	contents := ""

	extends := ""
	if s.Type == parser.StructTypeException {
		extends = "(TException)"
	}
	contents += fmt.Sprintf("class %s%s:
", s.Name, extends)
	contents += g.generateClassDocstring(s)

	contents += g.generateDefaultMarkers(s)
	contents += g.generateInit(s)

	contents += g.generateRead(s)
	contents += g.generateWrite(s)

	contents += g.generateValidate(s)
	contents += g.generateMagicMethods(s)

	return contents
}

// generateDefaultMarkers generates marker objects to provide as defaults to
// an __init__ method. The __init__ method can then determine if the default
// was provided and generate the constant appropriately. Don't generate the
// constant as a class variable to avoid order of declaration issues.
func (g *Generator) generateDefaultMarkers(s *parser.Struct) string {
	contents := ""
	for _, field := range s.Fields {
		if field.Default != nil {
			underlyingType := g.Frugal.UnderlyingType(field.Type)
			// use 'object()' as a marker value to avoid instantiating
			// a class defined later in the file
			defaultVal := "object()"
			if parser.IsThriftPrimitive(underlyingType) || g.Frugal.IsEnum(underlyingType) {
				defaultVal = g.generateConstantValue(underlyingType, field.Default, tab)
			}
			contents += fmt.Sprintf(tab+"_DEFAULT_%s_MARKER = %s
", field.Name, defaultVal)
		}
	}
	return contents
}

// generateInit generates the init method for a class.
func (g *Generator) generateInit(s *parser.Struct) string {
	if len(s.Fields) == 0 {
		return ""
	}

	contents := ""
	argList := ""
	for _, field := range s.Fields {
		defaultVal := "None"
		if field.Default != nil {
			defaultVal = fmt.Sprintf("_DEFAULT_%s_MARKER", field.Name)
		}
		argList += fmt.Sprintf(", %s=%s", field.Name, defaultVal)
	}
	contents += fmt.Sprintf(tab+"def __init__(self%s):
", argList)
	for _, field := range s.Fields {
		underlyingType := g.Frugal.UnderlyingType(field.Type)
		if !parser.IsThriftPrimitive(underlyingType) && !g.Frugal.IsEnum(underlyingType) && field.Default != nil {
			contents += fmt.Sprintf(tabtab+"if %s is self._DEFAULT_%s_MARKER:
", field.Name, field.Name)
			val := g.generateConstantValue(field.Type, field.Default, tabtabtab)
			contents += fmt.Sprintf(tabtabtab+"%s = %s\n", field.Name, val)
		}
		contents += fmt.Sprintf(tabtab+"self.%s = %s
", field.Name, field.Name)
	}
	contents += "\n"
	return contents
}

// generateClassDocstring generates a docstring for class. This includes a
// description of the class, if present, a list of attributes, and descriptions
// of each attribute, if present.
func (g *Generator) generateClassDocstring(s *parser.Struct) string {
	lines := []string{}
	if s.Comment != nil {
		lines = append(lines, s.Comment...)
		lines = append(lines, "")
	}

	if len(s.Fields) > 0 {
		lines = append(lines, "Attributes:")
		for _, field := range s.Fields {
			line := fmt.Sprintf(" - %s", field.Name)
			if len(field.Comment) > 0 {
				line = fmt.Sprintf("%s: %s", line, field.Comment[0])
				lines = append(lines, line)
				lines = append(lines, field.Comment[1:]...)
			} else {
				lines = append(lines, line)
			}
		}
	}

	if len(lines) == 0 {
		return ""
	}

	return g.generateDocString(lines, tab)
}

// generateRead generates the read method for a struct.
func (g *Generator) generateRead(s *parser.Struct) string {
	contents := ""
	contents += tab + "def read(self, iprot):
"
	contents += tabtab + "iprot.readStructBegin()
"
	contents += tabtab + "while True:
"
	contents += tabtabtab + "(fname, ftype, fid) = iprot.readFieldBegin()
"
	contents += tabtabtab + "if ftype == TType.STOP:
"
	contents += tabtabtabtab + "break
"
	ifstatement := "if"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtabtab+"%s fid == %d:
", ifstatement, field.ID)
		contents += fmt.Sprintf(tabtabtabtab+"if ftype == %s:
", g.getTType(field.Type))
		contents += g.generateReadFieldRec(field, true, tabtabtabtabtab)
		contents += tabtabtabtab + "else:
"
		contents += tabtabtabtabtab + "iprot.skip(ftype)
"
		ifstatement = "elif"
	}
	contents += tabtabtab + "else:
"
	contents += tabtabtabtab + "iprot.skip(ftype)
"
	contents += tabtabtab + "iprot.readFieldEnd()
"
	contents += tabtab + "iprot.readStructEnd()

"
	return contents
}

// generateWrite generates the write method for a struct.
func (g *Generator) generateWrite(s *parser.Struct) string {
	contents := ""
	contents += tab + "def write(self, oprot):
"
	contents += fmt.Sprintf(tabtab+"oprot.writeStructBegin('%s')
", s.Name)
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtab+"if self.%s is not None:
", field.Name)
		contents += fmt.Sprintf(tabtabtab+"oprot.writeFieldBegin('%s', %s, %d)
", field.Name, g.getTType(field.Type), field.ID)
		contents += g.generateWriteFieldRec(field, true, tabtabtab)
		contents += fmt.Sprintf(tabtabtab + "oprot.writeFieldEnd()
")
	}

	contents += tabtab + "oprot.writeFieldStop()
"
	contents += tabtab + "oprot.writeStructEnd()

"
	return contents
}

// generateValidate generates a validate method for a class. This ensures
// required fields are present.
func (g *Generator) generateValidate(s *parser.Struct) string {
	contents := ""
	contents += tab + "def validate(self):
"
	for _, field := range s.Fields {
		if field.Modifier == parser.Required {
			contents += fmt.Sprintf(tabtab+"if self.%s is None:
", field.Name)
			contents += fmt.Sprintf(tabtabtab+"raise TProtocol.TProtocolException(message='Required field %s is unset!')
", field.Name)
		}
	}
	contents += tabtab + "return

"
	return contents
}

// generateMagicMethods generates magic methods for the class, such as
// '__hash__', '__repr__', '__eq__', and '__ne__'.
func (g *Generator) generateMagicMethods(s *parser.Struct) string {
	contents := ""
	if s.Type == parser.StructTypeException {
		contents += tab + "def __str__(self):
"
		contents += tabtab + "return repr(self)

"
	}

	contents += tab + "def __hash__(self):
"
	contents += tabtab + "value = 17
"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtab+"value = (value * 31) ^ hash(self.%s)
", field.Name)
	}
	contents += tabtab + "return value

"

	contents += tab + "def __repr__(self):
"
	contents += tabtab + "L = ['%s=%r' % (key, value)
"
	contents += tabtabtab + "for key, value in self.__dict__.items()]
"
	contents += tabtab + "return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

"

	contents += tab + "def __eq__(self, other):
"
	contents += tabtab + "return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

"

	contents += tab + "def __ne__(self, other):
"
	contents += tabtab + "return not (self == other)

"
	return contents
}

// generateSpecArgs is a recursive function that returns the type of the
// argument in the format thrift_spec requires.
func (g *Generator) generateSpecArgs(t *parser.Type) string {
	underlyingType := g.Frugal.UnderlyingType(t)

	if parser.IsThriftPrimitive(underlyingType) {
		return "None"
	} else if parser.IsThriftContainer(underlyingType) {
		switch underlyingType.Name {
		case "list", "set":
			return fmt.Sprintf("(%s, %s)", g.getTType(underlyingType.ValueType), g.generateSpecArgs(underlyingType.ValueType))
		case "map":
			return fmt.Sprintf("(%s, %s, %s, %s)",
				g.getTType(underlyingType.KeyType), g.generateSpecArgs(underlyingType.KeyType),
				g.getTType(underlyingType.ValueType), g.generateSpecArgs(underlyingType.ValueType))
		}
	} else if g.Frugal.IsEnum(underlyingType) {
		return "None"
	} else if g.Frugal.IsStruct(underlyingType) {
		qualifiedName := g.qualifiedTypeName(underlyingType)
		return fmt.Sprintf("(%s, %s.thrift_spec)", qualifiedName, qualifiedName)
	}

	panic("unrecognized type: " + t.Name)
}

// generateReadFieldRec recursively generates code to read a field.
func (g *Generator) generateReadFieldRec(field *parser.Field, first bool, ind string) string {
	contents := ""

	prefix := ""
	if first {
		prefix = "self."
	}
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	isEnum := g.Frugal.IsEnum(underlyingType)
	if parser.IsThriftPrimitive(underlyingType) || isEnum {
		thriftType := ""
		switch underlyingType.Name {
		case "bool", "byte", "i16", "i32", "i64", "double", "string":
			thriftType = strings.Title(underlyingType.Name)
		case "i8":
			thriftType = "Byte"
		case "binary":
			thriftType = "String"
		default:
			if isEnum {
				thriftType = "I32"
			} else {
				panic("unknown type: " + underlyingType.Name)
			}
		}
		contents += fmt.Sprintf(ind+"%s%s = iprot.read%s()
", prefix, field.Name, thriftType)
	} else if g.Frugal.IsStruct(underlyingType) {
		g.qualifiedTypeName(underlyingType)
		contents += fmt.Sprintf(ind+"%s%s = %s()
", prefix, field.Name, g.qualifiedTypeName(underlyingType))
		contents += fmt.Sprintf(ind+"%s%s.read(iprot)
", prefix, field.Name)
	} else if parser.IsThriftContainer(underlyingType) {
		sizeElem := getElem()
		valElem := getElem()
		valField := parser.FieldFromType(underlyingType.ValueType, valElem)

		switch underlyingType.Name {
		case "list":
			contents += fmt.Sprintf(ind+"%s%s = []
", prefix, field.Name)
			contents += fmt.Sprintf(ind+"(_, %s) = iprot.readListBegin()
", sizeElem)
			contents += fmt.Sprintf(ind+"for _ in range(%s):
", sizeElem)
			contents += g.generateReadFieldRec(valField, false, ind+tab)
			contents += fmt.Sprintf(ind+tab+"%s%s.append(%s)
", prefix, field.Name, valElem)
			contents += fmt.Sprintf(ind + "iprot.readListEnd()
")
		case "set":
			contents += fmt.Sprintf(ind+"%s%s = set()
", prefix, field.Name)
			contents += fmt.Sprintf(ind+"(_, %s) = iprot.readSetBegin()
", sizeElem)
			contents += fmt.Sprintf(ind+"for _ in range(%s):
", sizeElem)
			contents += g.generateReadFieldRec(valField, false, ind+tab)
			contents += fmt.Sprintf(ind+tab+"%s%s.add(%s)
", prefix, field.Name, valElem)
			contents += fmt.Sprintf(ind + "iprot.readSetEnd()
")
		case "map":
			contents += fmt.Sprintf(ind+"%s%s = {}
", prefix, field.Name)
			contents += fmt.Sprintf(ind+"(_, _, %s) = iprot.readMapBegin()
", sizeElem)
			contents += fmt.Sprintf(ind+"for _ in range(%s):
", sizeElem)
			keyElem := getElem()
			keyField := parser.FieldFromType(underlyingType.KeyType, keyElem)
			contents += g.generateReadFieldRec(keyField, false, ind+tab)
			contents += g.generateReadFieldRec(valField, false, ind+tab)
			contents += fmt.Sprintf(ind+tab+"%s%s[%s] = %s
", prefix, field.Name, keyElem, valElem)
			contents += fmt.Sprintf(ind + "iprot.readMapEnd()
")
		}
	}

	return contents
}

// generateReadFieldRec recursively generates code to write a field.
func (g *Generator) generateWriteFieldRec(field *parser.Field, first bool, ind string) string {
	contents := ""

	prefix := ""
	if first {
		prefix = "self."
	}
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	isEnum := g.Frugal.IsEnum(underlyingType)
	if parser.IsThriftPrimitive(underlyingType) || isEnum {
		thriftType := ""
		switch underlyingType.Name {
		case "bool", "byte", "i16", "i32", "i64", "double", "string":
			thriftType = strings.Title(underlyingType.Name)
		case "i8":
			thriftType = "Byte"
		case "binary":
			thriftType = "String"
		default:
			if isEnum {
				thriftType = "I32"
			} else {
				panic("unknown type: " + underlyingType.Name)
			}
		}
		contents += fmt.Sprintf(ind+"oprot.write%s(%s%s)
", thriftType, prefix, field.Name)
	} else if g.Frugal.IsStruct(underlyingType) {
		contents += fmt.Sprintf(ind+"%s%s.write(oprot)
", prefix, field.Name)
	} else if parser.IsThriftContainer(underlyingType) {
		valElem := getElem()
		valField := parser.FieldFromType(underlyingType.ValueType, valElem)
		valTType := g.getTType(underlyingType.ValueType)

		switch underlyingType.Name {
		case "list":
			contents += fmt.Sprintf(ind+"oprot.writeListBegin(%s, len(%s%s))
", valTType, prefix, field.Name)
			contents += fmt.Sprintf(ind+"for %s in %s%s:
", valElem, prefix, field.Name)
			contents += g.generateWriteFieldRec(valField, false, ind+tab)
			contents += ind + "oprot.writeListEnd()
"
		case "set":
			contents += fmt.Sprintf(ind+"oprot.writeSetBegin(%s, len(%s%s))
", valTType, prefix, field.Name)
			contents += fmt.Sprintf(ind+"for %s in %s%s:
", valElem, prefix, field.Name)
			contents += g.generateWriteFieldRec(valField, false, ind+tab)
			contents += ind + "oprot.writeSetEnd()
"
		case "map":
			keyElem := getElem()
			keyField := parser.FieldFromType(underlyingType.KeyType, keyElem)
			keyTType := g.getTType(underlyingType.KeyType)
			contents += fmt.Sprintf(ind+"oprot.writeMapBegin(%s, %s, len(%s%s))
", keyTType, valTType, prefix, field.Name)
			contents += fmt.Sprintf(ind+"for %s, %s in %s%s.items():
", keyElem, valElem, prefix, field.Name)
			contents += g.generateWriteFieldRec(keyField, false, ind+tab)
			contents += g.generateWriteFieldRec(valField, false, ind+tab)
			contents += ind + "oprot.writeMapEnd()
"
		}
	}

	return contents
}

// GetOutputDir returns the output directory for generated files.
func (g *Generator) GetOutputDir(dir string) string {
	if pkg, ok := g.Frugal.Thrift.Namespace(lang); ok {
		path := generator.GetPackageComponents(pkg)
		dir = filepath.Join(append([]string{dir}, path...)...)
	} else {
		dir = filepath.Join(dir, g.Frugal.Name)
	}
	return dir
}

// DefaultOutputDir returns the default output directory for generated files.
func (g *Generator) DefaultOutputDir() string {
	dir := defaultOutputDir
	switch getAsyncOpt(g.Options) {
	case tornado:
		dir += ".tornado"
	case asyncio:
		dir += ".asyncio"
	}
	return dir
}

// PostProcess is called after generating each file.
func (g *Generator) PostProcess(f *os.File) error { return nil }

// GenerateDependencies is a no-op.
func (g *Generator) GenerateDependencies(dir string) error {
	return nil
}

// GenerateFile generates the given FileType.
func (g *Generator) GenerateFile(name, outputDir string, fileType generator.FileType) (*os.File, error) {
	switch fileType {
	case generator.PublishFile:
		return g.CreateFile(fmt.Sprintf("f_%s_publisher", name), outputDir, lang, false)
	case generator.SubscribeFile:
		return g.CreateFile(fmt.Sprintf("f_%s_subscriber", name), outputDir, lang, false)
	case generator.CombinedServiceFile:
		return g.CreateFile(fmt.Sprintf("f_%s", name), outputDir, lang, false)
	case generator.ObjectFile:
		return g.CreateFile(fmt.Sprintf("%s", name), outputDir, lang, false)
	default:
		return nil, fmt.Errorf("Bad file type for Python generator: %s", fileType)
	}
}

// GenerateDocStringComment generates the autogenerated notice.
func (g *Generator) GenerateDocStringComment(file *os.File) error {
	comment := fmt.Sprintf(
		"#
"+
			"# Autogenerated by Frugal Compiler (%s)
"+
			"#
"+
			"# DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
"+
			"#",
		globals.Version)

	_, err := file.WriteString(comment)
	return err
}

// GenerateServicePackage is a no-op.
func (g *Generator) GenerateServicePackage(file *os.File, s *parser.Service) error {
	return nil
}

// GenerateScopePackage is a no-op.
func (g *Generator) GenerateScopePackage(file *os.File, s *parser.Scope) error {
	return nil
}

func (g *Generator) GenerateTypesImports(file *os.File, isArgsOrResult bool) error {
	contents := ""
	contents += "from thrift.Thrift import TType, TMessageType, TException, TApplicationException\n"
	for _, include := range g.Frugal.Thrift.Includes {
		includeName, ok := g.Frugal.NamespaceForInclude(filepath.Base(include.Name), lang)
		if !ok {
			includeName = include.Name
		}
		contents += fmt.Sprintf("import %s.ttypes
", includeName)
	}
	contents += "\n"
	if isArgsOrResult {
		contents += "from .ttypes import *
"
	}
	contents += "from thrift.transport import TTransport
"
	contents += "from thrift.protocol import TBinaryProtocol, TProtocol
"

	_, err := file.WriteString(contents)
	return err
}

// GenerateServiceImports generates necessary imports for the given service.
func (g *Generator) GenerateServiceImports(file *os.File, s *parser.Service) error {
	// TODO
	return nil
}

// GenerateScopeImports generates necessary imports for the given scope.
func (g *Generator) GenerateScopeImports(file *os.File, s *parser.Scope) error {
	imports := "from thrift.Thrift import TMessageType
"
	imports += "from frugal.middleware import Method
"
	_, err := file.WriteString(imports)
	return err
}

// GenerateConstants generates any static constants.
func (g *Generator) GenerateConstants(file *os.File, name string) error {
	return nil
}

// GeneratePublisher generates the publisher for the given scope.
func (g *Generator) GeneratePublisher(file *os.File, scope *parser.Scope) error {
	publisher := ""
	publisher += fmt.Sprintf("class %sPublisher(object):
", scope.Name)
	if scope.Comment != nil {
		publisher += g.generateDocString(scope.Comment, tab)
	}
	publisher += "\n"

	publisher += tab + fmt.Sprintf("_DELIMITER = '%s'

", globals.TopicDelimiter)

	publisher += tab + "def __init__(self, provider, middleware=None):
"
	publisher += g.generateDocString([]string{
		fmt.Sprintf("Create a new %sPublisher.
", scope.Name),
		"Args:",
		tab + "provider: FScopeProvider",
		tab + "middleware: ServiceMiddleware or list of ServiceMiddleware",
	}, tabtab)
	publisher += "\n"

	publisher += tabtab + "if middleware and not isinstance(middleware, list):
"
	publisher += tabtabtab + "middleware = [middleware]
"
	publisher += tabtab + "self._transport, protocol_factory = provider.new()
"
	publisher += tabtab + "self._protocol = protocol_factory.get_protocol(self._transport)
"
	publisher += tabtab + "self._methods = {
"
	for _, op := range scope.Operations {
		publisher += tabtabtab + fmt.Sprintf("'publish_%s': Method(self._publish_%s, middleware),
", op.Name, op.Name)
	}
	publisher += tabtab + "}

"

	publisher += tab
	switch getAsyncOpt(g.Options) {
	case tornado:
		publisher += "@gen.coroutine
" + tab
	case asyncio:
		publisher += "async "
	}
	publisher += "def open(self):
"

	publisher += tabtab
	switch getAsyncOpt(g.Options) {
	case tornado:
		publisher += "yield "
	case asyncio:
		publisher += "await "
	}
	publisher += "self._transport.open()

"

	publisher += tab
	switch getAsyncOpt(g.Options) {
	case tornado:
		publisher += "@gen.coroutine
" + tab
	case asyncio:
		publisher += "async "
	}
	publisher += "def close(self):
"

	publisher += tabtab
	switch getAsyncOpt(g.Options) {
	case tornado:
		publisher += "yield "
	case asyncio:
		publisher += "await "
	}
	publisher += "self._transport.close()

"

	prefix := ""
	for _, op := range scope.Operations {
		publisher += prefix + g.generatePublishMethod(scope, op)
		prefix = "\n\n"
	}

	_, err := file.WriteString(publisher)
	return err
}

func (g *Generator) generatePublishMethod(scope *parser.Scope, op *parser.Operation) string {
	args := ""
	docstr := []string{"Args:", tab + "ctx: FContext"}
	if len(scope.Prefix.Variables) > 0 {
		prefix := ""
		for _, variable := range scope.Prefix.Variables {
			docstr = append(docstr, tab+fmt.Sprintf("%s: string", variable))
			args += prefix + variable
			prefix = ", "
		}
		args += ", "
	}
	docstr = append(docstr, tab+fmt.Sprintf("req: %s", op.Type.Name))
	if op.Comment != nil {
		docstr[0] = "\n" + tabtab + docstr[0]
		docstr = append(op.Comment, docstr...)
	}

	method := tab
	if getAsyncOpt(g.Options) == asyncio {
		method += "async "
	}
	method += fmt.Sprintf("def publish_%s(self, ctx, %sreq):
", op.Name, args)
	method += g.generateDocString(docstr, tabtab)
	method += tabtab
	if getAsyncOpt(g.Options) == asyncio {
		method += "await "
	}
	method += fmt.Sprintf("self._methods['publish_%s']([ctx, %sreq])

", op.Name, args)

	method += tab
	if getAsyncOpt(g.Options) == asyncio {
		method += "async "
	}
	method += fmt.Sprintf("def _publish_%s(self, ctx, %sreq):
", op.Name, args)
	method += tabtab + fmt.Sprintf("op = '%s'
", op.Name)
	method += tabtab + fmt.Sprintf("prefix = %s
", generatePrefixStringTemplate(scope))
	method += tabtab + fmt.Sprintf("topic = '{}%s{}{}'.format(prefix, self._DELIMITER, op)
", scope.Name)
	method += tabtab + "oprot = self._protocol
"
	method += tabtab
	if getAsyncOpt(g.Options) == asyncio {
		method += "await "
	}
	method += "self._transport.lock_topic(topic)
"
	method += tabtab + "try:
"
	method += tabtabtab + "oprot.write_request_headers(ctx)
"
	method += tabtabtab + "oprot.writeMessageBegin(op, TMessageType.CALL, 0)
"
	method += tabtabtab + "req.write(oprot)
"
	method += tabtabtab + "oprot.writeMessageEnd()
"
	method += tabtabtab
	if getAsyncOpt(g.Options) == asyncio {
		method += "await "
	}
	method += "oprot.get_transport().flush()
"
	method += tabtab + "finally:
"
	method += tabtabtab + "self._transport.unlock_topic()
"
	return method
}

func generatePrefixStringTemplate(scope *parser.Scope) string {
	if len(scope.Prefix.Variables) == 0 {
		if scope.Prefix.String == "" {
			return "''"
		}
		return fmt.Sprintf("'%s%s'", scope.Prefix.String, globals.TopicDelimiter)
	}
	template := fmt.Sprintf("'%s%s'.format(", scope.Prefix.Template("{}"), globals.TopicDelimiter)
	prefix := ""
	for _, variable := range scope.Prefix.Variables {
		template += prefix + variable
		prefix = ", "
	}
	template += ")"
	return template
}

// GenerateSubscriber generates the subscriber for the given scope.
func (g *Generator) GenerateSubscriber(file *os.File, scope *parser.Scope) error {
	// TODO
	globals.PrintWarning(fmt.Sprintf("%s: scope subscriber generation is not implemented for Python", scope.Name))
	return nil
}

// GenerateService generates the given service.
func (g *Generator) GenerateService(file *os.File, s *parser.Service) error {
	// TODO
	globals.PrintWarning(fmt.Sprintf("%s: service generation is not implemented for Python", s.Name))
	return nil
}

func (g *Generator) generateServiceInterface(service *parser.Service) string {
	contents := ""
	if service.Extends != "" {
		contents += fmt.Sprintf("class Iface(%s.Iface):
", g.getServiceExtendsName(service))
	} else {
		contents += "class Iface(object):
"
	}
	if service.Comment != nil {
		contents += g.generateDocString(service.Comment, tab)
	}
	contents += "\n"

	for _, method := range service.Methods {
		contents += g.generateMethodSignature(method)
		contents += tabtab + "pass

"
	}

	return contents
}

func (g *Generator) getServiceExtendsName(service *parser.Service) string {
	serviceName := "f_" + service.ExtendsService()
	include := service.ExtendsInclude()
	if include != "" {
		if inc, ok := g.Frugal.NamespaceForInclude(include, lang); ok {
			include = inc
		}
		serviceName = include + "." + serviceName
	}
	return serviceName
}

func (g *Generator) generateProcessor(service *parser.Service) string {
	contents := ""
	if service.Extends != "" {
		contents += fmt.Sprintf("class Processor(%s.Processor):

", g.getServiceExtendsName(service))
	} else {
		contents += "class Processor(FBaseProcessor):

"
	}

	contents += tab + "def __init__(self, handler):
"
	contents += g.generateDocString([]string{
		"Create a new Processor.
",
		"Args:",
		tab + "handler: Iface",
	}, tabtab)
	if service.Extends != "" {
		contents += tabtab + "super(Processor, self).__init__(handler)
"
	} else {
		contents += tabtab + "super(Processor, self).__init__()
"
	}
	for _, method := range service.Methods {
		contents += tabtab + fmt.Sprintf("self.add_to_processor_map('%s', _%s(handler, self.get_write_lock()))
",
			method.Name, method.Name)
	}
	contents += "\n\n"
	return contents
}

func (g *Generator) generateMethodSignature(method *parser.Method) string {
	contents := ""
	docstr := []string{"Args:", tab + "ctx: FContext"}
	for _, arg := range method.Arguments {
		docstr = append(docstr, tab+fmt.Sprintf("%s: %s", arg.Name, g.getPythonTypeName(arg.Type)))
	}
	if method.Comment != nil {
		docstr[0] = "\n" + tabtab + docstr[0]
		docstr = append(method.Comment, docstr...)
	}
	contents += tab
	if getAsyncOpt(g.Options) == asyncio {
		contents += "async "
	}
	contents += fmt.Sprintf("def %s(self, ctx%s):
", method.Name, g.generateClientArgs(method.Arguments))
	contents += g.generateDocString(docstr, tabtab)
	return contents
}

func (g *Generator) generateClientArgs(args []*parser.Field) string {
	return g.generateArgs(args, "")
}

func (g *Generator) generateServerArgs(args []*parser.Field) string {
	return g.generateArgs(args, "args.")
}

func (g *Generator) generateArgs(args []*parser.Field, prefix string) string {
	argsStr := ""
	for _, arg := range args {
		argsStr += fmt.Sprintf(", %s%s", prefix, arg.Name)
	}
	return argsStr
}

func (g *Generator) generateDocString(lines []string, tab string) string {
	docstr := tab + "\"\"\"\n"
	for _, line := range lines {
		docstr += tab + line + "\n"
	}
	docstr += tab + "\"\"\"\n"
	return docstr
}

func (g *Generator) getPythonTypeName(t *parser.Type) string {
	t = g.Frugal.UnderlyingType(t)
	switch t.Name {
	case "bool":
		return "boolean"
	case "byte", "i8":
		return "int (signed 8 bits)"
	case "i16":
		return "int (signed 16 bits)"
	case "i32":
		return "int (signed 32 bits)"
	case "i64":
		return "int (signed 64 bits)"
	case "double":
		return "float"
	case "string":
		return "string"
	case "binary":
		return "binary string"
	case "list":
		typ := g.Frugal.UnderlyingType(t.ValueType)
		return fmt.Sprintf("list of %s", g.getPythonTypeName(typ))
	case "set":
		typ := g.Frugal.UnderlyingType(t.ValueType)
		return fmt.Sprintf("set of %s", g.getPythonTypeName(typ))
	case "map":
		return fmt.Sprintf("dict of <%s, %s>",
			g.getPythonTypeName(t.KeyType), g.getPythonTypeName(t.ValueType))
	default:
		// Custom type, either typedef or struct.
		return t.Name
	}
}

func (g *Generator) qualifiedTypeName(t *parser.Type) string {
	param := t.ParamName()
	include := t.IncludeName()

	if include == "" {
		return param
	}

	namespace, ok := g.Frugal.NamespaceForInclude(include, lang)
	if !ok {
		namespace = include
	}
	return fmt.Sprintf("%s.ttypes.%s", namespace, param)
}

func (g *Generator) getTType(t *parser.Type) string {
	underlyingType := g.Frugal.UnderlyingType(t)

	ttype := ""
	switch underlyingType.Name {
	case "bool", "byte", "double", "i16", "i32", "i64", "list", "set", "map", "string":
		ttype = strings.ToUpper(underlyingType.Name)
	case "binary":
		ttype = "STRING"
	default:
		if g.Frugal.IsStruct(t) {
			ttype = "STRUCT"
		} else if g.Frugal.IsEnum(t) {
			ttype = "I32"
		} else {
			panic("unrecognized type: " + underlyingType.Name)
		}
	}
	return "TType." + ttype
}

var elemNum int

// getElem returns a unique identifier name
func getElem() string {
	s := fmt.Sprintf("_elem%d", elemNum)
	elemNum++
	return s
}

func getAsyncOpt(options map[string]string) int {
	if _, ok := options["tornado"]; ok {
		return tornado
	} else if _, ok := options["asyncio"]; ok {
		return asyncio
	}
	return vanilla
}