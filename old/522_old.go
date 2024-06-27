package java

	publisher += fmt.Sprintf(tab+"private static final String DELIMITER = \"%s\";\n\n", globals.TopicDelimiter)
	publisher += fmt.Sprintf(tab+"private final Internal%sPublisher target;\n", scopeTitle)
	publisher += fmt.Sprintf(tab+"private final Internal%sPublisher proxy;\n\n", scopeTitle)
	publisher += fmt.Sprintf(tab+"public %sPublisher(FScopeProvider provider, ServiceMiddleware... middleware) {\n", scopeTitle)
	publisher += fmt.Sprintf(tabtab+"target = new Internal%sPublisher(provider);\n", scopeTitle)
	publisher += fmt.Sprintf(tabtab+"proxy = (Internal%sPublisher) InvocationHandler.composeMiddlewareClass(target, Internal%sPublisher.class, middleware);\n",
		scopeTitle, scopeTitle)
	publisher += tab + "}\n\n"
	publisher += tab + "public void open() throws TException {\n"
	publisher += tabtab + "target.open();\n"
	publisher += tab + "}\n\n"
	publisher += tab + "public void close() throws TException {\n"
	publisher += tabtab + "target.close();\n"
	publisher += tab + "}\n\n"
import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

const (
	lang                        = "java"
	defaultOutputDir            = "gen-java"
	tab                         = "\t"
	generatedAnnotations        = "generated_annotations"
	tabtab                      = tab + tab
	tabtabtab                   = tab + tab + tab
	tabtabtabtab                = tab + tab + tab + tab
	tabtabtabtabtab             = tab + tab + tab + tab + tab
	tabtabtabtabtabtab          = tab + tab + tab + tab + tab + tab
	tabtabtabtabtabtabtab       = tab + tab + tab + tab + tab + tab + tab
	tabtabtabtabtabtabtabtab    = tab + tab + tab + tab + tab + tab + tab + tab
	tabtabtabtabtabtabtabtabtab = tab + tab + tab + tab + tab + tab + tab + tab + tab
)

type Generator struct {
	*generator.BaseGenerator
	time      time.Time
	outputDir string
}

func NewGenerator(options map[string]string) generator.LanguageGenerator {
	return &Generator{
		&generator.BaseGenerator{Options: options},
		globals.Now,
		"",
	}
}

// ADTs would be really nice
type IsSetType int64

const (
	IsSetNone IsSetType = iota
	IsSetBitfield
	IsSetBitSet
)

// This is how java does isset checks, I'm open to changing this.
func (g *Generator) getIsSetType(s *parser.Struct) (IsSetType, string) {
	primitiveCount := 0
	for _, field := range s.Fields {
		if g.isJavaPrimitive(field.Type) {
			primitiveCount += 1
		}
	}

	switch {
	case primitiveCount == 0:
		return IsSetNone, ""
	case 0 < primitiveCount && primitiveCount <= 8:
		return IsSetBitfield, "byte"
	case 8 < primitiveCount && primitiveCount <= 16:
		return IsSetBitfield, "short"
	case 16 < primitiveCount && primitiveCount <= 32:
		return IsSetBitfield, "int"
	case 32 < primitiveCount && primitiveCount <= 64:
		return IsSetBitfield, "long"
	default:
		return IsSetBitSet, ""
	}
}

func (g *Generator) SetupGenerator(outputDir string) error {
	g.outputDir = outputDir
	return nil
}

func (g *Generator) TeardownGenerator() error {
	return nil
}

func (g *Generator) GenerateConstantsContents(constants []*parser.Constant) error {
	if len(constants) == 0 {
		return nil
	}

	contents := ""

	if g.includeGeneratedAnnotation() {
		contents += g.generatedAnnotation()
	}
	contents += fmt.Sprintf("public class %sConstants {\n", g.Frugal.Name)

	for _, constant := range constants {
		val := g.generateConstantValueWrapper(constant.Name, constant.Type, constant.Value, true, true)
		contents += fmt.Sprintf("%s\n", val)
	}

	contents += "}\n"

	file, err := g.GenerateFile(fmt.Sprintf("%sConstants", g.Frugal.Name), g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.initStructFile(file); err != nil {
		return err
	}
	_, err = file.WriteString(contents)

	return err
}

// generateConstantValueWrapper generates a constant value. Unlike other languages,
// constants can't be initialized in a single expression, so temp variables
// are needed. Due to this, the entire constant, not just the value, is
// generated.
func (g *Generator) generateConstantValueWrapper(fieldName string, t *parser.Type, value interface{}, declare, needsStatic bool) string {
	underlyingType := g.Frugal.UnderlyingType(t)
	contents := tabtab

	if needsStatic {
		contents += "public static final "
	}
	if declare {
		contents += g.getJavaTypeFromThriftType(underlyingType) + " "
	}

	if value == nil {
		return fmt.Sprintf("%s%s = %s;\n", contents, fieldName, "null")
	}

	if underlyingType.IsPrimitive() || g.Frugal.IsEnum(underlyingType) {
		_, val := g.generateConstantValueRec(t, value)
		return fmt.Sprintf("%s%s = %s;\n", contents, fieldName, val)
	} else if g.Frugal.IsStruct(underlyingType) {
		var s *parser.Struct
		for _, potential := range g.Frugal.Thrift.Structs {
			if underlyingType.Name == potential.Name {
				s = potential
				break
			}
		}

		contents += fmt.Sprintf("%s = new %s();\n", fieldName, g.getJavaTypeFromThriftType(underlyingType))

		ind := tabtab
		if needsStatic {
			contents += ind + "static {\n"
			ind += tab
		}

		for _, pair := range value.([]parser.KeyValue) {
			name := pair.Key.(string)
			for _, field := range s.Fields {
				if name == field.Name {
					preamble, val := g.generateConstantValueRec(field.Type, pair.Value)
					contents += preamble
					contents += fmt.Sprintf(ind+"%s.set%s(%s);\n", fieldName, strings.Title(name), val)
				}
			}
		}

		if needsStatic {
			contents += tabtab + "}\n"
		}

		return contents
	} else if underlyingType.IsContainer() {
		switch underlyingType.Name {
		case "list":
			contents += fmt.Sprintf("%s = new ArrayList<%s>();\n", fieldName, containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
			ind := tabtab
			if needsStatic {
				contents += ind + "static {\n"
				ind += tab
			}

			for _, v := range value.([]interface{}) {
				preamble, val := g.generateConstantValueRec(underlyingType.ValueType, v)
				contents += preamble
				contents += fmt.Sprintf(ind+"%s.add(%s);\n", fieldName, val)
			}

			if needsStatic {
				contents += tabtab + "}\n"
			}
			return contents
		case "set":
			contents += fmt.Sprintf("%s = new HashSet<%s>();\n", fieldName, containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
			ind := tabtab
			if needsStatic {
				contents += ind + "static {\n"
				ind += tab
			}

			for _, v := range value.([]interface{}) {
				preamble, val := g.generateConstantValueRec(underlyingType.ValueType, v)
				contents += preamble
				contents += fmt.Sprintf(ind+"%s.add(%s);\n", fieldName, val)
			}

			if needsStatic {
				contents += tabtab + "}\n"
			}
			return contents
		case "map":
			contents += fmt.Sprintf("%s = new HashMap<%s,%s>();\n",
				fieldName, containerType(g.getJavaTypeFromThriftType(underlyingType.KeyType)),
				containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
			ind := tabtab
			if needsStatic {
				contents += tabtab + "static {\n"
				ind += tab
			}

			for _, pair := range value.([]parser.KeyValue) {
				preamble, key := g.generateConstantValueRec(underlyingType.KeyType, pair.Key)
				contents += preamble
				preamble, val := g.generateConstantValueRec(underlyingType.ValueType, pair.Value)
				contents += preamble
				contents += fmt.Sprintf(ind+"%s.put(%s, %s);\n", fieldName, key, val)
			}

			if needsStatic {
				contents += tabtab + "}\n"
			}
			return contents
		}
	}

	panic("Unrecognized type: " + underlyingType.Name)
}

func (g *Generator) generateEnumConstValue(frugal *parser.Frugal, pieces []string, t *parser.Type) (string, bool) {
	for _, enum := range frugal.Thrift.Enums {
		if pieces[0] == enum.Name {
			for _, value := range enum.Values {
				if pieces[1] == value.Name {
					return fmt.Sprintf("%s.%s", g.getJavaTypeFromThriftType(t), value.Name), true
				}
			}
			panic(fmt.Sprintf("referenced value '%s' of enum '%s' doesn't exist", pieces[1], pieces[0]))
		}
	}
	return "", false
}

func (g *Generator) generateEnumConstFromValue(t *parser.Type, value int) string {
	frugal := g.Frugal
	if t.IncludeName() != "" {
		// The type is from an include
		frugal = g.Frugal.ParsedIncludes[t.IncludeName()]
	}

	for _, enum := range frugal.Thrift.Enums {
		if enum.Name == t.ParamName() {
			// found the enum
			for _, enumValue := range enum.Values {
				if enumValue.Value == value {
					// found the value
					return fmt.Sprintf("%s.%s", g.getJavaTypeFromThriftType(t), enumValue.Name)
				}
			}
		}
	}

	panic("value not found")
}

func (g *Generator) generateConstantValueRec(t *parser.Type, value interface{}) (string, string) {
	underlyingType := g.Frugal.UnderlyingType(t)

	// TODO consolidate this between generators
	// If the value being referenced is of type Identifier, it's referencing
	// another constant. Need to recurse to get that value.
	identifier, ok := value.(parser.Identifier)
	if ok {
		name := string(identifier)

		// split based on '.', if present, it should be from an include
		pieces := strings.Split(name, ".")
		if len(pieces) == 1 {
			// From this file
			for _, constant := range g.Frugal.Thrift.Constants {
				if name == constant.Name {
					return g.generateConstantValueRec(t, constant.Value)
				}
			}
		} else if len(pieces) == 2 {
			// Either from an include, or part of an enum
			val, ok := g.generateEnumConstValue(g.Frugal, pieces, underlyingType)
			if ok {
				return "", val
			}

			// If not part of an enum, it's from an include
			include, ok := g.Frugal.ParsedIncludes[pieces[0]]
			if !ok {
				panic(fmt.Sprintf("referenced include '%s' in constant '%s' not present", pieces[0], name))
			}
			for _, constant := range include.Thrift.Constants {
				if pieces[1] == constant.Name {
					return g.generateConstantValueRec(t, constant.Value)
				}
			}
		} else if len(pieces) == 3 {
			// enum from an include
			include, ok := g.Frugal.ParsedIncludes[pieces[0]]
			if !ok {
				panic(fmt.Sprintf("referenced include '%s' in constant '%s' not present", pieces[0], name))
			}

			val, ok := g.generateEnumConstValue(include, pieces[1:], underlyingType)
			if ok {
				return "", val
			}
		}

		panic("referenced constant doesn't exist: " + name)
	}

	if underlyingType.IsPrimitive() {
		switch underlyingType.Name {
		case "bool":
			return "", fmt.Sprintf("%v", value)
		case "byte", "i8":
			return "", fmt.Sprintf("(byte)%v", value)
		case "i16":
			return "", fmt.Sprintf("(short)%v", value)
		case "i32":
			return "", fmt.Sprintf("%v", value)
		case "i64":
			return "", fmt.Sprintf("%vL", value)
		case "double":
			return "", fmt.Sprintf("%v", value)
		case "string":
			return "", fmt.Sprintf("%v", strconv.Quote(value.(string)))
		case "binary":
			return "", fmt.Sprintf("java.nio.ByteBuffer.wrap(\"%v\".getBytes())", value)
		}
	} else if g.Frugal.IsEnum(underlyingType) {
		return "", g.generateEnumConstFromValue(underlyingType, int(value.(int64)))
	}
	elem := g.GetElem()
	preamble := g.generateConstantValueWrapper(elem, t, value, true, false)
	return preamble, elem

}

func (g *Generator) GenerateTypeDef(*parser.TypeDef) error {
	// No typedefs in java
	return nil
}

func (g *Generator) GenerateEnum(enum *parser.Enum) error {
	contents := ""
	contents += fmt.Sprintf("public enum %s implements org.apache.thrift.TEnum {\n", enum.Name)
	for idx, value := range enum.Values {
		terminator := ","
		if idx == len(enum.Values)-1 {
			terminator = ";"
		}
		contents += fmt.Sprintf(tab+"%s(%d)%s\n", value.Name, value.Value, terminator)
	}
	contents += "\n"

	contents += tab + "private final int value;\n\n"
	contents += fmt.Sprintf(tab+"private %s(int value) {\n", enum.Name)
	contents += tabtab + "this.value = value;\n"
	contents += tab + "}\n\n"

	contents += tab + "public int getValue() {\n"
	contents += tabtab + "return value;\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"public static %s findByValue(int value) {\n", enum.Name)
	contents += tabtab + "switch (value) {\n"
	for _, value := range enum.Values {
		contents += fmt.Sprintf(tabtabtab+"case %d:\n", value.Value)
		contents += fmt.Sprintf(tabtabtabtab+"return %s;\n", value.Name)
	}
	contents += tabtabtab + "default:\n"
	contents += tabtabtabtab + "return null;\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n"

	contents += "}\n"

	file, err := g.GenerateFile(enum.Name, g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.GenerateDocStringComment(file); err != nil {
		return err
	}
	if _, err = file.WriteString("\n"); err != nil {
		return err
	}
	if err = g.generatePackage(file); err != nil {
		return err
	}
	if _, err = file.WriteString("\n\n"); err != nil {
		return err
	}
	if err = g.GenerateEnumImports(file); err != nil {
		return err
	}

	_, err = file.WriteString(contents)

	return err
}

func (g *Generator) initStructFile(file *os.File) error {
	if err := g.GenerateDocStringComment(file); err != nil {
		return err
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}
	if err := g.generatePackage(file); err != nil {
		return err
	}

	if _, err := file.WriteString("\n\n"); err != nil {
		return err
	}

	if err := g.GenerateStructImports(file); err != nil {
		return err
	}

	return nil
}

func (g *Generator) GenerateStruct(s *parser.Struct) error {
	file, err := g.GenerateFile(s.Name, g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.initStructFile(file); err != nil {
		return err
	}

	_, err = file.WriteString(g.generateStruct(s, false, false))
	return err
}

func (g *Generator) GenerateUnion(union *parser.Struct) error {
	// I have no idea why java uses this convention as the fields really
	// should be optional...
	for _, field := range union.Fields {
		field.Modifier = parser.Default
	}

	file, err := g.GenerateFile(union.Name, g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.initStructFile(file); err != nil {
		return err
	}

	contents := g.generateUnion(union, false, false)
	_, err = file.WriteString(contents)
	return err
}

func (g *Generator) generateUnion(union *parser.Struct, isArg, isResult bool) string {
	contents := ""

	if g.includeGeneratedAnnotation() && !isArg && !isResult {
		contents += g.generatedAnnotation()
	}

	static := ""
	if isArg || isResult {
		static = "static "
	}
	contents += fmt.Sprintf("public %sclass %s extends org.apache.thrift.TUnion<%s, %s._Fields> {\n",
		static, union.Name, union.Name, union.Name)

	contents += g.generateDescriptors(union)
	contents += g.generateFieldsEnum(union)
	contents += g.generateMetaDataMap(union)
	contents += g.generateUnionConstructors(union)
	contents += g.generateUnionFieldConstructors(union)
	contents += g.generateUnionCheckType(union)

	contents += g.generateUnionStandardRead(union)
	contents += g.generateUnionStandardWrite(union)
	contents += g.generateUnionTupleRead(union)
	contents += g.generateUnionTupleWrite(union)

	contents += g.generateUnionGetDescriptors(union)
	contents += g.generateUnionFieldForId()
	contents += g.generateUnionGetSetFields(union)
	contents += g.generateUnionIsSetFields(union)

	contents += g.generateUnionEquals(union)
	contents += g.generateUnionCompareTo(union)
	contents += g.generateUnionHashCode(union)
	contents += g.generateWriteObject(union)
	contents += g.generateReadObject(union)

	contents += "}\n"
	return contents
}

func (g *Generator) generateUnionConstructors(union *parser.Struct) string {
	contents := ""
	contents += fmt.Sprintf(tab+"public %s() {\n", union.Name)
	contents += tabtab + "super();\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"public %s(_Fields setField, Object value) {\n", union.Name)
	contents += tabtab + "super(setField, value);\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"public %s(%s other) {\n", union.Name, union.Name)
	contents += tabtab + "super(other);\n"
	contents += tab + "}\n"

	contents += fmt.Sprintf(tab+"public %s deepCopy() {\n", union.Name)
	contents += fmt.Sprintf(tabtab+"return new %s(this);\n", union.Name)
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionFieldConstructors(union *parser.Struct) string {
	contents := ""

	for _, field := range union.Fields {
		contents += fmt.Sprintf(tab+"public static %s %s(%s value) {\n",
			union.Name, field.Name, g.getJavaTypeFromThriftType(field.Type))
		contents += fmt.Sprintf(tabtab+"%s x = new %s();\n", union.Name, union.Name)
		contents += fmt.Sprintf(tabtab+"x.set%s(value);\n", strings.Title(field.Name))
		contents += tabtab + "return x;\n"
		contents += tab + "}\n\n"
	}

	return contents
}

func (g *Generator) generateUnionCheckType(union *parser.Struct) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += tab + "protected void checkType(_Fields setField, Object value) throws ClassCastException {\n"
	contents += tabtab + "switch (setField) {\n"
	for _, field := range union.Fields {
		fieldType := containerType(g.getJavaTypeFromThriftType(field.Type))
		unparametrizedType := containerType(g.getUnparametrizedJavaType(field.Type))
		contents += fmt.Sprintf(tabtabtab+"case %s:\n", toConstantName(field.Name))
		contents += fmt.Sprintf(tabtabtabtab+"if (value instanceof %s) {\n", unparametrizedType)
		contents += tabtabtabtabtab + "break;\n"
		contents += tabtabtabtab + "}\n"
		contents += fmt.Sprintf(tabtabtabtab+"throw new ClassCastException(\"Was expecting value of type %s for field '%s', but got \" + value.getClass().getSimpleName());\n",
			fieldType, field.Name)
	}
	contents += tabtabtab + "default:\n"
	contents += fmt.Sprintf(tabtabtabtab + "throw new IllegalArgumentException(\"Unknown field id \" + setField);\n")
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionStandardRead(union *parser.Struct) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += tab + "protected Object standardSchemeReadValue(org.apache.thrift.protocol.TProtocol iprot, org.apache.thrift.protocol.TField field) throws org.apache.thrift.TException {\n"
	contents += tabtab + "_Fields setField = _Fields.findByThriftId(field.id);\n"
	contents += tabtab + "if (setField != null) {\n"
	contents += tabtabtab + "switch (setField) {\n"
	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		contents += fmt.Sprintf(tabtabtabtab+"case %s:\n", constantName)
		contents += fmt.Sprintf(tabtabtabtabtab+"if (field.type == %s_FIELD_DESC.type) {\n", constantName)
		contents += g.generateReadFieldRec(field, false, false, true, tabtabtabtabtabtab)
		contents += fmt.Sprintf(tabtabtabtabtabtab+"return %s;\n", field.Name)
		contents += tabtabtabtabtab + "} else {\n"
		contents += tabtabtabtabtabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, field.type);\n"
		contents += tabtabtabtabtabtab + "return null;\n"
		contents += tabtabtabtabtab + "}\n"
	}
	contents += tabtabtabtab + "default:\n"
	contents += tabtabtabtabtab + "throw new IllegalStateException(\"setField wasn't null, but didn't match any of the case statements!\");\n"
	contents += tabtabtab + "}\n"
	contents += tabtab + "} else {\n"
	contents += tabtabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, field.type);\n"
	contents += tabtabtab + "return null;\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionStandardWrite(union *parser.Struct) string {
	return g.generateUnionWrite(union, "standard")
}

func (g *Generator) generateUnionTupleRead(union *parser.Struct) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += tab + "protected Object tupleSchemeReadValue(org.apache.thrift.protocol.TProtocol iprot, short fieldID) throws org.apache.thrift.TException {\n"
	contents += tabtab + "_Fields setField = _Fields.findByThriftId(fieldID);\n"
	contents += tabtab + "if (setField != null) {\n"
	contents += tabtabtab + "switch (setField) {\n"
	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		contents += fmt.Sprintf(tabtabtabtab+"case %s:\n", constantName)
		contents += g.generateReadFieldRec(field, false, false, true, tabtabtabtabtab)
		contents += fmt.Sprintf(tabtabtabtabtab+"return %s;\n", field.Name)
	}
	contents += tabtabtabtab + "default:\n"
	contents += tabtabtabtabtab + "throw new IllegalStateException(\"setField wasn't null, but didn't match any of the case statements!\");\n"
	contents += tabtabtab + "}\n"
	contents += tabtab + "} else {\n"
	contents += tabtabtab + "throw new TProtocolException(\"Couldn't find a field with field id \" + fieldID);
"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionTupleWrite(union *parser.Struct) string {
	return g.generateUnionWrite(union, "tuple")
}

func (g *Generator) generateUnionWrite(union *parser.Struct, scheme string) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += fmt.Sprintf(tab+"protected void %sSchemeWriteValue(org.apache.thrift.protocol.TProtocol oprot) throws org.apache.thrift.TException {
", scheme)
	contents += tabtab + "switch (setField_) {
"
	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		javaContainerType := containerType(g.getJavaTypeFromThriftType(field.Type))
		contents += fmt.Sprintf(tabtabtab+"case %s:\n", constantName)
		contents += fmt.Sprintf(tabtabtabtab+"%s %s = (%s)value_;
", javaContainerType, field.Name, javaContainerType)
		contents += g.generateWriteFieldRec(field, false, false, tabtabtabtab)
		contents += tabtabtabtab + "return;
"
	}
	contents += tabtabtab + "default:\n"
	contents += tabtabtabtab + "throw new IllegalStateException(\"Cannot write union with unknown field \" + setField_);
"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionGetDescriptors(union *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += tab + "protected org.apache.thrift.protocol.TField getFieldDesc(_Fields setField) {
"
	contents += tabtab + "switch (setField) {\n"

	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		contents += fmt.Sprintf(tabtabtab+"case %s:\n", constantName)
		contents += fmt.Sprintf(tabtabtabtab+"return %s_FIELD_DESC;
", constantName)
	}

	contents += tabtabtab + "default:\n"
	contents += tabtabtabtab + "throw new IllegalArgumentException(\"Unknown field id \" + setField);\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	contents += tab + "@Override\n"
	contents += tab + "protected org.apache.thrift.protocol.TStruct getStructDesc() {
"
	contents += tabtab + "return STRUCT_DESC;
"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateUnionFieldForId() string {
	contents := ""
	contents += tab + "@Override\n"
	contents += tab + "protected _Fields enumForId(short id) {
"
	contents += tabtab + "return _Fields.findByThriftIdOrThrow(id);
"
	contents += tab + "}\n\n"

	contents += tab + "public _Fields fieldForId(int fieldId) {
"
	contents += tabtab + "return _Fields.findByThriftId(fieldId);
"
	contents += tab + "}


"
	return contents
}

func (g *Generator) generateUnionGetSetFields(union *parser.Struct) string {
	contents := ""

	for _, field := range union.Fields {
		titleName := strings.Title(field.Name)
		constantName := toConstantName(field.Name)
		javaType := g.getJavaTypeFromThriftType(field.Type)

		// get
		contents += fmt.Sprintf(tab+"public %s get%s() {
", javaType, titleName)
		contents += fmt.Sprintf(tabtab+"if (getSetField() == _Fields.%s) {
", constantName)
		contents += fmt.Sprintf(tabtabtab+"return (%s)getFieldValue();
", containerType(javaType))
		contents += tabtab + "} else {\n"
		contents += fmt.Sprintf(tabtabtab+"throw new RuntimeException(\"Cannot get field '%s' because union is currently set to \" + getFieldDesc(getSetField()).name);
", field.Name)
		contents += tabtab + "}\n"
		contents += tab + "}\n\n"

		// set
		contents += fmt.Sprintf(tab+"public void set%s(%s value) {
", titleName, javaType)
		if !g.isJavaPrimitive(field.Type) {
			contents += tabtab + "if (value == null) throw new NullPointerException();
"
		}
		contents += fmt.Sprintf(tabtab+"setField_ = _Fields.%s;
", constantName)
		contents += tabtab + "value_ = value;
"
		contents += tab + "}\n\n"
	}

	return contents
}

func (g *Generator) generateUnionIsSetFields(union *parser.Struct) string {
	contents := ""

	for _, field := range union.Fields {
		contents += fmt.Sprintf(tab+"public boolean isSet%s() {
", strings.Title(field.Name))
		contents += fmt.Sprintf(tabtab+"return setField_ == _Fields.%s;
", toConstantName(field.Name))
		contents += tab + "}\n\n"
	}

	return contents
}

func (g *Generator) generateUnionEquals(union *parser.Struct) string {
	contents := "\n"

	contents += tab + "public boolean equals(Object other) {
"
	contents += fmt.Sprintf(tabtab+"if (other instanceof %s) {
", union.Name)
	contents += fmt.Sprintf(tabtabtab+"return equals((%s)other);
", union.Name)
	contents += tabtab + "} else {\n"
	contents += tabtabtab + "return false;
"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"public boolean equals(%s other) {
", union.Name)
	contents += tabtab + "return other != null && getSetField() == other.getSetField() && getFieldValue().equals(other.getFieldValue());
"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionCompareTo(union *parser.Struct) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += fmt.Sprintf(tab+"public int compareTo(%s other) {
", union.Name)
	contents += tabtab + "int lastComparison = org.apache.thrift.TBaseHelper.compareTo(getSetField(), other.getSetField());
"
	contents += tabtab + "if (lastComparison == 0) {
"
	contents += tabtabtab + "return org.apache.thrift.TBaseHelper.compareTo(getFieldValue(), other.getFieldValue());
"
	contents += tabtab + "}\n"
	contents += tabtab + "return lastComparison;
"
	contents += tab + "}


"

	return contents
}

func (g *Generator) generateUnionHashCode(union *parser.Struct) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += tab + "public int hashCode() {
"
	contents += tabtab + "List<Object> list = new ArrayList<Object>();
"
	contents += tabtab + "list.add(this.getClass().getName());
"
	contents += tabtab + "org.apache.thrift.TFieldIdEnum setField = getSetField();
"
	contents += tabtab + "if (setField != null) {\n"
	contents += tabtabtab + "list.add(setField.getThriftFieldId());
"
	contents += tabtabtab + "Object value = getFieldValue();
"
	contents += tabtabtab + "if (value instanceof org.apache.thrift.TEnum) {
"
	contents += tabtabtabtab + "list.add(((org.apache.thrift.TEnum)getFieldValue()).getValue());
"
	contents += tabtabtab + "} else {\n"
	contents += tabtabtabtab + "list.add(value);
"
	contents += tabtabtab + "}\n"
	contents += tabtab + "}\n"
	contents += tabtab + "return list.hashCode();
"
	contents += tab + "}\n"

	return contents
}

func (g *Generator) GenerateException(exception *parser.Struct) error {
	return g.GenerateStruct(exception)
}

func (g *Generator) GenerateServiceArgsResults(serviceName string, outputDir string, structs []*parser.Struct) error {
	contents := ""
	if g.includeGeneratedAnnotation() {
		contents += g.generatedAnnotation()
	}
	contents += fmt.Sprintf("public class %s {
", serviceName)

	for _, s := range structs {
		for _, field := range s.Fields {
			if field.Modifier == parser.Optional {
				field.Modifier = parser.Default
			}
		}
		contents += g.generateStruct(s, strings.HasSuffix(s.Name, "_args"), strings.HasSuffix(s.Name, "_result"))
	}

	contents += "}\n"

	file, err := g.GenerateFile(strings.Title(serviceName), g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.initStructFile(file); err != nil {
		return err
	}
	if _, err = file.WriteString(contents); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateStruct(s *parser.Struct, isArg, isResult bool) string {
	contents := ""

	if s.Comment != nil {
		contents += g.GenerateBlockComment(s.Comment, "")
	}
	if g.includeGeneratedAnnotation() && !isArg && !isResult {
		contents += g.generatedAnnotation()
	}
	static := ""
	if isArg || isResult {
		static = "static "
	}
	exception := ""
	if s.Type == parser.StructTypeException {
		exception = "extends TException "
	}
	contents += fmt.Sprintf("public %sclass %s %simplements org.apache.thrift.TBase<%s, %s._Fields>, java.io.Serializable, Cloneable, Comparable<%s> {
",
		static, s.Name, exception, s.Name, s.Name, s.Name)

	contents += g.generateDescriptors(s)

	contents += g.generateSchemeMap(s)

	contents += g.generateInstanceVars(s)

	contents += g.generateFieldsEnum(s)

	contents += g.generateIsSetVars(s)

	contents += g.generateOptionals(s)
	contents += g.generateMetaDataMap(s)

	contents += g.generateDefaultConstructor(s)
	contents += g.generateFullConstructor(s)
	contents += g.generateCopyConstructor(s)
	contents += g.generateDeepCopyMethod(s)
	contents += g.generateClear(s)

	for _, field := range s.Fields {
		underlyingType := g.Frugal.UnderlyingType(field.Type)
		if underlyingType.IsContainer() {
			contents += g.generateContainerGetSize(field)
			contents += g.generateContainerIterator(field)
			contents += g.generateContainerAddTo(field)

		}

		contents += g.generateGetField(field)
		contents += g.generateSetField(s.Name, field)
		contents += g.generateUnsetField(s, field)
		contents += g.generateIsSetField(s, field)
		contents += g.generateSetIsSetField(s, field)
	}

	contents += g.generateSetValue(s)
	contents += g.generateGetValue(s)
	contents += g.generateIsSetValue(s)

	contents += g.generateEquals(s)
	contents += g.generateHashCode(s)
	contents += g.generateCompareTo(s)

	contents += g.generateFieldForId(s)
	contents += g.generateReadWrite(s)

	contents += g.generateToString(s)
	contents += g.generateValidate(s)

	contents += g.generateWriteObject(s)
	contents += g.generateReadObject(s)

	contents += g.generateStandardScheme(s, isResult)
	contents += g.generateTupleScheme(s)

	contents += "}\n"
	return contents
}

func (g *Generator) generateDescriptors(s *parser.Struct) string {
	contents := ""
	contents += fmt.Sprintf(tab+"private static final org.apache.thrift.protocol.TStruct STRUCT_DESC = new org.apache.thrift.protocol.TStruct(\"%s\");

",
		s.Name)
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tab+"private static final org.apache.thrift.protocol.TField %s_FIELD_DESC = new org.apache.thrift.protocol.TField(\"%s\", %s, (short)%d);
",
			toConstantName(field.Name), field.Name, g.getTType(field.Type), field.ID)
	}
	contents += "\n"
	return contents
}

func (g *Generator) generateSchemeMap(s *parser.Struct) string {
	contents := ""
	contents += tab + "private static final Map<Class<? extends IScheme>, SchemeFactory> schemes = new HashMap<Class<? extends IScheme>, SchemeFactory>();
"
	contents += tab + "static {\n"
	contents += fmt.Sprintf(tabtab+"schemes.put(StandardScheme.class, new %sStandardSchemeFactory());
", s.Name)
	contents += fmt.Sprintf(tabtab+"schemes.put(TupleScheme.class, new %sTupleSchemeFactory());
", s.Name)
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateInstanceVars(s *parser.Struct) string {
	contents := ""
	for _, field := range s.Fields {
		if field.Comment != nil {
			contents += g.GenerateBlockComment(field.Comment, tab)
		}
		modifier := "required"
		if field.Modifier == parser.Optional {
			modifier = "optional"
		}
		contents += fmt.Sprintf(tab+"public %s %s; // %s
",
			g.getJavaTypeFromThriftType(field.Type), field.Name, modifier)
	}
	return contents
}

func (g *Generator) generateFieldsEnum(s *parser.Struct) string {
	contents := ""
	contents += tab + "/** The set of fields this struct contains, along with convenience methods for finding and manipulating them. */
"
	contents += tab + "public enum _Fields implements org.apache.thrift.TFieldIdEnum {
"

	for idx, field := range s.Fields {
		terminator := ""
		if idx != len(s.Fields)-1 {
			terminator = ","
		}

		if field.Comment != nil {
			contents += g.GenerateBlockComment(field.Comment, tabtab)
		}
		contents += fmt.Sprintf(tabtab+"%s((short)%d, \"%s\")%s
", toConstantName(field.Name), field.ID, field.Name, terminator)
	}
	// Do it this was as the semi colon is needed no matter what
	contents += ";
"
	contents += "\n"

	contents += tabtab + "private static final Map<String, _Fields> byName = new HashMap<String, _Fields>();

"
	contents += tabtab + "static {\n"
	contents += tabtabtab + "for (_Fields field : EnumSet.allOf(_Fields.class)) {
"
	contents += tabtabtabtab + "byName.put(field.getFieldName(), field);
"
	contents += tabtabtab + "}\n"
	contents += tabtab + "}\n\n"

	contents += g.GenerateBlockComment([]string{"Find the _Fields constant that matches fieldId, or null if its not found."}, tabtab)
	contents += tabtab + "public static _Fields findByThriftId(int fieldId) {
"
	contents += tabtabtab + "switch(fieldId) {
"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtabtabtab+"case %d: // %s\n", field.ID, toConstantName(field.Name))
		contents += fmt.Sprintf(tabtabtabtabtab+"return %s;\n", toConstantName(field.Name))
	}
	contents += tabtabtabtab + "default:\n"
	contents += tabtabtabtabtab + "return null;\n"
	contents += tabtabtab + "}\n"
	contents += tabtab + "}\n\n"

	contents += g.GenerateBlockComment([]string{
		"Find the _Fields constant that matches fieldId, throwing an exception",
		"if it is not found.",
	}, tabtab)
	contents += tabtab + "public static _Fields findByThriftIdOrThrow(int fieldId) {
"
	contents += tabtabtab + "_Fields fields = findByThriftId(fieldId);
"
	contents += tabtabtab + "if (fields == null) throw new IllegalArgumentException(\"Field \" + fieldId + \" doesn't exist!\");\n"
	contents += tabtabtab + "return fields;\n"
	contents += tabtab + "}\n\n"

	contents += g.GenerateBlockComment([]string{"Find the _Fields constant that matches name, or null if its not found."}, tabtab)
	contents += tabtab + "public static _Fields findByName(String name) {\n"
	contents += tabtabtab + "return byName.get(name);\n"
	contents += tabtab + "}\n\n"

	contents += tabtab + "private final short _thriftId;\n"
	contents += tabtab + "private final String _fieldName;\n\n"

	contents += tabtab + "_Fields(short thriftId, String fieldName) {\n"
	contents += tabtabtab + "_thriftId = thriftId;\n"
	contents += tabtabtab + "_fieldName = fieldName;\n"
	contents += tabtab + "}\n\n"

	contents += tabtab + "public short getThriftFieldId() {\n"
	contents += tabtabtab + "return _thriftId;\n"
	contents += tabtab + "}\n\n"

	contents += tabtab + "public String getFieldName() {\n"
	contents += tabtabtab + "return _fieldName;\n"
	contents += tabtab + "}\n"

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateIsSetVars(s *parser.Struct) string {
	contents := ""
	contents += tab + "// isset id assignments\n"
	primitiveCount := 0
	for _, field := range s.Fields {
		if g.isJavaPrimitive(field.Type) {
			contents += fmt.Sprintf(tab+"private static final int %s = %d;\n",
				g.getIsSetID(field.Name), primitiveCount)
			primitiveCount += 1
		}
	}
	isSetType, bitFieldType := g.getIsSetType(s)
	switch isSetType {
	case IsSetNone:
	// Do nothing
	case IsSetBitfield:
		contents += fmt.Sprintf(tab+"private %s __isset_bitfield = 0;\n", bitFieldType)
	case IsSetBitSet:
		contents += fmt.Sprintf(tab+"private BitSet __isset_bit_vector = new BitSet(%d);\n", primitiveCount)
	}
	return contents
}

func (g *Generator) generateOptionals(s *parser.Struct) string {
	// TODO 2.0 These don't appear to be used by anything
	contents := ""

	optionals := ""
	sep := ""
	for _, field := range s.Fields {
		if field.Modifier != parser.Optional {
			continue
		}
		optionals += fmt.Sprintf(sep+"_Fields.%s", toConstantName(field.Name))
		sep = ","
	}

	if len(optionals) > 0 {
		contents += fmt.Sprintf(tab+"private static final _Fields optionals[] = {%s};
", optionals)
	}

	return contents
}

func (g *Generator) generateMetaDataMap(s *parser.Struct) string {
	contents := ""
	contents += tab + "public static final Map<_Fields, org.apache.thrift.meta_data.FieldMetaData> metaDataMap;
"
	contents += tab + "static {\n"
	contents += tabtab + "Map<_Fields, org.apache.thrift.meta_data.FieldMetaData> tmpMap = new EnumMap<_Fields, org.apache.thrift.meta_data.FieldMetaData>(_Fields.class);
"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtab+"tmpMap.put(_Fields.%s, new org.apache.thrift.meta_data.FieldMetaData(\"%s\", org.apache.thrift.TFieldRequirementType.%s,
",
			toConstantName(field.Name), field.Name, field.Modifier.String())
		contents += fmt.Sprintf("%s));
", g.generateMetaDataMapEntry(field.Type, tabtabtabtab))
	}
	contents += tabtab + "metaDataMap = Collections.unmodifiableMap(tmpMap);
"
	contents += fmt.Sprintf(tabtab+"org.apache.thrift.meta_data.FieldMetaData.addStructMetaDataMap(%s.class, metaDataMap);
", s.Name)
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateDefaultConstructor(s *parser.Struct) string {
	contents := ""
	contents += fmt.Sprintf(tab+"public %s() {\n", s.Name)
	for _, field := range s.Fields {
		if field.Default != nil {
			val := g.generateConstantValueWrapper("this."+field.Name, field.Type, field.Default, false, false)
			contents += fmt.Sprintf("%s
", val)
		}
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateFullConstructor(s *parser.Struct) string {
	contents := ""
	argsList := ""
	sep := "\n" + tabtab
	numNonOptional := 0
	for _, field := range s.Fields {
		if field.Modifier == parser.Optional {
			continue
		}
		argsList += fmt.Sprintf("%s%s %s", sep, g.getJavaTypeFromThriftType(field.Type), field.Name)
		sep = ",
" + tabtab
		numNonOptional += 1
	}

	if numNonOptional > 0 {
		contents += fmt.Sprintf(tab+"public %s(%s) {
", s.Name, argsList)
		contents += fmt.Sprintf(tabtab + "this();
")
		for _, field := range s.Fields {
			if field.Modifier == parser.Optional {
				continue
			}

			if g.Frugal.UnderlyingType(field.Type).Name == "binary" {
				contents += fmt.Sprintf(tabtab+"this.%s = org.apache.thrift.TBaseHelper.copyBinary(%s);
", field.Name, field.Name)
			} else {
				contents += fmt.Sprintf(tabtab+"this.%s = %s;\n", field.Name, field.Name)
			}

			if g.isJavaPrimitive(field.Type) {
				contents += fmt.Sprintf(tabtab+"set%sIsSet(true);
", strings.Title(field.Name))
			}
		}
		contents += tab + "}\n\n"
	}
	return contents
}

func (g *Generator) generateCopyConstructor(s *parser.Struct) string {
	contents := ""
	contents += g.GenerateBlockComment([]string{"Performs a deep copy on <i>other</i>."}, tab)
	contents += fmt.Sprintf(tab+"public %s(%s other) {\n", s.Name, s.Name)

	isSetType, _ := g.getIsSetType(s)
	switch isSetType {
	case IsSetNone:
		// do nothing
	case IsSetBitfield:
		contents += tabtab + "__isset_bitfield = other.__isset_bitfield;
"
	case IsSetBitSet:
		contents += tabtab + "__isset_bit_vector.clear();
"
		contents += tabtab + "__isset_bit_vector.or(other.__isset_bit_vector);
"
	}

	for _, field := range s.Fields {
		isPrimitive := g.isJavaPrimitive(g.Frugal.UnderlyingType(field.Type))
		ind := tabtab
		if !isPrimitive {
			contents += fmt.Sprintf(tabtab+"if (other.isSet%s()) {
", strings.Title(field.Name))
			ind += tab
		}
		contents += g.generateCopyConstructorField(field, "other."+field.Name, true, ind)
		if !isPrimitive {
			contents += tabtab + "}\n"
		}
	}

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateDeepCopyMethod(s *parser.Struct) string {
	contents := ""
	contents += fmt.Sprintf(tab+"public %s deepCopy() {\n", s.Name)
	contents += fmt.Sprintf(tabtab+"return new %s(this);\n", s.Name)
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateClear(s *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += fmt.Sprintf(tab + "public void clear() {
")
	for _, field := range s.Fields {
		underlyingType := g.Frugal.UnderlyingType(field.Type)

		if field.Default != nil {
			val := g.generateConstantValueWrapper("this."+field.Name, field.Type, field.Default, false, false)
			contents += fmt.Sprintf("%s
", val)
		} else if g.isJavaPrimitive(field.Type) {
			contents += fmt.Sprintf(tabtab+"set%sIsSet(false);
", strings.Title(field.Name))
			val := ""
			switch underlyingType.Name {
			case "i8", "byte", "i16", "i32", "i64":
				val = "0"
			case "double":
				val = "0.0"
			case "bool":
				val = "false"
			default:
				panic("invalid type: " + underlyingType.Name)
			}
			contents += fmt.Sprintf(tabtab+"this.%s = %s;

", field.Name, val)
		} else {
			contents += fmt.Sprintf(tabtab+"this.%s = null;

", field.Name)
		}
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateContainerGetSize(field *parser.Field) string {
	contents := ""
	contents += fmt.Sprintf(tab+"public int get%sSize() {
", strings.Title(field.Name))
	contents += fmt.Sprintf(tabtab+"return (this.%s == null) ? 0 : this.%s.size();
", field.Name, field.Name)
	contents += fmt.Sprintf(tab + "}\n\n")
	return contents
}

func (g *Generator) generateContainerIterator(field *parser.Field) string {
	underlyingType := g.Frugal.UnderlyingType(field.Type)

	// maps don't get iterators
	if underlyingType.Name == "map" {
		return ""
	}

	contents := ""
	contents += fmt.Sprintf(tab+"public java.util.Iterator<%s> get%sIterator() {\n",
		containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)), strings.Title(field.Name))
	contents += fmt.Sprintf(tabtab+"return (this.%s == null) ? null : this.%s.iterator();\n", field.Name, field.Name)
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateContainerAddTo(field *parser.Field) string {
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	valType := g.getJavaTypeFromThriftType(underlyingType.ValueType)
	fieldTitle := strings.Title(field.Name)

	contents := ""

	if underlyingType.Name == "list" || underlyingType.Name == "set" {
		contents += fmt.Sprintf(tab+"public void addTo%s(%s elem) {\n", fieldTitle, valType)
		newContainer := ""
		if underlyingType.Name == "list" {
			newContainer = fmt.Sprintf("new ArrayList<%s>()", containerType(valType))
		} else {
			newContainer = fmt.Sprintf("new HashSet<%s>()", containerType(valType))
		}
		contents += fmt.Sprintf(tabtab+"if (this.%s == null) {\n", field.Name)
		contents += fmt.Sprintf(tabtabtab+"this.%s = %s;\n", field.Name, newContainer)
		contents += tabtab + "}\n"
		contents += fmt.Sprintf(tabtab+"this.%s.add(elem);\n", field.Name)
		contents += tab + "}\n\n"
	} else {
		contents += fmt.Sprintf(tab+"public void putTo%s(%s key, %s val) {\n",
			fieldTitle, g.getJavaTypeFromThriftType(underlyingType.KeyType), valType)
		contents += fmt.Sprintf(tabtab+"if (this.%s == null) {\n", field.Name)
		contents += fmt.Sprintf(tabtabtab+"this.%s = new HashMap<%s,%s>();\n",
			field.Name, containerType(g.getJavaTypeFromThriftType(underlyingType.KeyType)), containerType(valType))
		contents += tabtab + "}\n"
		contents += fmt.Sprintf(tabtab+"this.%s.put(key, val);\n", field.Name)
		contents += tab + "}\n\n"
	}

	return contents
}

func (g *Generator) getAccessorPrefix(t *parser.Type) string {
	if g.Frugal.UnderlyingType(t).Name == "bool" {
		return "is"
	}
	return "get"
}

func (g *Generator) generateGetField(field *parser.Field) string {
	contents := ""
	fieldTitle := strings.Title(field.Name)
	if field.Comment != nil {
		contents += g.GenerateBlockComment(field.Comment, tab)
	}

	underlyingType := g.Frugal.UnderlyingType(field.Type)
	returnType := g.getJavaTypeFromThriftType(underlyingType)
	// There's some weird overlap between the ByteBuffer and byte[] methods
	if underlyingType.Name == "binary" {
		returnType = "byte[]"
	}

	accessPrefix := g.getAccessorPrefix(field.Type)
	contents += fmt.Sprintf(tab+"public %s %s%s() {
", returnType, accessPrefix, fieldTitle)
	if underlyingType.Name == "binary" {
		contents += fmt.Sprintf(tabtab+"set%s(org.apache.thrift.TBaseHelper.rightSize(%s));
",
			strings.Title(field.Name), field.Name)
		contents += fmt.Sprintf(tabtab+"return %s == null ? null : %s.array();
", field.Name, field.Name)
	} else {
		contents += fmt.Sprintf(tabtab+"return this.%s;
", field.Name)
	}
	contents += tab + "}\n\n"

	if underlyingType.Name == "binary" {
		contents += fmt.Sprintf(tab+"public java.nio.ByteBuffer bufferFor%s() {
", fieldTitle)
		contents += fmt.Sprintf(tabtab+"return org.apache.thrift.TBaseHelper.copyBinary(%s);
", field.Name)
		contents += fmt.Sprintf(tab + "}\n\n")
	}

	return contents
}

func (g *Generator) generateSetField(structName string, field *parser.Field) string {
	fieldTitle := strings.Title(field.Name)
	underlyingType := g.Frugal.UnderlyingType(field.Type)

	contents := ""

	if underlyingType.Name == "binary" {
		// Special additional binary set
		contents += fmt.Sprintf(tab+"public %s set%s(byte[] %s) {
", structName, fieldTitle, field.Name)
		contents += fmt.Sprintf(tabtab+"this.%s = %s == null ? (java.nio.ByteBuffer)null : java.nio.ByteBuffer.wrap(Arrays.copyOf(%s, %s.length));
",
			field.Name, field.Name, field.Name, field.Name)
		contents += tabtab + "return this;
"
		contents += tab + "}\n\n"
	}

	if field.Comment != nil {
		contents += g.GenerateBlockComment(field.Comment, tab)
	}
	contents += fmt.Sprintf(tab+"public %s set%s(%s %s) {
",
		structName, fieldTitle, g.getJavaTypeFromThriftType(field.Type), field.Name)

	if underlyingType.Name == "binary" {
		contents += fmt.Sprintf(tabtab+"this.%s = org.apache.thrift.TBaseHelper.copyBinary(%s);
", field.Name, field.Name)

	} else {
		contents += fmt.Sprintf(tabtab+"this.%s = %s;\n", field.Name, field.Name)
	}

	if g.isJavaPrimitive(field.Type) {
		contents += fmt.Sprintf(tabtab+"set%sIsSet(true);
", fieldTitle)
	}

	contents += tabtab + "return this;
"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnsetField(s *parser.Struct, field *parser.Field) string {
	contents := ""

	contents += fmt.Sprintf(tab+"public void unset%s() {
", strings.Title(field.Name))
	if g.isJavaPrimitive(field.Type) {
		isSetType, _ := g.getIsSetType(s)
		isSetID := g.getIsSetID(field.Name)
		switch isSetType {
		case IsSetNone:
			panic("IsSetNone occurred with a primitive")
		case IsSetBitfield:
			contents += fmt.Sprintf(tabtab+"__isset_bitfield = EncodingUtils.clearBit(__isset_bitfield, %s);
", isSetID)
		case IsSetBitSet:
			contents += fmt.Sprintf(tabtab+"__isset_bit_vector.clear(%s);
", isSetID)
		}
	} else {
		contents += fmt.Sprintf(tabtab+"this.%s = null;
", field.Name)
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) getIsSetID(fieldName string) string {
	return fmt.Sprintf("__%s_ISSET_ID", strings.ToUpper(fieldName))
}

func (g *Generator) generateIsSetField(s *parser.Struct, field *parser.Field) string {
	contents := ""
	contents += fmt.Sprintf(tab+"/** Returns true if field %s is set (has been assigned a value) and false otherwise */
", field.Name)
	contents += fmt.Sprintf(tab+"public boolean isSet%s() {
", strings.Title(field.Name))
	if g.isJavaPrimitive(field.Type) {
		isSetType, _ := g.getIsSetType(s)
		isSetID := g.getIsSetID(field.Name)
		switch isSetType {
		case IsSetNone:
			panic("IsSetNone occurred with a primitive")
		case IsSetBitfield:
			contents += fmt.Sprintf(tabtab+"return EncodingUtils.testBit(__isset_bitfield, %s);
", isSetID)
		case IsSetBitSet:
			contents += fmt.Sprintf(tabtab+"return __isset_bit_vector.get(%s);
", isSetID)
		}
	} else {
		contents += fmt.Sprintf(tabtab+"return this.%s != null;
", field.Name)
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateSetIsSetField(s *parser.Struct, field *parser.Field) string {
	contents := ""
	contents += fmt.Sprintf(tab+"public void set%sIsSet(boolean value) {
", strings.Title(field.Name))
	if g.isJavaPrimitive(field.Type) {
		isSetType, _ := g.getIsSetType(s)
		isSetID := g.getIsSetID(field.Name)
		switch isSetType {
		case IsSetNone:
			panic("IsSetNone occurred with a primitive")
		case IsSetBitfield:
			contents += fmt.Sprintf(tabtab+"__isset_bitfield = EncodingUtils.setBit(__isset_bitfield, %s, value);
", isSetID)
		case IsSetBitSet:
			contents += fmt.Sprintf(tabtab+"__isset_bit_vector.set(%s, value);
", isSetID)
		}
	} else {
		contents += tabtab + "if (!value) {
"
		contents += fmt.Sprintf(tabtabtab+"this.%s = null;
", field.Name)
		contents += tabtab + "}\n"
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateSetValue(s *parser.Struct) string {
	contents := ""
	contents += tab + "public void setFieldValue(_Fields field, Object value) {
"
	contents += tabtab + "switch (field) {
"
	for _, field := range s.Fields {
		fieldTitle := strings.Title(field.Name)
		contents += fmt.Sprintf(tabtab+"case %s:\n", toConstantName(field.Name))
		contents += tabtabtab + "if (value == null) {
"
		contents += fmt.Sprintf(tabtabtabtab+"unset%s();
", fieldTitle)
		contents += tabtabtab + "} else {\n"
		contents += fmt.Sprintf(tabtabtabtab+"set%s((%s)value);
", fieldTitle, containerType(g.getJavaTypeFromThriftType(field.Type)))
		contents += tabtabtab + "}\n"
		contents += tabtabtab + "break;

"
	}
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateGetValue(s *parser.Struct) string {
	contents := ""
	contents += tab + "public Object getFieldValue(_Fields field) {
"
	contents += tabtab + "switch (field) {
"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtab+"case %s:\n", toConstantName(field.Name))
		contents += fmt.Sprintf(tabtabtab+"return %s%s();

",
			g.getAccessorPrefix(field.Type), strings.Title(field.Name))
	}
	contents += tabtab + "}\n"
	contents += tabtab + "throw new IllegalStateException();
"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateIsSetValue(s *parser.Struct) string {
	contents := ""
	contents += tab + "/** Returns true if field corresponding to fieldID is set (has been assigned a value) and false otherwise */
"
	contents += tab + "public boolean isSet(_Fields field) {
"
	contents += tabtab + "if (field == null) {
"
	contents += tabtabtab + "throw new IllegalArgumentException();
"
	contents += tabtab + "}\n\n"

	contents += tabtab + "switch (field) {
"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtab+"case %s:\n", toConstantName(field.Name))
		contents += fmt.Sprintf(tabtabtab+"return isSet%s();
", strings.Title(field.Name))
	}
	contents += tabtab + "}\n"
	contents += tabtab + "throw new IllegalStateException();
"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateEquals(s *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += tab + "public boolean equals(Object that) {
"
	contents += tabtab + "if (that == null)
"
	contents += tabtabtab + "return false;
"
	contents += fmt.Sprintf(tabtab+"if (that instanceof %s)
", s.Name)
	contents += fmt.Sprintf(tabtabtab+"return this.equals((%s)that);
", s.Name)
	contents += tabtab + "return false;
"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"public boolean equals(%s that) {
", s.Name)
	contents += tabtab + "if (that == null)
"
	contents += tabtabtab + "return false;

"

	for _, field := range s.Fields {
		optional := field.Modifier == parser.Optional
		primitive := g.isJavaPrimitive(field.Type)

		// TODO 2.0 this looks so ugly
		thisPresentArg := "true"
		thatPresentArg := "true"
		if optional || !primitive {
			thisPresentArg += fmt.Sprintf(" && this.isSet%s()", strings.Title(field.Name))
			thatPresentArg += fmt.Sprintf(" && that.isSet%s()", strings.Title(field.Name))
		}

		contents += fmt.Sprintf(tabtab+"boolean this_present_%s = %s;
", field.Name, thisPresentArg)
		contents += fmt.Sprintf(tabtab+"boolean that_present_%s = %s;
", field.Name, thatPresentArg)
		contents += fmt.Sprintf(tabtab+"if (this_present_%s || that_present_%s) {
", field.Name, field.Name)
		contents += fmt.Sprintf(tabtabtab+"if (!(this_present_%s && that_present_%s))
", field.Name, field.Name)
		contents += tabtabtabtab + "return false;
"

		unequalTest := ""
		if primitive {
			unequalTest = fmt.Sprintf("this.%s != that.%s", field.Name, field.Name)
		} else {
			unequalTest = fmt.Sprintf("!this.%s.equals(that.%s)", field.Name, field.Name)
		}
		contents += fmt.Sprintf(tabtabtab+"if (%s)
", unequalTest)
		contents += tabtabtabtab + "return false;
"
		contents += tabtab + "}\n\n"
	}

	contents += tabtab + "return true;
"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateHashCode(s *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += tab + "public int hashCode() {
"
	contents += tabtab + "List<Object> list = new ArrayList<Object>();

"
	for _, field := range s.Fields {
		optional := field.Modifier == parser.Optional
		primitive := g.isJavaPrimitive(field.Type)

		presentArg := "true"
		if optional || !primitive {
			presentArg += fmt.Sprintf(" && (isSet%s())", strings.Title(field.Name))
		}

		contents += fmt.Sprintf(tabtab+"boolean present_%s = %s;
", field.Name, presentArg)
		contents += fmt.Sprintf(tabtab+"list.add(present_%s);
", field.Name)
		contents += fmt.Sprintf(tabtab+"if (present_%s)
", field.Name)
		if g.Frugal.IsEnum(field.Type) {
			contents += fmt.Sprintf(tabtabtab+"list.add(%s.getValue());

", field.Name)
		} else {
			contents += fmt.Sprintf(tabtabtab+"list.add(%s);

", field.Name)
		}
	}
	contents += tabtab + "return list.hashCode();
"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateCompareTo(s *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += fmt.Sprintf(tab+"public int compareTo(%s other) {
", s.Name)
	contents += tabtab + "if (!getClass().equals(other.getClass())) {
"
	contents += tabtabtab + "return getClass().getName().compareTo(other.getClass().getName());
"
	contents += tabtab + "}\n\n"
	contents += tabtab + "int lastComparison = 0;

"
	for _, field := range s.Fields {
		fieldTitle := strings.Title(field.Name)
		contents += fmt.Sprintf(tabtab+"lastComparison = Boolean.valueOf(isSet%s()).compareTo(other.isSet%s());
", fieldTitle, fieldTitle)
		contents += tabtab + "if (lastComparison != 0) {
"
		contents += tabtabtab + "return lastComparison;
"
		contents += tabtab + "}\n"
		contents += fmt.Sprintf(tabtab+"if (isSet%s()) {
", fieldTitle)
		contents += fmt.Sprintf(tabtabtab+"lastComparison = org.apache.thrift.TBaseHelper.compareTo(this.%s, other.%s);
", field.Name, field.Name)
		contents += tabtabtab + "if (lastComparison != 0) {
"
		contents += tabtabtabtab + "return lastComparison;
"
		contents += tabtabtab + "}\n"
		contents += tabtab + "}\n"
	}
	contents += tabtab + "return 0;
"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateFieldForId(s *parser.Struct) string {
	contents := ""
	contents += tab + "public _Fields fieldForId(int fieldId) {
"
	contents += tabtab + "return _Fields.findByThriftId(fieldId);
"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateReadWrite(s *parser.Struct) string {
	contents := ""
	contents += tab + "public void read(org.apache.thrift.protocol.TProtocol iprot) throws org.apache.thrift.TException {
"
	contents += tabtab + "schemes.get(iprot.getScheme()).getScheme().read(iprot, this);
"
	contents += tab + "}\n\n"

	contents += tab + "public void write(org.apache.thrift.protocol.TProtocol oprot) throws org.apache.thrift.TException {
"
	contents += tabtab + "schemes.get(oprot.getScheme()).getScheme().write(oprot, this);
"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateToString(s *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += tab + "public String toString() {
"
	contents += fmt.Sprintf(tabtab+"StringBuilder sb = new StringBuilder(\"%s(\");\n", s.Name)
	contents += tabtab + "boolean first = true;

"
	first := true
	for _, field := range s.Fields {
		optional := field.Modifier == parser.Optional
		ind := ""
		if optional {
			contents += fmt.Sprintf(tabtab+"if (isSet%s()) {
", strings.Title(field.Name))
			ind = tab
		}

		if !first {
			contents += tabtab + ind + "if (!first) sb.append(\", \");\n"
		}
		contents += fmt.Sprintf(tabtab+ind+"sb.append(\"%s:\");\n", field.Name)
		if !g.isJavaPrimitive(field.Type) {
			contents += fmt.Sprintf(tabtab+ind+"if (this.%s == null) {\n", field.Name)
			contents += tabtabtab + ind + "sb.append(\"null\");\n"
			contents += tabtab + ind + "} else {\n"
			if g.Frugal.UnderlyingType(field.Type).Name == "binary" {
				contents += fmt.Sprintf(tabtabtab+ind+"org.apache.thrift.TBaseHelper.toString(this.%s, sb);
", field.Name)
			} else {
				contents += fmt.Sprintf(tabtabtab+ind+"sb.append(this.%s);
", field.Name)
			}
			contents += tabtab + ind + "}\n"
		} else {
			contents += fmt.Sprintf(tabtab+ind+"sb.append(this.%s);
", field.Name)
		}
		contents += tabtab + ind + "first = false;
"

		if optional {
			contents += tabtab + "}\n"
		}
		first = false
	}

	contents += tabtab + "sb.append(\")\");\n"
	contents += tabtab + "return sb.toString();
"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateValidate(s *parser.Struct) string {
	contents := ""
	contents += tab + "public void validate() throws org.apache.thrift.TException {
"
	contents += tabtab + "// check for required fields
"
	for _, field := range s.Fields {
		if field.Modifier == parser.Required && !g.isJavaPrimitive(field.Type) {
			contents += fmt.Sprintf(tabtab+"if (%s == null) {
", field.Name)
			contents += fmt.Sprintf(tabtabtab+"throw new org.apache.thrift.protocol.TProtocolException(\"Required field '%s' was not present! Struct: \" + toString());\n",
				field.Name)
			contents += tabtab + "}\n"
		}
	}

	contents += tabtab + "// check for sub-struct validity
"
	for _, field := range s.Fields {
		if g.Frugal.IsStruct(field.Type) && !g.Frugal.IsUnion(field.Type) {
			contents += fmt.Sprintf(tabtab+"if (%s != null) {
", field.Name)
			contents += fmt.Sprintf(tabtabtab+"%s.validate();
", field.Name)
			contents += tabtab + "}\n"
		}
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateWriteObject(s *parser.Struct) string {
	contents := ""
	contents += tab + "private void writeObject(java.io.ObjectOutputStream out) throws java.io.IOException {
"
	contents += tabtab + "try {\n"
	contents += tabtabtab + "write(new org.apache.thrift.protocol.TCompactProtocol(new org.apache.thrift.transport.TIOStreamTransport(out)));
"
	contents += tabtab + "} catch (org.apache.thrift.TException te) {\n"
	contents += tabtabtab + "throw new java.io.IOException(te);\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateReadObject(s *parser.Struct) string {
	contents := ""
	contents += tab + "private void readObject(java.io.ObjectInputStream in) throws java.io.IOException, ClassNotFoundException {
"
	// isset stuff, don't do for unions
	contents += tabtab + "try {\n"
	if s.Type != parser.StructTypeUnion {
		contents += tabtabtab + "// it doesn't seem like you should have to do this, but java serialization is wacky, and doesn't call the default constructor.\n"
		isSetType, _ := g.getIsSetType(s)
		switch isSetType {
		case IsSetNone:
		// Do nothing
		case IsSetBitfield:
			contents += tabtabtab + "__isset_bitfield = 0;\n"
		case IsSetBitSet:
			contents += tabtabtab + "__isset_bit_vector = new BitSet(1);\n"
		}
	}

	contents += tabtabtab + "read(new org.apache.thrift.protocol.TCompactProtocol(new org.apache.thrift.transport.TIOStreamTransport(in)));\n"
	contents += tabtab + "} catch (org.apache.thrift.TException te) {\n"
	contents += tabtabtab + "throw new java.io.IOException(te);\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateStandardScheme(s *parser.Struct, isResult bool) string {
	contents := ""
	contents += fmt.Sprintf(tab+"private static class %sStandardSchemeFactory implements SchemeFactory {\n", s.Name)
	contents += fmt.Sprintf(tabtab+"public %sStandardScheme getScheme() {\n", s.Name)
	contents += fmt.Sprintf(tabtabtab+"return new %sStandardScheme();\n", s.Name)
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"private static class %sStandardScheme extends StandardScheme<%s> {\n\n", s.Name, s.Name)

	// read
	contents += fmt.Sprintf(tabtab+"public void read(org.apache.thrift.protocol.TProtocol iprot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += tabtabtab + "org.apache.thrift.protocol.TField schemeField;\n"
	contents += tabtabtab + "iprot.readStructBegin();\n"
	contents += tabtabtab + "while (true) {\n"
	contents += tabtabtabtab + "schemeField = iprot.readFieldBegin();\n"
	contents += tabtabtabtab + "if (schemeField.type == org.apache.thrift.protocol.TType.STOP) {\n"
	contents += tabtabtabtabtab + "break;\n"
	contents += tabtabtabtab + "}\n"
	contents += tabtabtabtab + "switch (schemeField.id) {\n"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtabtabtabtab+"case %d: // %s\n", field.ID, toConstantName(field.Name))
		contents += fmt.Sprintf(tabtabtabtabtabtab+"if (schemeField.type == %s) {\n", g.getTType(field.Type))
		contents += g.generateReadFieldRec(field, true, false, false, tabtabtabtabtabtabtab)
		contents += fmt.Sprintf(tabtabtabtabtabtabtab+"struct.set%sIsSet(true);\n", strings.Title(field.Name))
		contents += tabtabtabtabtabtab + "} else {\n"
		contents += tabtabtabtabtabtabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, schemeField.type);\n"
		contents += tabtabtabtabtabtab + "}\n"
		contents += tabtabtabtabtabtab + "break;\n"
	}
	contents += tabtabtabtabtab + "default:\n"
	contents += tabtabtabtabtabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, schemeField.type);\n"
	contents += tabtabtabtab + "}\n"
	contents += tabtabtabtab + "iprot.readFieldEnd();\n"
	contents += tabtabtab + "}\n"
	contents += tabtabtab + "iprot.readStructEnd();\n\n"

	contents += tabtabtab + "// check for required fields of primitive type, which can't be checked in the validate method\n"
	for _, field := range s.Fields {
		if field.Modifier == parser.Required && g.isJavaPrimitive(field.Type) {
			contents += fmt.Sprintf(tabtabtab+"if (!struct.isSet%s()) {\n", strings.Title(field.Name))
			contents += fmt.Sprintf(tabtabtabtab+"throw new org.apache.thrift.protocol.TProtocolException(\"Required field '%s' was not found in serialized data! Struct: \" + toString());\n", field.Name)
			contents += tabtabtab + "}\n"
		}
	}

	contents += tabtabtab + "struct.validate();\n"
	contents += tabtab + "}\n\n"

	// write
	contents += fmt.Sprintf(tabtab+"public void write(org.apache.thrift.protocol.TProtocol oprot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += tabtabtab + "struct.validate();\n\n"
	contents += tabtabtab + "oprot.writeStructBegin(STRUCT_DESC);\n"
	for _, field := range s.Fields {
		isPrimitive := g.isJavaPrimitive(field.Type)
		ind := tabtabtab
		optInd := tabtabtab
		if !isPrimitive {
			contents += fmt.Sprintf(ind+"if (struct.%s != null) {\n", field.Name)
			ind += tab
			optInd += tab
		}
		opt := field.Modifier == parser.Optional || (isResult && isPrimitive)
		if opt {
			contents += fmt.Sprintf(ind+"if (struct.isSet%s()) {\n", strings.Title(field.Name))
			ind += tab
		}

		contents += fmt.Sprintf(ind+"oprot.writeFieldBegin(%s_FIELD_DESC);\n", toConstantName(field.Name))
		contents += g.generateWriteFieldRec(field, true, false, ind)
		contents += ind + "oprot.writeFieldEnd();\n"

		if opt {
			contents += optInd + "}\n"
		}
		if !isPrimitive {
			contents += tabtabtab + "}\n"
		}
	}
	contents += tabtabtab + "oprot.writeFieldStop();\n"
	contents += tabtabtab + "oprot.writeStructEnd();\n"

	contents += tabtab + "}\n\n"

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateTupleScheme(s *parser.Struct) string {
	contents := ""
	contents += fmt.Sprintf(tab+"private static class %sTupleSchemeFactory implements SchemeFactory {\n", s.Name)
	contents += fmt.Sprintf(tabtab+"public %sTupleScheme getScheme() {\n", s.Name)
	contents += fmt.Sprintf(tabtabtab+"return new %sTupleScheme();\n", s.Name)
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"private static class %sTupleScheme extends TupleScheme<%s> {\n\n", s.Name, s.Name)
	contents += tabtab + "@Override\n"
	contents += fmt.Sprintf(tabtab+"public void write(org.apache.thrift.protocol.TProtocol prot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += tabtabtab + "TTupleProtocol oprot = (TTupleProtocol) prot;\n"
	// write required fields
	numNonReqs := 0
	for _, field := range s.Fields {
		if field.Modifier != parser.Required {
			numNonReqs += 1
			continue
		}

		contents += g.generateWriteFieldRec(field, true, true, tabtabtab)
	}

	if numNonReqs > 0 {
		// write optional/default fields
		nonReqFieldCount := 0
		contents += tabtabtab + "BitSet optionals = new BitSet();\n"
		for _, field := range s.Fields {
			if field.Modifier == parser.Required {
				continue
			}

			contents += fmt.Sprintf(tabtabtab+"if (struct.isSet%s()) {\n", strings.Title(field.Name))
			contents += fmt.Sprintf(tabtabtabtab+"optionals.set(%d);\n", nonReqFieldCount)
			contents += tabtabtab + "}\n"
			nonReqFieldCount += 1
		}

		contents += fmt.Sprintf(tabtabtab+"oprot.writeBitSet(optionals, %d);\n", numNonReqs)
		for _, field := range s.Fields {
			if field.Modifier == parser.Required {
				continue
			}

			contents += fmt.Sprintf(tabtabtab+"if (struct.isSet%s()) {\n", strings.Title(field.Name))
			contents += g.generateWriteFieldRec(field, true, true, tabtabtabtab)
			contents += tabtabtab + "}\n"
		}
	}

	contents += tabtab + "}\n\n"

	contents += tabtab + "@Override\n"
	contents += fmt.Sprintf(tabtab+"public void read(org.apache.thrift.protocol.TProtocol prot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += tabtabtab + "TTupleProtocol iprot = (TTupleProtocol) prot;\n"
	// read required fields
	for _, field := range s.Fields {
		if field.Modifier != parser.Required {
			continue
		}

		contents += g.generateReadFieldRec(field, true, true, false, tabtabtab)
		contents += fmt.Sprintf(tabtabtab+"struct.set%sIsSet(true);\n", strings.Title(field.Name))
	}

	if numNonReqs > 0 {
		// read default/optional fields
		nonReqFieldCount := 0
		contents += fmt.Sprintf(tabtabtab+"BitSet incoming = iprot.readBitSet(%d);\n", numNonReqs)
		for _, field := range s.Fields {
			if field.Modifier == parser.Required {
				continue
			}

			contents += fmt.Sprintf(tabtabtab+"if (incoming.get(%d)) {\n", nonReqFieldCount)
			contents += g.generateReadFieldRec(field, true, true, false, tabtabtabtab)
			contents += fmt.Sprintf(tabtabtabtab+"struct.set%sIsSet(true);\n", strings.Title(field.Name))
			contents += tabtabtab + "}\n"
			nonReqFieldCount += 1
		}
	}
	contents += tabtab + "}\n\n"

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateCopyConstructorField(field *parser.Field, otherFieldName string, first bool, ind string) string {
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	isPrimitive := g.isJavaPrimitive(underlyingType)
	accessPrefix := "this."
	declPrefix := "this."
	if !first {
		accessPrefix = ""
		declPrefix = g.getJavaTypeFromThriftType(underlyingType) + " "
	}

	if isPrimitive || underlyingType.Name == "string" {
		return fmt.Sprintf(ind+"%s%s = %s;\n", declPrefix, field.Name, otherFieldName)
	} else if underlyingType.Name == "binary" {
		return fmt.Sprintf(ind+"%s%s = org.apache.thrift.TBaseHelper.copyBinary(%s);\n", declPrefix, field.Name, otherFieldName)
	} else if g.Frugal.IsStruct(underlyingType) {
		return fmt.Sprintf(ind+"%s%s = new %s(%s);\n", declPrefix, field.Name, g.getJavaTypeFromThriftType(underlyingType), otherFieldName)
	} else if g.Frugal.IsEnum(underlyingType) {
		return fmt.Sprintf(ind+"%s%s = %s;\n", declPrefix, field.Name, otherFieldName)
	} else if underlyingType.IsContainer() {
		contents := ""
		valueType := g.getJavaTypeFromThriftType(underlyingType.ValueType)
		containerValType := containerType(valueType)
		otherValElem := g.GetElem()
		thisValElem := g.GetElem()
		thisValField := parser.FieldFromType(underlyingType.ValueType, thisValElem)

		switch underlyingType.Name {
		case "list":
			contents += fmt.Sprintf(ind+"%s%s = new ArrayList<%s>(%s.size());\n", declPrefix, field.Name, containerValType, otherFieldName)
			contents += fmt.Sprintf(ind+"for (%s %s : %s) {\n", valueType, otherValElem, otherFieldName)
			contents += g.generateCopyConstructorField(thisValField, otherValElem, false, ind+tab)
			contents += fmt.Sprintf(ind+tab+"%s%s.add(%s);\n", accessPrefix, field.Name, thisValElem)
			contents += ind + "}\n"
		case "set":
			contents += fmt.Sprintf(ind+"%s%s = new HashSet<%s>(%s.size());\n", declPrefix, field.Name, containerValType, otherFieldName)
			contents += fmt.Sprintf(ind+"for (%s %s : %s) {\n", valueType, otherValElem, otherFieldName)
			contents += g.generateCopyConstructorField(thisValField, otherValElem, false, ind+tab)
			contents += fmt.Sprintf(ind+tab+"%s%s.add(%s);\n", accessPrefix, field.Name, thisValElem)
			contents += ind + "}\n"
		case "map":
			keyType := g.getJavaTypeFromThriftType(underlyingType.KeyType)
			keyUnderlying := g.Frugal.UnderlyingType(underlyingType.KeyType)
			valUnderlying := g.Frugal.UnderlyingType(underlyingType.ValueType)
			containerKeyType := containerType(keyType)

			// If it's all primitives, optimization. Otherwise need to iterate
			if (g.isJavaPrimitive(keyUnderlying) || keyUnderlying.Name == "string") &&
				(g.isJavaPrimitive(valUnderlying) || valUnderlying.Name == "string") {
				contents += fmt.Sprintf(ind+"%s%s = new HashMap<%s,%s>(%s);
",
					declPrefix, field.Name, containerKeyType, containerValType, otherFieldName)
			} else {
				thisKeyElem := g.GetElem()
				thisKeyField := parser.FieldFromType(underlyingType.KeyType, thisKeyElem)

				contents += fmt.Sprintf(ind+"%s%s = new HashMap<%s,%s>(%s.size());
",
					declPrefix, field.Name, containerKeyType, containerValType, otherFieldName)
				contents += fmt.Sprintf(ind+"for (Map.Entry<%s, %s> %s : %s.entrySet()) {
",
					containerKeyType, containerValType, otherValElem, otherFieldName)
				contents += g.generateCopyConstructorField(thisKeyField, otherValElem+".getKey()", false, ind+tab)
				contents += g.generateCopyConstructorField(thisValField, otherValElem+".getValue()", false, ind+tab)
				contents += fmt.Sprintf(ind+tab+"%s%s.put(%s, %s);\n", accessPrefix, field.Name, thisKeyElem, thisValElem)
				contents += ind + "}\n"
			}
		}

		return contents
	}
	panic("unrecognized type: " + underlyingType.Name)
}

func (g *Generator) generateMetaDataMapEntry(t *parser.Type, ind string) string {
	underlyingType := g.Frugal.UnderlyingType(t)
	ttype := g.getTType(underlyingType)
	isThriftPrimitive := underlyingType.IsPrimitive()

	// This indicates a typedef. For some reason java doesn't recurse on
	// typedef'd types for meta data map entries
	if t != underlyingType {
		secondArg := ""
		if underlyingType.Name == "binary" {
			secondArg = ", true"
		} else {
			secondArg = fmt.Sprintf(", \"%s\"", t.Name)
		}
		return fmt.Sprintf(ind+"new org.apache.thrift.meta_data.FieldValueMetaData(%s%s)", ttype, secondArg)
	}

	if isThriftPrimitive {
		boolArg := ""
		if underlyingType.Name == "binary" {
			boolArg = ", true"
		}
		return fmt.Sprintf(ind+"new org.apache.thrift.meta_data.FieldValueMetaData(%s%s)", ttype, boolArg)
	} else if g.Frugal.IsStruct(underlyingType) {
		return fmt.Sprintf(ind+"new org.apache.thrift.meta_data.StructMetaData(org.apache.thrift.protocol.TType.STRUCT, %s.class)", g.qualifiedTypeName(underlyingType))
	} else if g.Frugal.IsEnum(underlyingType) {
		return fmt.Sprintf(ind+"new org.apache.thrift.meta_data.EnumMetaData(org.apache.thrift.protocol.TType.ENUM, %s.class)", g.qualifiedTypeName(underlyingType))
	} else if underlyingType.IsContainer() {
		switch underlyingType.Name {
		case "list":
			valEntry := g.generateMetaDataMapEntry(underlyingType.ValueType, ind+tabtab)
			return fmt.Sprintf(ind+"new org.apache.thrift.meta_data.ListMetaData(org.apache.thrift.protocol.TType.LIST,\n%s)", valEntry)
		case "set":
			valEntry := g.generateMetaDataMapEntry(underlyingType.ValueType, ind+tabtab)
			return fmt.Sprintf(ind+"new org.apache.thrift.meta_data.SetMetaData(org.apache.thrift.protocol.TType.SET,\n%s)", valEntry)
		case "map":
			keyEntry := g.generateMetaDataMapEntry(underlyingType.KeyType, ind+tabtab)
			valEntry := g.generateMetaDataMapEntry(underlyingType.ValueType, ind+tabtab)
			return fmt.Sprintf(ind+"new org.apache.thrift.meta_data.MapMetaData(org.apache.thrift.protocol.TType.MAP,\n%s,\n%s)", keyEntry, valEntry)
		}
	}
	panic("unrecognized type: " + underlyingType.Name)
}

// succinct means only read collection length instead of the whole header,
// and don't read collection end.
// containerTypes causes variables to be declared as objects instead of
// potentially primitives
func (g *Generator) generateReadFieldRec(field *parser.Field, first bool, succinct bool, containerTypes bool, ind string) string {
	contents := ""
	declPrefix := "struct."
	accessPrefix := "struct."
	javaType := g.getJavaTypeFromThriftType(field.Type)
	if !first {
		if containerTypes {
			declPrefix = containerType(javaType) + " "
		} else {
			declPrefix = javaType + " "
		}
		accessPrefix = ""
	}

	underlyingType := g.Frugal.UnderlyingType(field.Type)
	if underlyingType.IsPrimitive() {
		thriftType := ""
		switch underlyingType.Name {
		case "bool":
			thriftType = "Bool"
		case "byte", "i8":
			thriftType = "Byte"
		case "i16":
			thriftType = "I16"
		case "i32":
			thriftType = "I32"
		case "i64":
			thriftType = "I64"
		case "double":
			thriftType = "Double"
		case "string":
			thriftType = "String"
		case "binary":
			thriftType = "Binary"
		default:
			panic("unkown thrift type: " + underlyingType.Name)
		}

		contents += fmt.Sprintf(ind+"%s%s = iprot.read%s();\n", declPrefix, field.Name, thriftType)
	} else if g.Frugal.IsEnum(underlyingType) {
		contents += fmt.Sprintf(ind+"%s%s = %s.findByValue(iprot.readI32());\n", declPrefix, field.Name, javaType)
	} else if g.Frugal.IsStruct(underlyingType) {
		contents += fmt.Sprintf(ind+"%s%s = new %s();\n", declPrefix, field.Name, javaType)
		contents += fmt.Sprintf(ind+"%s%s.read(iprot);\n", accessPrefix, field.Name)
	} else if underlyingType.IsContainer() {
		containerElem := g.GetElem()
		counterElem := g.GetElem()

		valType := containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType))
		valElem := g.GetElem()
		valField := parser.FieldFromType(underlyingType.ValueType, valElem)
		valContents := g.generateReadFieldRec(valField, false, succinct, containerTypes, ind+tab)
		valTType := g.getTType(underlyingType.ValueType)

		switch underlyingType.Name {
		case "list":
			if succinct {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TList %s = new org.apache.thrift.protocol.TList(%s, iprot.readI32());\n",
					containerElem, valTType)
			} else {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TList %s = iprot.readListBegin();\n", containerElem)
			}
			contents += fmt.Sprintf(ind+"%s%s = new ArrayList<%s>(%s.size);\n", declPrefix, field.Name, valType, containerElem)
			contents += fmt.Sprintf(ind+"for (int %s = 0; %s < %s.size; ++%s) {\n", counterElem, counterElem, containerElem, counterElem)
			contents += valContents
			contents += fmt.Sprintf(ind+tab+"%s%s.add(%s);\n", accessPrefix, field.Name, valElem)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "iprot.readListEnd();\n"
			}
		case "set":
			if succinct {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TSet %s = new org.apache.thrift.protocol.TSet(%s, iprot.readI32());\n",
					containerElem, valTType)
			} else {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TSet %s = iprot.readSetBegin();\n", containerElem)
			}
			contents += fmt.Sprintf(ind+"%s%s = new HashSet<%s>(2*%s.size);\n", declPrefix, field.Name, valType, containerElem)
			contents += fmt.Sprintf(ind+"for (int %s = 0; %s < %s.size; ++%s) {\n", counterElem, counterElem, containerElem, counterElem)
			contents += valContents
			contents += fmt.Sprintf(ind+tab+"%s%s.add(%s);\n", accessPrefix, field.Name, valElem)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "iprot.readSetEnd();\n"
			}
		case "map":
			keyTType := g.getTType(underlyingType.KeyType)
			if succinct {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TMap %s = new org.apache.thrift.protocol.TMap(%s, %s, iprot.readI32());\n",
					containerElem, keyTType, valTType)
			} else {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TMap %s = iprot.readMapBegin();\n", containerElem)
			}

			keyType := containerType(g.getJavaTypeFromThriftType(underlyingType.KeyType))
			contents += fmt.Sprintf(ind+"%s%s = new HashMap<%s,%s>(2*%s.size);\n", declPrefix, field.Name, keyType, valType, containerElem)
			contents += fmt.Sprintf(ind+"for (int %s = 0; %s < %s.size; ++%s) {\n", counterElem, counterElem, containerElem, counterElem)
			keyElem := g.GetElem()
			keyField := parser.FieldFromType(underlyingType.KeyType, keyElem)
			contents += g.generateReadFieldRec(keyField, false, succinct, containerTypes, ind+tab)
			contents += valContents
			contents += fmt.Sprintf(ind+tab+"%s%s.put(%s, %s);\n", accessPrefix, field.Name, keyElem, valElem)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "iprot.readMapEnd();\n"
			}
		}
	}

	return contents
}

// succinct means only write collection length instead of the whole header,
// and don't write collection end.
func (g *Generator) generateWriteFieldRec(field *parser.Field, first bool, succinct bool, ind string) string {
	contents := ""
	accessPrefix := "struct."
	if !first {
		accessPrefix = ""
	}

	underlyingType := g.Frugal.UnderlyingType(field.Type)
	isEnum := g.Frugal.IsEnum(underlyingType)
	if underlyingType.IsPrimitive() || isEnum {
		write := ind + "oprot.write"
		switch underlyingType.Name {
		case "bool":
			write += "Bool(%s%s);
"
		case "byte", "i8":
			write += "Byte(%s%s);
"
		case "i16":
			write += "I16(%s%s);
"
		case "i32":
			write += "I32(%s%s);
"
		case "i64":
			write += "I64(%s%s);
"
		case "double":
			write += "Double(%s%s);
"
		case "string":
			write += "String(%s%s);
"
		case "binary":
			write += "Binary(%s%s);
"
		default:
			if isEnum {
				write += "I32(%s%s.getValue());
"
			} else {
				panic("unknown thrift type: " + underlyingType.Name)
			}
		}

		contents += fmt.Sprintf(write, accessPrefix, field.Name)
	} else if g.Frugal.IsStruct(underlyingType) {
		contents += fmt.Sprintf(ind+"%s%s.write(oprot);
", accessPrefix, field.Name)
	} else if underlyingType.IsContainer() {
		iterElem := g.GetElem()
		valJavaType := g.getJavaTypeFromThriftType(underlyingType.ValueType)
		valTType := g.getTType(underlyingType.ValueType)

		switch underlyingType.Name {
		case "list":
			if succinct {
				contents += fmt.Sprintf(ind+"oprot.writeI32(%s%s.size());
", accessPrefix, field.Name)
			} else {
				contents += fmt.Sprintf(ind+"oprot.writeListBegin(new org.apache.thrift.protocol.TList(%s, %s%s.size()));
",
					valTType, accessPrefix, field.Name)
			}
			contents += fmt.Sprintf(ind+"for (%s %s : %s%s) {
", valJavaType, iterElem, accessPrefix, field.Name)
			iterField := parser.FieldFromType(underlyingType.ValueType, iterElem)
			contents += g.generateWriteFieldRec(iterField, false, succinct, ind+tab)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "oprot.writeListEnd();
"
			}
		case "set":
			if succinct {
				contents += fmt.Sprintf(ind+"oprot.writeI32(%s%s.size());
", accessPrefix, field.Name)
			} else {
				contents += fmt.Sprintf(ind+"oprot.writeSetBegin(new org.apache.thrift.protocol.TSet(%s, %s%s.size()));
",
					valTType, accessPrefix, field.Name)
			}
			contents += fmt.Sprintf(ind+"for (%s %s : %s%s) {
", valJavaType, iterElem, accessPrefix, field.Name)
			iterField := parser.FieldFromType(underlyingType.ValueType, iterElem)
			contents += g.generateWriteFieldRec(iterField, false, succinct, ind+tab)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "oprot.writeSetEnd();
"
			}
		case "map":
			keyJavaType := g.getJavaTypeFromThriftType(underlyingType.KeyType)
			keyTType := g.getTType(underlyingType.KeyType)
			if succinct {
				contents += fmt.Sprintf(ind+"oprot.writeI32(%s%s.size());
", accessPrefix, field.Name)
			} else {
				contents += fmt.Sprintf(ind+"oprot.writeMapBegin(new org.apache.thrift.protocol.TMap(%s, %s, %s%s.size()));
",
					keyTType, valTType, accessPrefix, field.Name)
			}
			contents += fmt.Sprintf(ind+"for (Map.Entry<%s, %s> %s : %s%s.entrySet()) {
",
				containerType(keyJavaType), containerType(valJavaType), iterElem, accessPrefix, field.Name)
			keyField := parser.FieldFromType(underlyingType.KeyType, iterElem+".getKey()")
			valField := parser.FieldFromType(underlyingType.ValueType, iterElem+".getValue()")
			contents += g.generateWriteFieldRec(keyField, false, succinct, ind+tab)
			contents += g.generateWriteFieldRec(valField, false, succinct, ind+tab)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "oprot.writeMapEnd();
"
			}
		default:
			panic("unknown type: " + underlyingType.Name)
		}
	}

	return contents
}

func (g *Generator) GetOutputDir(dir string) string {
	if pkg, ok := g.Frugal.Thrift.Namespace(lang); ok {
		path := generator.GetPackageComponents(pkg)
		dir = filepath.Join(append([]string{dir}, path...)...)
	}
	return dir
}

func (g *Generator) DefaultOutputDir() string {
	return defaultOutputDir
}

func (g *Generator) PostProcess(f *os.File) error { return nil }

func (g *Generator) GenerateDependencies(dir string) error {
	return nil
}

func (g *Generator) GenerateFile(name, outputDir string, fileType generator.FileType) (*os.File, error) {
	switch fileType {
	case generator.PublishFile:
		return g.CreateFile(strings.Title(name)+"Publisher", outputDir, lang, false)
	case generator.SubscribeFile:
		return g.CreateFile(strings.Title(name)+"Subscriber", outputDir, lang, false)
	case generator.CombinedServiceFile:
		return g.CreateFile("F"+name, outputDir, lang, false)
	case generator.ObjectFile:
		return g.CreateFile(name, outputDir, lang, false)
	default:
		return nil, fmt.Errorf("Bad file type for Java generator: %s", fileType)
	}
}

func (g *Generator) GenerateDocStringComment(file *os.File) error {
	comment := fmt.Sprintf(
		"/**
"+
			" * Autogenerated by Frugal Compiler (%s)
"+
			" * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
"+
			" *
"+
			" * @generated
"+
			" */",
		globals.Version)

	_, err := file.WriteString(comment)
	return err
}

func (g *Generator) GenerateServicePackage(file *os.File, s *parser.Service) error {
	return g.generatePackage(file)
}

func (g *Generator) GenerateScopePackage(file *os.File, s *parser.Scope) error {
	return g.generatePackage(file)
}

func (g *Generator) generatePackage(file *os.File) error {
	pkg, ok := g.Frugal.Thrift.Namespace(lang)
	if !ok {
		return nil
	}
	_, err := file.WriteString(fmt.Sprintf("package %s;", pkg))
	return err
}

func (g *Generator) GenerateEnumImports(file *os.File) error {
	imports := ""
	imports += "import java.util.Map;
"
	imports += "import java.util.HashMap;
"
	imports += "import org.apache.thrift.TEnum;
"
	imports += "\n"

	_, err := file.WriteString(imports)
	return err
}

func (g *Generator) GenerateStructImports(file *os.File) error {
	imports := ""
	imports += "import org.apache.thrift.scheme.IScheme;
"
	imports += "import org.apache.thrift.scheme.SchemeFactory;
"
	imports += "import org.apache.thrift.scheme.StandardScheme;

"
	imports += "import org.apache.thrift.scheme.TupleScheme;
"
	imports += "import org.apache.thrift.protocol.TTupleProtocol;
"
	imports += "import org.apache.thrift.protocol.TProtocolException;
"
	imports += "import org.apache.thrift.EncodingUtils;
"
	imports += "import org.apache.thrift.TException;
"
	imports += "import org.apache.thrift.async.AsyncMethodCallback;
"
	imports += "import org.apache.thrift.server.AbstractNonblockingServer.*;
"
	imports += "import java.util.List;
"
	imports += "import java.util.ArrayList;
"
	imports += "import java.util.Map;
"
	imports += "import java.util.HashMap;
"
	imports += "import java.util.EnumMap;
"
	imports += "import java.util.Set;
"
	imports += "import java.util.HashSet;
"
	imports += "import java.util.EnumSet;
"
	imports += "import java.util.Collections;
"
	imports += "import java.util.BitSet;
"
	imports += "import java.nio.ByteBuffer;
"
	imports += "import java.util.Arrays;
"
	imports += "import javax.annotation.Generated;
"
	imports += "import org.slf4j.Logger;
"
	imports += "import org.slf4j.LoggerFactory;
"

	imports += "\n"

	_, err := file.WriteString(imports)
	return err
}

func (g *Generator) GenerateServiceImports(file *os.File, s *parser.Service) error {
	imports := "import com.workiva.frugal.exception.FMessageSizeException;
"
	imports += "import com.workiva.frugal.exception.FTimeoutException;
"
	imports += "import com.workiva.frugal.middleware.InvocationHandler;
"
	imports += "import com.workiva.frugal.middleware.ServiceMiddleware;
"
	imports += "import com.workiva.frugal.processor.FBaseProcessor;
"
	imports += "import com.workiva.frugal.processor.FProcessor;
"
	imports += "import com.workiva.frugal.processor.FProcessorFunction;
"
	imports += "import com.workiva.frugal.protocol.*;
"
	imports += "import com.workiva.frugal.transport.FTransport;
"
	imports += "import org.apache.thrift.TApplicationException;
"
	imports += "import org.apache.thrift.TException;
"
	imports += "import org.apache.thrift.protocol.TMessage;
"
	imports += "import org.apache.thrift.protocol.TMessageType;
"
	imports += "import org.apache.thrift.transport.TTransport;

"

	imports += "import javax.annotation.Generated;
"
	imports += "import java.util.concurrent.*;
"

	_, err := file.WriteString(imports)
	return err
}

func (g *Generator) GenerateScopeImports(file *os.File, s *parser.Scope) error {
	imports := "import com.workiva.frugal.middleware.InvocationHandler;
"
	imports += "import com.workiva.frugal.middleware.ServiceMiddleware;
"
	imports += "import com.workiva.frugal.protocol.*;
"
	imports += "import com.workiva.frugal.provider.FScopeProvider;
"
	imports += "import com.workiva.frugal.transport.FScopeTransport;
"
	imports += "import com.workiva.frugal.transport.FSubscription;
"
	imports += "import org.apache.thrift.TException;
"
	imports += "import org.apache.thrift.TApplicationException;
"
	imports += "import org.apache.thrift.transport.TTransportException;
"
	imports += "import org.apache.thrift.protocol.*;

"

	imports += "import javax.annotation.Generated;
"
	imports += "import java.util.logging.Logger;
"

	_, err := file.WriteString(imports)
	return err
}

func (g *Generator) GenerateConstants(file *os.File, name string) error {
	return nil
}

func (g *Generator) GeneratePublisher(file *os.File, scope *parser.Scope) error {
	scopeTitle := strings.Title(scope.Name)
	publisher := ""
	if scope.Comment != nil {
		publisher += g.GenerateBlockComment(scope.Comment, "")
	}
	if g.includeGeneratedAnnotation() {
		publisher += g.generatedAnnotation()
	}
	publisher += fmt.Sprintf("public class %sPublisher {

", scopeTitle)

	publisher += g.generatePublisherIface(scope)
	publisher += g.generatePublisherClient(scope)

	publisher += "}"

	_, err := file.WriteString(publisher)
	return err
}

func (g *Generator) generatePublisherIface(scope *parser.Scope) string {
	contents := ""
	contents += tab + "public interface Iface {
"

	contents += tabtab + "public void open() throws TException;

"
	contents += tabtab + "public void close() throws TException;

"

	args := g.generateScopePrefixArgs(scope)

	for _, op := range scope.Operations {
		contents += fmt.Sprintf(tabtab+"public void publish%s(FContext ctx, %s%s req) throws TException;

", op.Name, args, g.qualifiedTypeName(op.Type))
	}

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generatePublisherClient(scope *parser.Scope) string {
	publisher := ""

	scopeTitle := strings.Title(scope.Name)
	publisher += tab + "public static class Client implements Iface {
"
	publisher += fmt.Sprintf(tabtab+"private static final String DELIMITER = \"%s\";

", globals.TopicDelimiter)
	publisher += tabtab + "private final Iface target;
"
	publisher += tabtab + "private final Iface proxy;

"

	publisher += tabtab + "public Client(FScopeProvider provider, ServiceMiddleware... middleware) {
"
	publisher += fmt.Sprintf(tabtabtab+"target = new Internal%sPublisher(provider);
", scopeTitle)
	publisher += tabtabtab + "proxy = InvocationHandler.composeMiddleware(target, Iface.class, middleware);
"
	publisher += tabtab + "}\n\n"

	publisher += tabtab + "public void open() throws TException {
"
	publisher += tabtabtab + "target.open();
"
	publisher += tabtab + "}\n\n"

	publisher += tabtab + "public void close() throws TException {
"
	publisher += tabtabtab + "target.close();
"
	publisher += tabtab + "}\n\n"

	args := g.generateScopePrefixArgs(scope)

	for _, op := range scope.Operations {
		if op.Comment != nil {
			publisher += g.GenerateBlockComment(op.Comment, tab)
		}
		publisher += fmt.Sprintf(tabtab+"public void publish%s(FContext ctx, %s%s req) throws TException {
", op.Name, args, g.qualifiedTypeName(op.Type))
		publisher += fmt.Sprintf(tabtabtab+"proxy.publish%s(%s);
", op.Name, g.generateScopeArgs(scope))
		publisher += tabtab + "}\n\n"
	}

	publisher += fmt.Sprintf(tabtab+"protected static class Internal%sPublisher implements Iface {

", scopeTitle)

	publisher += tabtabtab + "private FScopeProvider provider;
"
	publisher += tabtabtab + "private FScopeTransport transport;
"
	publisher += tabtabtab + "private FProtocol protocol;

"

	publisher += fmt.Sprintf(tabtabtab+"protected Internal%sPublisher() {
", scopeTitle)
	publisher += tabtabtab + "}\n\n"

	publisher += fmt.Sprintf(tabtabtab+"public Internal%sPublisher(FScopeProvider provider) {
", scopeTitle)
	publisher += tabtabtabtab + "this.provider = provider;
"
	publisher += tabtabtab + "}\n\n"

	publisher += tabtabtab + "public void open() throws TException {
"
	publisher += tabtabtabtab + "FScopeProvider.Client client = provider.build();
"
	publisher += tabtabtabtab + "transport = client.getTransport();
"
	publisher += tabtabtabtab + "protocol = client.getProtocol();
"
	publisher += tabtabtabtab + "transport.open();
"
	publisher += tabtabtab + "}\n\n"

	publisher += tabtabtab + "public void close() throws TException {
"
	publisher += tabtabtabtab + "transport.close();
"
	publisher += tabtabtab + "}\n\n"

	prefix := ""
	for _, op := range scope.Operations {
		publisher += prefix
		prefix = "\n\n"
		if op.Comment != nil {
			publisher += g.GenerateBlockComment(op.Comment, tabtabtab)
		}
		publisher += fmt.Sprintf(tabtabtab+"public void publish%s(FContext ctx, %s%s req) throws TException {
", op.Name, args, g.qualifiedTypeName(op.Type))
		publisher += fmt.Sprintf(tabtabtabtab+"String op = \"%s\";
", op.Name)
		publisher += fmt.Sprintf(tabtabtabtab+"String prefix = %s;
", generatePrefixStringTemplate(scope))
		publisher += tabtabtabtab + "String topic = String.format(\"%s" + strings.Title(scope.Name) + "%s%s\", prefix, DELIMITER, op);
"
		publisher += tabtabtabtab + "transport.lockTopic(topic);
"
		publisher += tabtabtabtab + "try {\n"
		publisher += tabtabtabtabtab + "protocol.writeRequestHeader(ctx);
"
		publisher += tabtabtabtabtab + "protocol.writeMessageBegin(new TMessage(op, TMessageType.CALL, 0));
"
		publisher += tabtabtabtabtab + "req.write(protocol);
"
		publisher += tabtabtabtabtab + "protocol.writeMessageEnd();
"
		publisher += tabtabtabtabtab + "transport.flush();
"
		publisher += tabtabtabtab + "} finally {
"
		publisher += tabtabtabtabtab + "transport.unlockTopic();
"
		publisher += tabtabtabtab + "}\n"
		publisher += tabtabtab + "}\n"
	}

	publisher += tabtab + "}\n"
	publisher += tab + "}\n"

	return publisher
}

func generatePrefixStringTemplate(scope *parser.Scope) string {
	if len(scope.Prefix.Variables) == 0 {
		if scope.Prefix.String == "" {
			return `""`
		}
		return fmt.Sprintf(`"%s%s"`, scope.Prefix.String, globals.TopicDelimiter)
	}
	template := "String.format(\""
	template += scope.Prefix.Template("%s")
	template += globals.TopicDelimiter + "\", "
	prefix := ""
	for _, variable := range scope.Prefix.Variables {
		template += prefix + variable
		prefix = ", "
	}
	template += ")"
	return template
}

func (g *Generator) GenerateSubscriber(file *os.File, scope *parser.Scope) error {
	subscriber := ""
	if scope.Comment != nil {
		subscriber += g.GenerateBlockComment(scope.Comment, "")
	}
	scopeName := strings.Title(scope.Name)
	if g.includeGeneratedAnnotation() {
		subscriber += g.generatedAnnotation()
	}
	subscriber += fmt.Sprintf("public class %sSubscriber {

", scopeName)

	subscriber += g.generateSubscriberIface(scope)
	subscriber += g.generateHandlerIfaces(scope)
	subscriber += g.generateSubscriberClient(scope)

	subscriber += "
}"

	_, err := file.WriteString(subscriber)
	return err
}

func (g *Generator) generateSubscriberIface(scope *parser.Scope) string {
	contents := ""

	contents += tab + "public interface Iface {
"

	args := g.generateScopePrefixArgs(scope)

	for _, op := range scope.Operations {
		contents += fmt.Sprintf(tabtab+"public FSubscription subscribe%s(%sfinal %sHandler handler) throws TException;
",
			op.Name, args, op.Name)
	}

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateHandlerIfaces(scope *parser.Scope) string {
	contents := ""

	for _, op := range scope.Operations {
		contents += fmt.Sprintf(tab+"public interface %sHandler {
", op.Name)
		contents += fmt.Sprintf(tabtab+"void on%s(FContext ctx, %s req);
", op.Name, g.qualifiedTypeName(op.Type))
		contents += tab + "}\n\n"
	}

	return contents
}

func (g *Generator) generateSubscriberClient(scope *parser.Scope) string {
	subscriber := ""

	prefix := ""
	args := g.generateScopePrefixArgs(scope)

	subscriber += tab + "public static class Client implements Iface {
"

	subscriber += fmt.Sprintf(tabtab+"private static final String DELIMITER = \"%s\";
", globals.TopicDelimiter)
	subscriber += tabtab + "private static final Logger LOGGER = Logger.getLogger(Client.class.getName());

"

	subscriber += tabtab + "private final FScopeProvider provider;
"
	subscriber += tabtab + "private final ServiceMiddleware[] middleware;

"

	subscriber += tabtab + "public Client(FScopeProvider provider, ServiceMiddleware... middleware) {
"
	subscriber += tabtabtab + "this.provider = provider;
"
	subscriber += tabtabtab + "this.middleware = middleware;
"
	subscriber += tabtab + "}\n\n"

	for _, op := range scope.Operations {
		subscriber += prefix
		prefix = "\n\n"
		if op.Comment != nil {
			subscriber += g.GenerateBlockComment(op.Comment, tabtab)
		}
		subscriber += fmt.Sprintf(tabtab+"public FSubscription subscribe%s(%sfinal %sHandler handler) throws TException {
",
			op.Name, args, op.Name)
		subscriber += fmt.Sprintf(tabtabtab+"final String op = \"%s\";
", op.Name)
		subscriber += fmt.Sprintf(tabtabtab+"String prefix = %s;
", generatePrefixStringTemplate(scope))
		subscriber += tabtabtab + "final String topic = String.format(\"%s" + strings.Title(scope.Name) + "%s%s\", prefix, DELIMITER, op);
"
		subscriber += tabtabtab + "final FScopeProvider.Client client = provider.build();
"
		subscriber += tabtabtab + "final FScopeTransport transport = client.getTransport();
"
		subscriber += tabtabtab + "transport.subscribe(topic);

"

		subscriber += tabtabtab + fmt.Sprintf(
			"final %sHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, %sHandler.class, middleware);
",
			op.Name, op.Name)
		subscriber += tabtabtab + "final FSubscription sub = FSubscription.of(topic, transport);
"
		subscriber += tabtabtab + "new Thread(new Runnable() {
"
		subscriber += tabtabtabtab + "public void run() {
"
		subscriber += tabtabtabtabtab + "while (true) {\n"
		subscriber += tabtabtabtabtabtab + "try {\n"
		subscriber += tabtabtabtabtabtabtab + "FContext ctx = client.getProtocol().readRequestHeader();
"
		subscriber += tabtabtabtabtabtabtab + fmt.Sprintf("%s received = recv%s(op, client.getProtocol());
",
			g.qualifiedTypeName(op.Type), op.Name)
		subscriber += tabtabtabtabtabtabtab + fmt.Sprintf("proxiedHandler.on%s(ctx, received);
", op.Name)
		subscriber += tabtabtabtabtabtab + "} catch (TException e) {
"
		subscriber += tabtabtabtabtabtabtab + "if (e instanceof TTransportException) {
"
		subscriber += tabtabtabtabtabtabtabtab + "TTransportException transportException = (TTransportException) e;
"
		subscriber += tabtabtabtabtabtabtabtab + "if (transportException.getType() == TTransportException.END_OF_FILE) {
"
		subscriber += tabtabtabtabtabtabtabtabtab + "return;
"
		subscriber += tabtabtabtabtabtabtabtab + "}\n"
		subscriber += tabtabtabtabtabtabtab + "}\n"
		subscriber += tabtabtabtabtabtabtab + "LOGGER.warning(String.format(\"Subscriber error receiving %s, discarding frame: %s\", topic, e.getMessage()));
"
		subscriber += tabtabtabtabtabtabtab + "transport.discardFrame();
"
		subscriber += tabtabtabtabtabtab + "}\n"
		subscriber += tabtabtabtabtab + "}\n"
		subscriber += tabtabtabtab + "}\n"
		subscriber += tabtabtab + "}, \"subscription\").start();

"

		subscriber += tabtabtab + "return sub;
"
		subscriber += tabtab + "}\n\n"

		subscriber += tabtab + fmt.Sprintf("private %s recv%s(String op, FProtocol iprot) throws TException {
", g.qualifiedTypeName(op.Type), op.Name)
		subscriber += tabtabtab + "TMessage msg = iprot.readMessageBegin();
"
		subscriber += tabtabtab + "if (!msg.name.equals(op)) {
"
		subscriber += tabtabtabtab + "TProtocolUtil.skip(iprot, TType.STRUCT);
"
		subscriber += tabtabtabtab + "iprot.readMessageEnd();
"
		subscriber += tabtabtabtab + "throw new TApplicationException(TApplicationException.UNKNOWN_METHOD);
"
		subscriber += tabtabtab + "}\n"
		subscriber += tabtabtab + fmt.Sprintf("%s req = new %s();
", g.qualifiedTypeName(op.Type), g.qualifiedTypeName(op.Type))
		subscriber += tabtabtab + "req.read(iprot);
"
		subscriber += tabtabtab + "iprot.readMessageEnd();
"
		subscriber += tabtabtab + "return req;
"
		subscriber += tabtab + "}\n\n"
	}

	subscriber += tab + "}\n"

	return subscriber
}

func (g *Generator) generateScopePrefixArgs(scope *parser.Scope) string {
	args := ""
	if len(scope.Prefix.Variables) > 0 {
		for _, variable := range scope.Prefix.Variables {
			args = fmt.Sprintf("%sString %s, ", args, variable)
		}
	}
	return args
}

func (g *Generator) GenerateService(file *os.File, s *parser.Service) error {
	contents := ""
	if g.includeGeneratedAnnotation() {
		contents += g.generatedAnnotation()
	}
	contents += fmt.Sprintf("public class F%s {

", s.Name)
	contents += g.generateServiceInterface(s)
	contents += g.generateClient(s)
	contents += g.generateServer(s)

	_, err := file.WriteString(contents)
	return err
}

func (g *Generator) generateServiceInterface(service *parser.Service) string {
	contents := ""
	if service.Comment != nil {
		contents += g.GenerateBlockComment(service.Comment, tab)
	}
	if service.Extends != "" {
		contents += tab + fmt.Sprintf("public interface Iface extends %s.Iface {

",
			g.getServiceExtendsName(service))
	} else {
		contents += tab + "public interface Iface {

"
	}
	for _, method := range service.Methods {
		if method.Comment != nil {
			contents += g.GenerateBlockComment(method.Comment, tabtab)
		}
		contents += fmt.Sprintf(tabtab+"public %s %s(FContext ctx%s) %s;

",
			g.generateReturnValue(method), method.Name, g.generateArgs(method.Arguments, false), g.generateExceptions(method.Exceptions))
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) getServiceExtendsName(service *parser.Service) string {
	serviceName := "F" + service.ExtendsService()
	include := service.ExtendsInclude()
	if include != "" {
		if inc, ok := g.Frugal.NamespaceForInclude(include, lang); ok {
			include = inc
		} else {
			return serviceName
		}
		serviceName = include + "." + serviceName
	}
	return serviceName
}

func (g *Generator) generateReturnValue(method *parser.Method) string {
	return g.generateContextualReturnValue(method, false)
}

func (g *Generator) generateBoxedReturnValue(method *parser.Method) string {
	return g.generateContextualReturnValue(method, true)
}

func (g *Generator) generateContextualReturnValue(method *parser.Method, boxed bool) string {
	if method.ReturnType == nil {
		ret := "void"
		if boxed {
			ret = "Void"
		}
		return ret
	}
	ret := g.getJavaTypeFromThriftType(method.ReturnType)
	if boxed {
		ret = containerType(ret)
	}
	return ret
}

func (g *Generator) generateArgs(args []*parser.Field, final bool) string {
	argStr := ""
	modifier := ""
	if final {
		modifier = "final "
	}
	for _, arg := range args {
		argStr += ", " + modifier + g.getJavaTypeFromThriftType(arg.Type) + " " + arg.Name
	}
	return argStr
}

func (g *Generator) generateClient(service *parser.Service) string {
	contents := ""
	if service.Extends != "" {
		contents += tab + fmt.Sprintf("public static class Client extends %s.Client implements Iface {

",
			g.getServiceExtendsName(service))
	} else {
		contents += tab + "public static class Client implements Iface {

"
	}
	if service.Extends == "" {
		contents += tabtab + "protected final Object writeLock = new Object();
"
		if g.generateAsync() {
			contents += tabtab + "protected ExecutorService asyncExecutor = Executors.newFixedThreadPool(2);
"
		}
	}
	contents += tabtab + "private Iface proxy;

"

	contents += tabtab + "public Client(FTransport transport, FProtocolFactory protocolFactory, ServiceMiddleware... middleware) {
"
	if service.Extends != "" {
		contents += tabtabtab + "super(transport, protocolFactory, middleware);
"
	}
	contents += tabtabtab + "Iface client = new InternalClient(transport, protocolFactory, writeLock);
"
	contents += tabtabtab + "proxy = InvocationHandler.composeMiddleware(client, Iface.class, middleware);
"
	contents += tabtab + "}\n\n"

	for _, method := range service.Methods {
		if method.Comment != nil {
			contents += g.GenerateBlockComment(method.Comment, tabtab)
		}
		contents += tabtab + fmt.Sprintf("public %s %s(FContext ctx%s) %s {
",
			g.generateReturnValue(method), method.Name, g.generateArgs(method.Arguments, false), g.generateExceptions(method.Exceptions))
		if method.ReturnType != nil {
			contents += tabtabtab + fmt.Sprintf("return proxy.%s(%s);
", method.Name, g.generateClientCallArgs(method.Arguments))
		} else {
			contents += tabtabtab + fmt.Sprintf("proxy.%s(%s);
", method.Name, g.generateClientCallArgs(method.Arguments))
		}
		contents += tabtab + "}\n\n"

		if g.generateAsync() {
			contents += g.generateAsyncClientMethod(service, method)
		}
	}
	contents += tab + "}\n\n"
	contents += g.generateInternalClient(service)
	return contents
}

func (g *Generator) generateAsyncClientMethod(service *parser.Service, method *parser.Method) string {
	contents := ""
	if method.Comment != nil {
		contents += g.GenerateBlockComment(method.Comment, tabtab)
	}
	contents += tabtab + fmt.Sprintf("public Future<%s> %sAsync(final FContext ctx%s) {
",
		g.generateBoxedReturnValue(method), method.Name, g.generateArgs(method.Arguments, true))
	contents += tabtabtab + fmt.Sprintf("return asyncExecutor.submit(new Callable<%s>() {
", g.generateBoxedReturnValue(method))
	contents += tabtabtabtab + fmt.Sprintf("public %s call() throws Exception {
", g.generateBoxedReturnValue(method))
	if method.ReturnType != nil {
		contents += tabtabtabtabtab + fmt.Sprintf("return %s(%s);
", method.Name, g.generateClientCallArgs(method.Arguments))
	} else {
		contents += tabtabtabtabtab + fmt.Sprintf("%s(%s);
", method.Name, g.generateClientCallArgs(method.Arguments))
		contents += tabtabtabtabtab + "return null;\n"
	}
	contents += tabtabtabtab + "}\n"
	contents += tabtabtab + "});
"
	contents += tabtab + "}\n\n"
	return contents
}

func (g *Generator) generateInternalClient(service *parser.Service) string {
	contents := ""
	if service.Extends != "" {
		contents += tab + fmt.Sprintf("private static class InternalClient extends %s.Client implements Iface {

",
			g.getServiceExtendsName(service))
	} else {
		contents += tab + "private static class InternalClient implements Iface {

"
	}

	contents += tabtab + "private FTransport transport;
"
	contents += tabtab + "private FProtocolFactory protocolFactory;
"
	contents += tabtab + "private FProtocol inputProtocol;
"
	contents += tabtab + "private FProtocol outputProtocol;
"
	contents += tabtab + "private final Object writeLock;

"

	contents += tabtab + "public InternalClient(FTransport transport, FProtocolFactory protocolFactory, Object writeLock) {
"
	if service.Extends != "" {
		contents += tabtabtab + "super(transport, protocolFactory);
"
	}
	contents += tabtabtab + "this.transport = transport;
"
	contents += tabtabtab + "this.transport.setRegistry(new FClientRegistry());
"
	contents += tabtabtab + "this.protocolFactory = protocolFactory;
"
	contents += tabtabtab + "this.inputProtocol = this.protocolFactory.getProtocol(this.transport);
"
	contents += tabtabtab + "this.outputProtocol = this.protocolFactory.getProtocol(this.transport);
"
	contents += tabtabtab + "this.writeLock = writeLock;
"
	contents += tabtab + "}\n\n"

	for _, method := range service.Methods {
		contents += g.generateClientMethod(service, method)
	}
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateClientMethod(service *parser.Service, method *parser.Method) string {
	servTitle := strings.Title(service.Name)

	contents := ""
	if method.Comment != nil {
		contents += g.GenerateBlockComment(method.Comment, tabtab)
	}
	contents += tabtab + fmt.Sprintf("public %s %s(FContext ctx%s) %s {
",
		g.generateReturnValue(method), method.Name, g.generateArgs(method.Arguments, false), g.generateExceptions(method.Exceptions))
	contents += tabtabtab + "FProtocol oprot = this.outputProtocol;
"
	indent := tabtabtab
	if !method.Oneway {
		contents += tabtabtab + "BlockingQueue<Object> result = new ArrayBlockingQueue<>(1);
"
		contents += tabtabtab + fmt.Sprintf("this.transport.register(ctx, recv%sHandler(ctx, result));
", strings.Title(method.Name))
		contents += tabtabtab + "try {\n"
		indent += tab
	}
	contents += indent + "synchronized (writeLock) {
"
	contents += indent + tab + "oprot.writeRequestHeader(ctx);
"
	msgType := "CALL"
	if method.Oneway {
		msgType = "ONEWAY"
	}
	contents += indent + tab + fmt.Sprintf("oprot.writeMessageBegin(new TMessage(\"%s\", TMessageType.%s, 0));
", method.Name, msgType)
	contents += indent + tab + fmt.Sprintf("%s.%s_args args = new %s.%s_args();
", servTitle, method.Name, servTitle, method.Name)
	for _, arg := range method.Arguments {
		contents += indent + tab + fmt.Sprintf("args.set%s(%s);
", strings.Title(arg.Name), arg.Name)
	}
	contents += indent + tab + "args.write(oprot);
"
	contents += indent + tab + "oprot.writeMessageEnd();
"
	contents += indent + tab + "oprot.getTransport().flush();
"
	contents += indent + "}\n"

	if method.Oneway {
		contents += tabtab + "}\n"
		return contents
	}

	contents += "\n"
	contents += tabtabtabtab + "Object res = null;
"
	contents += tabtabtabtab + "try {\n"
	contents += tabtabtabtabtab + "res = result.poll(ctx.getTimeout(), TimeUnit.MILLISECONDS);
"
	contents += tabtabtabtab + "} catch (InterruptedException e) {
"
	contents += tabtabtabtabtab + fmt.Sprintf(
		"throw new TApplicationException(TApplicationException.INTERNAL_ERROR, \"%s interrupted: \" + e.getMessage());
",
		method.Name)
	contents += tabtabtabtab + "}\n"
	contents += tabtabtabtab + "if (res == null) {
"
	contents += tabtabtabtabtab + fmt.Sprintf("throw new FTimeoutException(\"%s timed out\");\n", method.Name)
	contents += tabtabtabtab + "}\n"
	contents += tabtabtabtab + "if (res instanceof TException) {
"
	contents += tabtabtabtabtab + "throw (TException) res;
"
	contents += tabtabtabtab + "}\n"
	contents += tabtabtabtab + fmt.Sprintf("%s.%s_result r = (%s.%s_result) res;
", servTitle, method.Name, servTitle, method.Name)
	if method.ReturnType != nil {
		contents += tabtabtabtab + "if (r.isSetSuccess()) {
"
		contents += tabtabtabtabtab + "return r.success;
"
		contents += tabtabtabtab + "}\n"
	}
	for _, exception := range method.Exceptions {
		contents += tabtabtabtab + fmt.Sprintf("if (r.%s != null) {
", exception.Name)
		contents += tabtabtabtabtab + fmt.Sprintf("throw r.%s;
", exception.Name)
		contents += tabtabtabtab + "}\n"
	}
	if method.ReturnType != nil {
		contents += tabtabtabtab + fmt.Sprintf(
			"throw new TApplicationException(TApplicationException.MISSING_RESULT, \"%s failed: unknown result\");\n",
			method.Name)
	}
	contents += tabtabtab + "} finally {
"
	contents += tabtabtabtab + "this.transport.unregister(ctx);
"
	contents += tabtabtab + "}\n"
	contents += tabtab + "}\n\n"

	contents += tabtab + fmt.Sprintf(
		"private FAsyncCallback recv%sHandler(final FContext ctx, final BlockingQueue<Object> result) {
",
		strings.Title(method.Name))
	contents += tabtabtab + "return new FAsyncCallback() {
"
	contents += tabtabtabtab + "public void onMessage(TTransport tr) throws TException {
"
	contents += tabtabtabtabtab + "FProtocol iprot = InternalClient.this.protocolFactory.getProtocol(tr);
"
	contents += tabtabtabtabtab + "try {\n"
	contents += tabtabtabtabtabtab + "iprot.readResponseHeader(ctx);
"
	contents += tabtabtabtabtabtab + "TMessage message = iprot.readMessageBegin();
"
	contents += tabtabtabtabtabtab + fmt.Sprintf("if (!message.name.equals(\"%s\")) {
", method.Name)
	contents += tabtabtabtabtabtabtab + fmt.Sprintf(
		"throw new TApplicationException(TApplicationException.WRONG_METHOD_NAME, \"%s failed: wrong method name\");\n",
		method.Name)
	contents += tabtabtabtabtabtab + "}\n"
	contents += tabtabtabtabtabtab + "if (message.type == TMessageType.EXCEPTION) {
"
	contents += tabtabtabtabtabtabtab + "TApplicationException e = TApplicationException.read(iprot);
"
	contents += tabtabtabtabtabtabtab + "iprot.readMessageEnd();
"
	contents += tabtabtabtabtabtabtab + "if (e.getType() == FTransport.RESPONSE_TOO_LARGE) {
"
	contents += tabtabtabtabtabtabtabtab + "FMessageSizeException ex = new FMessageSizeException(FTransport.RESPONSE_TOO_LARGE, \"response too large for transport\");\n"
	contents += tabtabtabtabtabtabtabtab + "try {\n"
	contents += tabtabtabtabtabtabtabtabtab + "result.put(ex);
"
	contents += tabtabtabtabtabtabtabtabtab + "return;
"
	contents += tabtabtabtabtabtabtabtab + "} catch (InterruptedException ie) {
"
	contents += tabtabtabtabtabtabtabtabtab + fmt.Sprintf(
		"throw new TApplicationException(TApplicationException.INTERNAL_ERROR, \"%s interrupted: \" + ie.getMessage());
",
		method.Name)
	contents += tabtabtabtabtabtabtabtab + "}\n"
	contents += tabtabtabtabtabtabtab + "}\n"
	contents += tabtabtabtabtabtabtab + "try {\n"
	contents += tabtabtabtabtabtabtabtab + "result.put(e);
"
	contents += tabtabtabtabtabtabtab + "} finally {
"
	contents += tabtabtabtabtabtabtabtab + "throw e;
"
	contents += tabtabtabtabtabtabtab + "}\n"
	contents += tabtabtabtabtabtab + "}\n"
	contents += tabtabtabtabtabtab + "if (message.type != TMessageType.REPLY) {
"
	contents += tabtabtabtabtabtabtab + fmt.Sprintf(
		"throw new TApplicationException(TApplicationException.INVALID_MESSAGE_TYPE, \"%s failed: invalid message type\");\n",
		method.Name)
	contents += tabtabtabtabtabtab + "}\n"
	contents += tabtabtabtabtabtab + fmt.Sprintf("%s.%s_result res = new %s.%s_result();
", servTitle, method.Name, servTitle, method.Name)
	contents += tabtabtabtabtabtab + "res.read(iprot);
"
	contents += tabtabtabtabtabtab + "iprot.readMessageEnd();
"
	contents += tabtabtabtabtabtab + "try {\n"
	contents += tabtabtabtabtabtabtab + "result.put(res);
"
	contents += tabtabtabtabtabtab + "} catch (InterruptedException e) {
"
	contents += tabtabtabtabtabtabtab + fmt.Sprintf(
		"throw new TApplicationException(TApplicationException.INTERNAL_ERROR, \"%s interrupted: \" + e.getMessage());
",
		method.Name)
	contents += tabtabtabtabtabtab + "}\n"
	contents += tabtabtabtabtab + "} catch (TException e) {
"
	contents += tabtabtabtabtabtab + "try {\n"
	contents += tabtabtabtabtabtabtab + "result.put(e);
"
	contents += tabtabtabtabtabtab + "} finally {
"
	contents += tabtabtabtabtabtabtab + "throw e;
"
	contents += tabtabtabtabtabtab + "}\n"
	contents += tabtabtabtabtab + "}\n"
	contents += tabtabtabtab + "}\n"
	contents += tabtabtab + "};
"
	contents += tabtab + "}\n\n"

	return contents
}

func (g *Generator) generateExceptions(exceptions []*parser.Field) string {
	contents := "throws TException"
	for _, exception := range exceptions {
		contents += ", " + g.getJavaTypeFromThriftType(exception.Type)
	}
	return contents
}

func (g *Generator) generateServer(service *parser.Service) string {
	servTitle := strings.Title(service.Name)

	contents := ""
	extends := "FBaseProcessor"
	if service.Extends != "" {
		extends = g.getServiceExtendsName(service) + ".Processor"
	}
	contents += tab + fmt.Sprintf("public static class Processor extends %s implements FProcessor {

", extends)

	contents += tabtab + "public Processor(Iface iface, ServiceMiddleware... middleware) {
"
	if service.Extends != "" {
		contents += tabtabtab + "super(iface, getProcessMap(iface, new java.util.HashMap<String, FProcessorFunction>(), middleware), middleware);
"
	} else {
		contents += tabtabtab + "super(getProcessMap(iface, new java.util.HashMap<String, FProcessorFunction>(), middleware));
"
	}
	contents += tabtab + "}\n\n"

	contents += tabtab + "protected Processor(Iface iface, java.util.Map<String, FProcessorFunction> processMap, ServiceMiddleware[] middleware) {
"
	if service.Extends != "" {
		contents += tabtabtab + "super(iface, getProcessMap(iface, processMap, middleware), middleware);
"
	} else {
		contents += tabtabtab + "super(getProcessMap(iface, processMap, middleware));
"
	}
	contents += tabtab + "}\n\n"

	contents += tabtab + "private static java.util.Map<String, FProcessorFunction> getProcessMap(Iface handler, java.util.Map<String, FProcessorFunction> processMap, ServiceMiddleware[] middleware) {
"
	contents += tabtabtab + "handler = InvocationHandler.composeMiddleware(handler, Iface.class, middleware);
"
	for _, method := range service.Methods {
		contents += tabtabtab + fmt.Sprintf("processMap.put(\"%s\", new %s(handler));
", method.Name, strings.Title(method.Name))
	}
	contents += tabtabtab + "return processMap;
"
	contents += tabtab + "}\n\n"

	for _, method := range service.Methods {
		contents += tabtab + fmt.Sprintf("private static class %s implements FProcessorFunction {

", strings.Title(method.Name))

		contents += tabtabtab + "private Iface handler;

"

		contents += tabtabtab + fmt.Sprintf("public %s(Iface handler) {
", strings.Title(method.Name))
		contents += tabtabtabtab + "this.handler = handler;
"
		contents += tabtabtab + "}\n\n"

		contents += tabtabtab + "public void process(FContext ctx, FProtocol iprot, FProtocol oprot) throws TException {
"
		contents += tabtabtabtab + fmt.Sprintf("%s.%s_args args = new %s.%s_args();
", servTitle, method.Name, servTitle, method.Name)
		contents += tabtabtabtab + "try {\n"
		contents += tabtabtabtabtab + "args.read(iprot);
"
		contents += tabtabtabtab + "} catch (TException e) {
"
		contents += tabtabtabtabtab + "iprot.readMessageEnd();
"
		if !method.Oneway {
			contents += tabtabtabtabtab + "synchronized (WRITE_LOCK) {
"
			contents += tabtabtabtabtabtab + fmt.Sprintf("writeApplicationException(ctx, oprot, TApplicationException.PROTOCOL_ERROR, \"%s\", e.getMessage());
", method.Name)
			contents += tabtabtabtabtab + "}\n"
		}
		contents += tabtabtabtabtab + "throw e;
"
		contents += tabtabtabtab + "}\n\n"

		contents += tabtabtabtab + "iprot.readMessageEnd();
"

		if method.Oneway {
			contents += tabtabtabtab + fmt.Sprintf("this.handler.%s(%s);
", method.Name, g.generateServerCallArgs(method.Arguments))
			contents += tabtabtab + "}\n"
			contents += tabtab + "}\n\n"
			continue
		}

		contents += tabtabtabtab + fmt.Sprintf("%s.%s_result result = new %s.%s_result();
", servTitle, method.Name, servTitle, method.Name)
		contents += tabtabtabtab + "try {\n"
		if method.ReturnType == nil {
			contents += tabtabtabtabtab + fmt.Sprintf("this.handler.%s(%s);
", method.Name, g.generateServerCallArgs(method.Arguments))
		} else {
			contents += tabtabtabtabtab + fmt.Sprintf("result.success = this.handler.%s(%s);
", method.Name, g.generateServerCallArgs(method.Arguments))
			contents += tabtabtabtabtab + "result.setSuccessIsSet(true);
"
		}
		for _, exception := range method.Exceptions {
			contents += tabtabtabtab + fmt.Sprintf("} catch (%s %s) {
", g.getJavaTypeFromThriftType(exception.Type), exception.Name)
			contents += tabtabtabtabtab + fmt.Sprintf("result.%s = %s;
", exception.Name, exception.Name)
		}
		contents += tabtabtabtab + "} catch (TException e) {
"
		contents += tabtabtabtabtab + "synchronized (WRITE_LOCK) {
"
		contents += tabtabtabtabtabtab + fmt.Sprintf(
			"writeApplicationException(ctx, oprot, TApplicationException.INTERNAL_ERROR, \"%s\", \"Internal error processing %s: \" + e.getMessage());
",
			method.Name, method.Name)
		contents += tabtabtabtabtab + "}\n"
		contents += tabtabtabtabtab + "throw e;
"
		contents += tabtabtabtab + "}\n"
		contents += tabtabtabtab + "synchronized (WRITE_LOCK) {
"
		contents += tabtabtabtabtab + "try {\n"
		contents += tabtabtabtabtabtab + "oprot.writeResponseHeader(ctx);
"
		contents += tabtabtabtabtabtab + fmt.Sprintf("oprot.writeMessageBegin(new TMessage(\"%s\", TMessageType.REPLY, 0));
", method.Name)
		contents += tabtabtabtabtabtab + "result.write(oprot);
"
		contents += tabtabtabtabtabtab + "oprot.writeMessageEnd();
"
		contents += tabtabtabtabtabtab + "oprot.getTransport().flush();
"
		contents += tabtabtabtabtab + "} catch (TException e) {
"
		contents += tabtabtabtabtabtab + "if (e instanceof FMessageSizeException) {
"
		contents += tabtabtabtabtabtabtab + fmt.Sprintf(
			"writeApplicationException(ctx, oprot, FTransport.RESPONSE_TOO_LARGE, \"%s\", \"response too large: \" + e.getMessage());
",
			method.Name)
		contents += tabtabtabtabtabtab + "} else {\n"
		contents += tabtabtabtabtabtabtab + "throw e;
"
		contents += tabtabtabtabtabtab + "}\n"
		contents += tabtabtabtabtab + "}\n"
		contents += tabtabtabtab + "}\n"
		contents += tabtabtab + "}\n"
		contents += tabtab + "}\n\n"
	}

	contents += tabtab + "private static void writeApplicationException(FContext ctx, FProtocol oprot, int type, String method, String message) throws TException {
"
	contents += tabtabtab + "TApplicationException x = new TApplicationException(type, message);
"
	contents += tabtabtab + "oprot.writeResponseHeader(ctx);
"
	contents += tabtabtab + "oprot.writeMessageBegin(new TMessage(method, TMessageType.EXCEPTION, 0));
"
	contents += tabtabtab + "x.write(oprot);
"
	contents += tabtabtab + "oprot.writeMessageEnd();
"
	contents += tabtabtab + "oprot.getTransport().flush();
"
	contents += tabtab + "}\n\n"

	contents += tab + "}\n\n"

	contents += "}"

	return contents
}

func (g *Generator) generateScopeArgs(scope *parser.Scope) string {
	args := "ctx"
	for _, v := range scope.Prefix.Variables {
		args += ", " + v
	}
	args += ", req"
	return args
}

func (g *Generator) generateClientCallArgs(args []*parser.Field) string {
	return g.generateCallArgs(args, "")
}

func (g *Generator) generateServerCallArgs(args []*parser.Field) string {
	return g.generateCallArgs(args, "args.")
}

func (g *Generator) generateCallArgs(args []*parser.Field, prefix string) string {
	contents := "ctx"
	prefix = ", " + prefix
	for _, arg := range args {
		contents += prefix + arg.Name
	}
	return contents
}

func (g *Generator) getJavaTypeFromThriftType(t *parser.Type) string {
	return g._getJavaType(t, true)
}

func (g *Generator) getUnparametrizedJavaType(t *parser.Type) string {
	return g._getJavaType(t, false)
}

func (g *Generator) _getJavaType(t *parser.Type, parametrized bool) string {
	if t == nil {
		return "void"
	}
	underlyingType := g.Frugal.UnderlyingType(t)
	switch underlyingType.Name {
	case "bool":
		return "boolean"
	case "byte", "i8":
		return "byte"
	case "i16":
		return "short"
	case "i32":
		return "int"
	case "i64":
		return "long"
	case "double":
		return "double"
	case "string":
		return "String"
	case "binary":
		return "java.nio.ByteBuffer"
	case "list":
		if parametrized {
			return fmt.Sprintf("java.util.List<%s>",
				containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
		}
		return "java.util.List"
	case "set":
		if parametrized {
			return fmt.Sprintf("java.util.Set<%s>",
				containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
		}
		return "java.util.Set"
	case "map":
		if parametrized {
			return fmt.Sprintf("java.util.Map<%s, %s>",
				containerType(g.getJavaTypeFromThriftType(underlyingType.KeyType)),
				containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
		}
		return "java.util.Map"
	default:
		// This is a custom type, return a pointer to it
		return g.qualifiedTypeName(t)
	}
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
			panic("shouldn't happen: " + underlyingType.Name)
		}
	}

	return fmt.Sprintf("org.apache.thrift.protocol.TType.%s", ttype)
}

func (g *Generator) isJavaPrimitive(t *parser.Type) bool {
	underlyingType := g.Frugal.UnderlyingType(t)
	switch underlyingType.Name {
	case "bool", "byte", "i8", "i16", "i32", "i64", "double":
		return true
	default:
		return false
	}
}

func containerType(typeName string) string {
	switch typeName {
	case "int":
		return "Integer"
	case "boolean", "byte", "short", "long", "double", "void":
		return strings.Title(typeName)
	default:
		return typeName
	}
}

func (g *Generator) qualifiedTypeName(t *parser.Type) string {
	param := t.ParamName()
	include := t.IncludeName()
	if include != "" {
		namespace, ok := g.Frugal.NamespaceForInclude(include, lang)
		if ok {
			return fmt.Sprintf("%s.%s", namespace, param)
		}
	}
	return param
}

func toConstantName(name string) string {
	// TODO fix for identifiers like "ID2"
	ret := ""
	tmp := []rune(name)
	is_prev_lc := true
	is_current_lc := tmp[0] == unicode.ToLower(tmp[0])
	is_next_lc := false

	for i, _ := range tmp {
		lc := unicode.ToLower(tmp[i])

		if i == len(name)-1 {
			is_next_lc = false
		} else {
			is_next_lc = (tmp[i+1] == unicode.ToLower(tmp[i+1]))
		}

		if i != 0 && !is_current_lc && (is_prev_lc || is_next_lc) {
			ret += "_"
		}
		ret += string(lc)

		is_prev_lc = is_current_lc
		is_current_lc = is_next_lc
	}
	//return ret
	return strings.ToUpper(ret)
}

func (g *Generator) includeGeneratedAnnotation() bool {
	return g.Options[generatedAnnotations] != "suppress"
}

func (g *Generator) generatedAnnotation() string {
	anno := fmt.Sprintf("@Generated(value = \"Autogenerated by Frugal Compiler (%s)\"", globals.Version)
	if g.Options[generatedAnnotations] != "undated" {
		anno += fmt.Sprintf(", "+"date = \"%s\"", g.time.Format("2006-1-2"))
	}
	anno += ")
"
	return anno
}

func (g *Generator) generateAsync() bool {
	_, ok := g.Options["async"]
	return ok
}