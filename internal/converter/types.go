package converter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	protoc_gen_jsonschema "github.com/MytraAI/protoc-gen-jsonschema"
	protoc_gen_validate "github.com/envoyproxy/protoc-gen-validate/validate"
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/xeipuuv/gojsonschema"
	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

var (
	globalPkg = newProtoPackage(nil, "")

	wellKnownTypes = map[string]bool{
		"BoolValue":   true,
		"BytesValue":  true,
		"DoubleValue": true,
		"Duration":    true,
		"FloatValue":  true,
		"Int32Value":  true,
		"Int64Value":  true,
		"ListValue":   true,
		"StringValue": true,
		"Struct":      true,
		"UInt32Value": true,
		"UInt64Value": true,
		"Value":       true,
	}
)

func (c *Converter) registerEnum(pkgName string, enum *descriptor.EnumDescriptorProto) {
	pkg := globalPkg
	if pkgName != "" {
		for _, node := range strings.Split(pkgName, ".") {
			if pkg == globalPkg && node == "" {
				// Skips leading "."
				continue
			}
			child, ok := pkg.children[node]
			if !ok {
				child = newProtoPackage(pkg, node)
				pkg.children[node] = child
			}
			pkg = child
		}
	}
	pkg.enums[enum.GetName()] = enum
}

func (c *Converter) registerType(pkgName string, msgDesc *descriptor.DescriptorProto) {
	pkg := globalPkg
	if pkgName != "" {
		for _, node := range strings.Split(pkgName, ".") {
			if pkg == globalPkg && node == "" {
				// Skips leading "."
				continue
			}
			child, ok := pkg.children[node]
			if !ok {
				child = newProtoPackage(pkg, node)
				pkg.children[node] = child
			}
			pkg = child
		}
	}
	pkg.types[msgDesc.GetName()] = msgDesc
}

// Convert a proto "field" (essentially a type-switch with some recursion):
func (c *Converter) convertField(curPkg *ProtoPackage, desc *descriptor.FieldDescriptorProto, msgDesc *descriptor.DescriptorProto, duplicatedMessages map[*descriptor.DescriptorProto]string, messageFlags ConverterFlags) (*jsonschema.Schema, error) {

	// Prepare a new jsonschema.Schema for our eventual return value:
	jsonSchemaType := &jsonschema.Schema{}

	// Switch the types, and pick a JSONSchema equivalent:
	switch desc.GetType() {

	// Float32:
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
		descriptor.FieldDescriptorProto_TYPE_FLOAT:
		if messageFlags.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Schema{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_NUMBER},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_NUMBER
		}

	// Int32:
	case descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SINT32:
		if messageFlags.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Schema{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_INTEGER},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_INTEGER
		}

	// Int64:
	case descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_SINT64:

		// As integer:
		if c.Flags.DisallowBigIntsAsStrings {
			if messageFlags.AllowNullValues {
				jsonSchemaType.OneOf = []*jsonschema.Schema{
					{Type: gojsonschema.TYPE_INTEGER},
					{Type: gojsonschema.TYPE_NULL},
				}
			} else {
				jsonSchemaType.Type = gojsonschema.TYPE_INTEGER
			}
		}

		// As string:
		if !c.Flags.DisallowBigIntsAsStrings {
			if messageFlags.AllowNullValues {
				jsonSchemaType.OneOf = []*jsonschema.Schema{
					{Type: gojsonschema.TYPE_STRING},
					{Type: gojsonschema.TYPE_NULL},
				}
			} else {
				jsonSchemaType.Type = gojsonschema.TYPE_STRING
			}
		}

	// String:
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		stringDef := &jsonschema.Schema{Type: gojsonschema.TYPE_STRING}

		// Custom field options from protoc-gen-jsonschema:
		if opt := proto.GetExtension(desc.GetOptions(), protoc_gen_jsonschema.E_FieldOptions); opt != nil {
			if fieldOptions, ok := opt.(*protoc_gen_jsonschema.FieldOptions); ok {
				if fieldOptions.GetMinLength() != 0 { //exclude default value
					stringDef.MinLength = proto.Uint64(uint64(fieldOptions.GetMinLength()))
				}
				if fieldOptions.GetMaxLength() != 0 { //exclude default value
					stringDef.MaxLength = proto.Uint64(uint64(fieldOptions.GetMaxLength()))
				}
				if fieldOptions.GetPattern() != "" { //exclude default value
					stringDef.Pattern = fieldOptions.GetPattern()
				}
			}
		}

		// Custom field options from protoc-gen-validate:
		if opt := proto.GetExtension(desc.GetOptions(), protoc_gen_validate.E_Rules); opt != nil {
			if fieldRules, ok := opt.(*protoc_gen_validate.FieldRules); fieldRules != nil && ok {
				if stringRules := fieldRules.GetString_(); stringRules != nil {
					stringDef.MaxLength = proto.Uint64(stringRules.GetMaxLen())
					stringDef.MinLength = proto.Uint64(stringRules.GetMinLen())
					stringDef.Pattern = stringRules.GetPattern()
				}
			}
		}

		if messageFlags.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Schema{
				{Type: gojsonschema.TYPE_NULL},
				stringDef,
			}
		} else {
			jsonSchemaType.Type = stringDef.Type
			jsonSchemaType.MaxLength = stringDef.MaxLength
			jsonSchemaType.MinLength = stringDef.MinLength
			jsonSchemaType.Pattern = stringDef.Pattern
		}
	// Bytes:
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		if messageFlags.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Schema{
				{Type: gojsonschema.TYPE_NULL},
				{
					Type:            gojsonschema.TYPE_STRING,
					ContentEncoding: "base64",
				},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
			jsonSchemaType.ContentEncoding = "base64"
		}

	// ENUM:
	case descriptor.FieldDescriptorProto_TYPE_ENUM:

		// Go through all the enums we have, see if we can match any to this field.
		fullEnumIdentifier := strings.TrimPrefix(desc.GetTypeName(), ".")
		matchedEnum, _, ok := c.lookupEnum(curPkg, fullEnumIdentifier)
		if !ok {
			return nil, fmt.Errorf("unable to resolve enum type: %s", desc.GetType().String())
		}

		// We already have a converter for standalone ENUMs, so just use that:
		enumSchema, err := c.convertEnumType(matchedEnum, messageFlags)
		if err != nil {
			switch err {
			case errIgnored:
			default:
				return nil, err
			}
		}

		jsonSchemaType = &enumSchema

	// Bool:
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if messageFlags.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Schema{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_BOOLEAN},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_BOOLEAN
		}

	// Group (object):
	case descriptor.FieldDescriptorProto_TYPE_GROUP, descriptor.FieldDescriptorProto_TYPE_MESSAGE:

		switch desc.GetTypeName() {
		// Make sure that durations match a particular string pattern (eg 3.4s):
		case ".google.protobuf.Duration":
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
			jsonSchemaType.Format = "regex"
			jsonSchemaType.Pattern = `^([0-9]+\.?[0-9]*|\.[0-9]+)s$`
		case ".google.protobuf.Timestamp":
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
			jsonSchemaType.Format = "date-time"
		case ".google.protobuf.Value", ".google.protobuf.Struct":
			jsonSchemaType.Type = gojsonschema.TYPE_OBJECT
			jsonSchemaType.AdditionalProperties = jsonschema.TrueSchema
		default:
			jsonSchemaType.Type = gojsonschema.TYPE_OBJECT
			if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL {
				jsonSchemaType.AdditionalProperties = jsonschema.TrueSchema
			}
			if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
				jsonSchemaType.AdditionalProperties = jsonschema.FalseSchema
			}
			if messageFlags.DisallowAdditionalProperties {
				jsonSchemaType.AdditionalProperties = jsonschema.FalseSchema
			}
		}

	default:
		return nil, fmt.Errorf("unrecognized field type: %s", desc.GetType().String())
	}

	// Recurse basic array:
	if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED && jsonSchemaType.Type != gojsonschema.TYPE_OBJECT {
		jsonSchemaType.Items = &jsonschema.Schema{}

		// Add Comments and parameters
		src := c.sourceInfo.GetField(desc)
		title, comment, params := c.formatTitleAndDescription(nil, src)
		jsonSchemaType.Title = title
		jsonSchemaType.Description = comment
		if _, ok := params["readOnly"]; ok {
			jsonSchemaType.ReadOnly = true
		}
		if val, ok := params["maxLength"]; ok {
			if v, err := strconv.ParseUint(val, 10, 64); err == nil {
				jsonSchemaType.MaxLength = &v
			}
		}
		if val, ok := params["minLength"]; ok {
			if v, err := strconv.ParseUint(val, 10, 64); err == nil {
				jsonSchemaType.MinLength = &v
			}
		}
		// Custom field options from protoc-gen-validate:
		if opt := proto.GetExtension(desc.GetOptions(), protoc_gen_validate.E_Rules); opt != nil {
			if fieldRules, ok := opt.(*protoc_gen_validate.FieldRules); fieldRules != nil && ok {
				if repeatedRules := fieldRules.GetRepeated(); repeatedRules != nil {
					jsonSchemaType.MaxItems = proto.Uint64(uint64(repeatedRules.GetMaxItems()))
					jsonSchemaType.MinItems = proto.Uint64(uint64(repeatedRules.GetMinItems()))
				}
			}
		}

		if len(jsonSchemaType.Enum) > 0 {
			jsonSchemaType.Items.Enum = jsonSchemaType.Enum
			jsonSchemaType.Enum = nil
			jsonSchemaType.Items.OneOf = nil
		} else {
			jsonSchemaType.Items.Type = jsonSchemaType.Type
			jsonSchemaType.Items.OneOf = jsonSchemaType.OneOf
		}

		if messageFlags.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Schema{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_ARRAY},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_ARRAY
			jsonSchemaType.OneOf = []*jsonschema.Schema{}
		}
		return jsonSchemaType, nil
	}

	// Recurse nested objects / arrays of objects (if necessary):
	if jsonSchemaType.Type == gojsonschema.TYPE_OBJECT {

		recordType, pkgName, ok := c.lookupType(curPkg, desc.GetTypeName())
		if !ok {
			return nil, fmt.Errorf("no such message type named %s", desc.GetTypeName())
		}

		// Check if this type is defined in a dependency file
		// Skip Google protobuf well-known types as they should be handled specially
		typeName := strings.TrimPrefix(desc.GetTypeName(), ".")
		if !strings.HasPrefix(typeName, "google.protobuf.") {
			if isDependency, refFile := c.isMessageDefinedInDependency(recordType.GetName()); isDependency {
				if c.dependenciesAsExtRefs {
					// When dependencies_as_external_refs=true, use external file references
					refSchema := &jsonschema.Schema{
						Ref:  refFile + "#/$defs/" + recordType.GetName(),
						Type: "", // Clear type when using $ref
					}

					// Set additionalProperties based on field label and flags (same logic as for inline types)
					if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL {
						refSchema.AdditionalProperties = jsonschema.TrueSchema
					}
					if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
						refSchema.AdditionalProperties = jsonschema.FalseSchema
					}
					if messageFlags.DisallowAdditionalProperties {
						refSchema.AdditionalProperties = jsonschema.FalseSchema
					}

					// For repeated fields, we need to wrap the reference in array schema
					if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
						jsonSchemaType.Items = refSchema
						jsonSchemaType.Type = gojsonschema.TYPE_ARRAY
						return jsonSchemaType, nil
					}

					// For non-repeated fields, return the reference directly
					return refSchema, nil
				}
				// When dependencies_as_external_refs=false, dependencies will be included in $defs, so continue with normal processing
			}
		}

		// Recurse the recordType:
		recursedJSONSchemaType, err := c.recursiveConvertMessageType(curPkg, recordType, pkgName, duplicatedMessages, false)
		if err != nil {
			return nil, err
		}

		// Maps, arrays, and objects are structured in different ways:
		switch {

		// Maps:
		case recordType.Options.GetMapEntry():
			c.logger.
				WithField("field_name", recordType.GetName()).
				WithField("msgDesc_name", *msgDesc.Name).
				Tracef("Is a map")

			if recursedJSONSchemaType.Properties == nil {
				return nil, fmt.Errorf("Unable to find properties of MAP type")
			}

			// Make sure we have a "value":
			value, valuePresent := recursedJSONSchemaType.Properties.Get("value")
			if !valuePresent {
				return nil, fmt.Errorf("Unable to find 'value' property of MAP type")
			}

			jsonSchemaType.AdditionalProperties = value

		// Arrays:
		case desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED:
			jsonSchemaType.Items = recursedJSONSchemaType
			jsonSchemaType.Type = gojsonschema.TYPE_ARRAY

			// Build up the list of required fields:
			if messageFlags.AllFieldsRequired && len(recursedJSONSchemaType.OneOf) == 0 && recursedJSONSchemaType.Properties != nil {
				for pair := recursedJSONSchemaType.Properties.Oldest(); pair != nil; pair = pair.Next() {
					jsonSchemaType.Items.Required = append(jsonSchemaType.Items.Required, pair.Key)
				}
			}
			jsonSchemaType.Items.Required = dedupe(jsonSchemaType.Items.Required)

		// Not maps, not arrays:
		default:

			// If we've got optional types then just take those:
			if recursedJSONSchemaType.OneOf != nil {
				return recursedJSONSchemaType, nil
			}

			// If we're not an object then set the type from whatever we recursed:
			if recursedJSONSchemaType.Type != gojsonschema.TYPE_OBJECT {
				jsonSchemaType.Type = recursedJSONSchemaType.Type
			}

			// Assume the attrbutes of the recursed value:
			jsonSchemaType.Properties = recursedJSONSchemaType.Properties
			jsonSchemaType.Ref = recursedJSONSchemaType.Ref
			jsonSchemaType.Required = recursedJSONSchemaType.Required

			// Build up the list of required fields:
			if messageFlags.AllFieldsRequired && len(recursedJSONSchemaType.OneOf) == 0 && recursedJSONSchemaType.Properties != nil {
				//for _, property := range recursedJSONSchemaType.Properties.Keys() {
				for pair := recursedJSONSchemaType.Properties.Oldest(); pair != nil; pair = pair.Next() {
					jsonSchemaType.Required = append(jsonSchemaType.Required, pair.Key)
				}
			}
		}

		// Optionally allow NULL values:
		if messageFlags.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Schema{
				{Type: gojsonschema.TYPE_NULL},
				{Type: jsonSchemaType.Type, Items: jsonSchemaType.Items},
			}
			jsonSchemaType.Type = ""
			jsonSchemaType.Items = nil
		}
	}

	jsonSchemaType.Required = dedupe(jsonSchemaType.Required)

	// Generate a description from src comments (if available)
	if src := c.sourceInfo.GetField(desc); src != nil {
		title, comment, params := c.formatTitleAndDescription(nil, src)
		jsonSchemaType.Title = title
		jsonSchemaType.Description = comment
		if jsonSchemaType.Type == gojsonschema.TYPE_INTEGER || jsonSchemaType.Type == gojsonschema.TYPE_NUMBER {
			if val, ok := params["minimum"]; ok {
				jsonSchemaType.Minimum = json.Number(val)
			}
			if val, ok := params["maximum"]; ok {
				jsonSchemaType.Maximum = json.Number(val)
			}
		}
		if jsonSchemaType.Type == gojsonschema.TYPE_STRING {
			if val, ok := params["maxLength"]; ok {
				if v, err := strconv.ParseUint(val, 10, 64); err == nil {
					jsonSchemaType.MaxLength = &v
				}
			}
			if val, ok := params["minLength"]; ok {
				if v, err := strconv.ParseUint(val, 10, 64); err == nil {
					jsonSchemaType.MinLength = &v
				}
			}
		}
		if val, ok := params["default"]; ok {
			if jsonSchemaType.Type == "integer" || jsonSchemaType.Type == "number" {
				jsonSchemaType.Default = json.Number(val)
			} else {
				jsonSchemaType.Default = val
			}
		}
		if _, ok := params["readOnly"]; ok {
			jsonSchemaType.ReadOnly = true
		}
	}

	return jsonSchemaType, nil
}

// Converts a proto "MESSAGE" into a JSON-Schema:
func (c *Converter) convertMessageType(curPkg *ProtoPackage, msgDesc *descriptor.DescriptorProto) (*jsonschema.Schema, error) {
	// Get a list of any nested messages in our schema:
	duplicatedMessages, err := c.findNestedMessages(curPkg, msgDesc)
	if err != nil {
		return nil, err
	}
	// Build up a list of JSONSchema type definitions for every message:
	definitions := jsonschema.Definitions{}
	var msgName = msgDesc.GetName()
	for refmsgDesc, nameWithPackage := range duplicatedMessages {
		var typeName string
		if c.Flags.TypeNamesWithNoPackage {
			typeName = refmsgDesc.GetName()
		} else {
			typeName = nameWithPackage
			if refmsgDesc.GetName() == msgName {
				msgName = typeName
			}
		}
		refType, err := c.recursiveConvertMessageType(curPkg, refmsgDesc, "", duplicatedMessages, true)
		if err != nil {
			return nil, err
		}

		// Add the schema to our definitions:
		definitions[typeName] = refType
	}

	// Put together a JSON schema with our discovered definitions, and a $ref for the root type:
	//var currentMsg = definitions[msgName]
	//delete(definitions, msgName)
	//newJSONSchema =currentMsg

	newJSONSchema := &jsonschema.Schema{
		Version:     c.schemaVersion,
		Definitions: definitions,
		Ref:         c.refPrefix + msgName,
	}
	//c.logger.WithField("msgname", definitions[msgDesc.GetName()]).Error("aaa")
	return newJSONSchema, nil
}

// findNestedMessages takes a message, and returns a map mapping pointers to messages nested within it:
// these messages become definitions which can be referenced (instead of repeating them every time they're used)
func (c *Converter) findNestedMessages(curPkg *ProtoPackage, msgDesc *descriptor.DescriptorProto) (map[*descriptor.DescriptorProto]string, error) {

	// Get a list of all nested messages, and how often they occur:
	nestedMessages := make(map[*descriptor.DescriptorProto]string)
	if err := c.recursiveFindNestedMessages(curPkg, msgDesc, msgDesc.GetName(), nestedMessages); err != nil {
		return nil, err
	}

	// Now filter them:
	result := make(map[*descriptor.DescriptorProto]string)
	for message, messageName := range nestedMessages {
		if !message.GetOptions().GetMapEntry() && !strings.HasPrefix(messageName, ".google.protobuf.") {
			// Check if this message is defined in a dependency
			messageName := strings.TrimLeft(messageName, ".")
			if isDependency, _ := c.isMessageDefinedInDependency(message.GetName()); isDependency {
				// If dependencies_as_external_refs=false, include dependencies in local definitions
				// If dependencies_as_external_refs=true, exclude them (they'll be in separate files)
				if !c.dependenciesAsExtRefs {
					result[message] = messageName
				}
			} else {
				// Always include non-dependency messages
				result[message] = messageName
			}
		}
	}

	return result, nil
}

func (c *Converter) recursiveFindNestedMessages(curPkg *ProtoPackage, msgDesc *descriptor.DescriptorProto, typeName string, nestedMessages map[*descriptor.DescriptorProto]string) error {
	if _, present := nestedMessages[msgDesc]; present {
		return nil
	}
	nestedMessages[msgDesc] = typeName

	for _, desc := range msgDesc.GetField() {
		descType := desc.GetType()
		if descType != descriptor.FieldDescriptorProto_TYPE_MESSAGE && descType != descriptor.FieldDescriptorProto_TYPE_GROUP {
			// no nested messages
			continue
		}

		typeName := desc.GetTypeName()
		recordType, _, ok := c.lookupType(curPkg, typeName)
		if !ok {
			return fmt.Errorf("no such message type named %s", typeName)
		}
		if err := c.recursiveFindNestedMessages(curPkg, recordType, typeName, nestedMessages); err != nil {
			return err
		}
	}

	return nil
}

func (c *Converter) recursiveConvertMessageType(curPkg *ProtoPackage, msgDesc *descriptor.DescriptorProto, pkgName string, duplicatedMessages map[*descriptor.DescriptorProto]string, ignoreDuplicatedMessages bool) (*jsonschema.Schema, error) {

	// Prepare a new jsonschema:
	jsonSchemaType := new(jsonschema.Schema)

	// Set some per-message flags from config and options:
	messageFlags := c.Flags

	// Custom message options from protoc-gen-jsonschema:
	if opt := proto.GetExtension(msgDesc.GetOptions(), protoc_gen_jsonschema.E_MessageOptions); opt != nil {
		if messageOptions, ok := opt.(*protoc_gen_jsonschema.MessageOptions); ok {

			// AllFieldsRequired:
			if messageOptions.GetAllFieldsRequired() {
				messageFlags.AllFieldsRequired = true
			}

			// AllowNullValues:
			if messageOptions.GetAllowNullValues() {
				messageFlags.AllowNullValues = true
			}

			// DisallowAdditionalProperties:
			if messageOptions.GetDisallowAdditionalProperties() {
				messageFlags.DisallowAdditionalProperties = true
			}

			// ENUMs as constants:
			if messageOptions.GetEnumsAsConstants() {
				messageFlags.EnumsAsConstants = true
			}
		}
	}

	// Generate a description from src comments (if available)
	if src := c.sourceInfo.GetMessage(msgDesc); src != nil {
		jsonSchemaType.Title, jsonSchemaType.Description, _ = c.formatTitleAndDescription(strPtr(msgDesc.GetName()), src)
	}

	// Handle google's well-known types:
	if msgDesc.Name != nil && wellKnownTypes[*msgDesc.Name] && pkgName == ".google.protobuf" {
		switch *msgDesc.Name {
		case "DoubleValue", "FloatValue":
			jsonSchemaType.Type = gojsonschema.TYPE_NUMBER
		case "Int32Value", "UInt32Value":
			jsonSchemaType.Type = gojsonschema.TYPE_INTEGER
		case "Int64Value", "UInt64Value":
			// BigInt as ints
			if messageFlags.DisallowBigIntsAsStrings {
				jsonSchemaType.Type = gojsonschema.TYPE_INTEGER
			} else {

				// BigInt as strings
				jsonSchemaType.Type = gojsonschema.TYPE_STRING
			}

		case "BoolValue":
			jsonSchemaType.Type = gojsonschema.TYPE_BOOLEAN
		case "BytesValue", "StringValue":
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
		case "Value":
			jsonSchemaType.OneOf = []*jsonschema.Schema{
				{Type: gojsonschema.TYPE_ARRAY},
				{Type: gojsonschema.TYPE_BOOLEAN},
				{Type: gojsonschema.TYPE_NUMBER},
				{Type: gojsonschema.TYPE_OBJECT},
				{Type: gojsonschema.TYPE_STRING},
			}
			jsonSchemaType.AdditionalProperties = jsonschema.TrueSchema
		case "Duration":
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
		case "Struct":
			jsonSchemaType.Type = gojsonschema.TYPE_OBJECT
			jsonSchemaType.AdditionalProperties = jsonschema.TrueSchema
		case "ListValue":
			jsonSchemaType.Type = gojsonschema.TYPE_ARRAY
		}

		// If we're allowing nulls then prepare a OneOf:
		if messageFlags.AllowNullValues {
			jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Schema{Type: gojsonschema.TYPE_NULL}, &jsonschema.Schema{Type: jsonSchemaType.Type})
			// and clear the Type that was previously set.
			jsonSchemaType.Type = ""
			return jsonSchemaType, nil
		}

		// Otherwise just return this simple type:
		return jsonSchemaType, nil
	}

	// Set defaults:
	//jsonSchemaType.Properties = orderedmap.OrderedMap[string, *jsonschema.Schema]{}

	// Look up references:
	if nameWithPackage, ok := duplicatedMessages[msgDesc]; ok && !ignoreDuplicatedMessages {
		var typeName string
		if c.Flags.TypeNamesWithNoPackage {
			typeName = msgDesc.GetName()
		} else {
			typeName = nameWithPackage
		}

		return &jsonschema.Schema{
			Ref: fmt.Sprintf("%s%s", c.refPrefix, typeName),
		}, nil
	}

	// Optionally allow NULL values:
	if messageFlags.AllowNullValues {
		jsonSchemaType.OneOf = []*jsonschema.Schema{
			{Type: gojsonschema.TYPE_NULL},
			{Type: gojsonschema.TYPE_OBJECT},
		}
	} else {
		jsonSchemaType.Type = gojsonschema.TYPE_OBJECT
	}

	// disallowAdditionalProperties will prevent validation where extra fields are found (outside of the schema):
	if messageFlags.DisallowAdditionalProperties {
		jsonSchemaType.AdditionalProperties = jsonschema.FalseSchema //false
	} else {
		jsonSchemaType.AdditionalProperties = jsonschema.TrueSchema //true
	}

	c.logger.WithField("message_str", msgDesc.String()).Trace("Converting message")
	for _, fieldDesc := range msgDesc.GetField() {

		// Custom field options from protoc-gen-jsonschema:
		if opt := proto.GetExtension(fieldDesc.GetOptions(), protoc_gen_jsonschema.E_FieldOptions); opt != nil {
			if fieldOptions, ok := opt.(*protoc_gen_jsonschema.FieldOptions); ok {
				// "Ignored" fields are simply skipped:
				if fieldOptions.GetIgnore() {
					c.logger.WithField("field_name", fieldDesc.GetName()).WithField("message_name", msgDesc.GetName()).Debug("Skipping ignored field")
					continue
				}

				// "Required" fields are added to the list of required attributes in our schema:
				if fieldOptions.GetRequired() {
					c.logger.WithField("field_name", fieldDesc.GetName()).WithField("message_name", msgDesc.GetName()).Debug("Marking required field")
					if c.Flags.UseJSONFieldnamesOnly {
						jsonSchemaType.Required = append(jsonSchemaType.Required, fieldDesc.GetJsonName())
					} else {
						jsonSchemaType.Required = append(jsonSchemaType.Required, fieldDesc.GetName())
					}
				}
			}
		}

		// Convert the field into a JSONSchema type:
		recursedJSONSchemaType, err := c.convertField(curPkg, fieldDesc, msgDesc, duplicatedMessages, messageFlags)
		if err != nil {
			c.logger.WithError(err).WithField("field_name", fieldDesc.GetName()).WithField("message_name", msgDesc.GetName()).Error("Failed to convert field")
			return nil, err
		}
		c.logger.WithField("field_name", fieldDesc.GetName()).WithField("type", recursedJSONSchemaType.Type).Trace("Converted field")

		// If this field is part of a OneOf declaration then build that here:
		if c.Flags.EnforceOneOf && fieldDesc.OneofIndex != nil && !fieldDesc.GetProto3Optional() {
			for {
				if *fieldDesc.OneofIndex < int32(len(jsonSchemaType.AllOf)) {
					break
				}
				var notAnyOf = &jsonschema.Schema{Not: &jsonschema.Schema{AnyOf: []*jsonschema.Schema{}}}
				jsonSchemaType.AllOf = append(jsonSchemaType.AllOf, &jsonschema.Schema{OneOf: []*jsonschema.Schema{notAnyOf}})
			}
			if c.Flags.UseJSONFieldnamesOnly {
				jsonSchemaType.AllOf[*fieldDesc.OneofIndex].OneOf = append(jsonSchemaType.AllOf[*fieldDesc.OneofIndex].OneOf, &jsonschema.Schema{Required: []string{fieldDesc.GetJsonName()}})
				jsonSchemaType.AllOf[*fieldDesc.OneofIndex].OneOf[0].Not.AnyOf = append(jsonSchemaType.AllOf[*fieldDesc.OneofIndex].OneOf[0].Not.AnyOf, &jsonschema.Schema{Required: []string{fieldDesc.GetJsonName()}})
			} else {
				jsonSchemaType.AllOf[*fieldDesc.OneofIndex].OneOf = append(jsonSchemaType.AllOf[*fieldDesc.OneofIndex].OneOf, &jsonschema.Schema{Required: []string{fieldDesc.GetName()}})
				jsonSchemaType.AllOf[*fieldDesc.OneofIndex].OneOf[0].Not.AnyOf = append(jsonSchemaType.AllOf[*fieldDesc.OneofIndex].OneOf[0].Not.AnyOf, &jsonschema.Schema{Required: []string{fieldDesc.GetName()}})
			}
		}

		//Output:
		// Figure out which field names we want to use:
		if jsonSchemaType.Properties == nil {
			jsonSchemaType.Properties = orderedmap.New[string, *jsonschema.Schema]()
		}
		//c.logger.WithField("prop", jsonSchemaType.Properties).Error()
		switch {

		case c.Flags.UseJSONFieldnamesOnly:
			jsonSchemaType.Properties.Set(fieldDesc.GetJsonName(), recursedJSONSchemaType)
		case c.Flags.UseProtoAndJSONFieldNames:
			jsonSchemaType.Properties.Set(fieldDesc.GetName(), recursedJSONSchemaType)
			jsonSchemaType.Properties.Set(fieldDesc.GetJsonName(), recursedJSONSchemaType)
		default:
			jsonSchemaType.Properties.Set(fieldDesc.GetName(), recursedJSONSchemaType)
		}

		// Enforce all_fields_required:
		if messageFlags.AllFieldsRequired {
			if fieldDesc.OneofIndex == nil && !fieldDesc.GetProto3Optional() {
				if c.Flags.UseJSONFieldnamesOnly {
					jsonSchemaType.Required = append(jsonSchemaType.Required, fieldDesc.GetJsonName())
				} else {
					jsonSchemaType.Required = append(jsonSchemaType.Required, fieldDesc.GetName())
				}
			}
		}

		// Look for required fields by the proto2 "required" flag:
		if fieldDesc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REQUIRED && fieldDesc.OneofIndex == nil {
			if c.Flags.UseJSONFieldnamesOnly {
				jsonSchemaType.Required = append(jsonSchemaType.Required, fieldDesc.GetJsonName())
			} else {
				jsonSchemaType.Required = append(jsonSchemaType.Required, fieldDesc.GetName())
			}
		}
	}

	// Remove empty properties to keep the final output as clean as possible:
	if jsonSchemaType.Properties.Len() == 0 {
		jsonSchemaType.Properties = nil
	}

	// Dedupe required fields:
	jsonSchemaType.Required = dedupe(jsonSchemaType.Required)

	return jsonSchemaType, nil
}

func dedupe(inputStrings []string) []string {
	appended := make(map[string]bool)
	outputStrings := []string{}

	for _, inputString := range inputStrings {
		if !appended[inputString] {
			outputStrings = append(outputStrings, inputString)
			appended[inputString] = true
		}
	}
	return outputStrings
}
