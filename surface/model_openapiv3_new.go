package surface_v1

import (
	openapiv3 "github.com/googleapis/gnostic/OpenAPIv3"
	"log"
	nethttp "net/http"
	"strconv"
	"strings"
)

type FieldInfo struct {
	fieldKind   FieldKind
	fieldType   string
	fieldFormat string
	// For parameters
	fieldPosition Position
	fieldName     string
}

// NewModelFromOpenAPIv3 builds a model of an API service for use in code generation.
func NewModelFromOpenAPI3New(document *openapiv3.Document) (*Model, error) {
	return newOpenAPI3Builder().buildModelNew(document)
}

func (b *OpenAPI3Builder) buildModelNew(document *openapiv3.Document) (*Model, error) {
	// Set model properties from passed-in document.
	b.model.Types = make([]*Type, 0)
	b.model.Methods = make([]*Method, 0)
	b.model.Name = document.Info.Title
	b.buildFromDocument(document)
	return b.model, nil
}

// build builds an API service description, preprocessing its types and methods for code generation.
func (b *OpenAPI3Builder) buildFromDocument(document *openapiv3.Document) {
	b.buildFromComponents(document.Components)
	b.buildFromPaths(document.Paths)
}

func (b *OpenAPI3Builder) buildFromComponents(components *openapiv3.Components) {
	if components == nil {
		return
	}

	if schemas := components.Schemas; schemas != nil {
		for _, namedSchema := range schemas.AdditionalProperties {
			_ = b.buildFromSchemaOrReference(namedSchema.Name, namedSchema.Value)
		}
	}

	if parameters := components.Parameters; parameters != nil {
		for _, namedParameter := range parameters.AdditionalProperties {
			// Parameters in OpenAPI have a name field. See: https://swagger.io/specification/#parameterObject
			// We use that later the buildFromParam method to name the type we build.
			b.buildFromParamOrRef(namedParameter.Value)
		}
	}

	if responses := components.Responses; responses != nil {
		for _, namedResponses := range responses.AdditionalProperties {
			b.buildFromResponseOrRef(namedResponses.Name, namedResponses.Value)
		}
	}

	if requestBodies := components.RequestBodies; requestBodies != nil {
		for _, namedRequestBody := range requestBodies.AdditionalProperties {
			b.buildFromRequestBodyOrRef(namedRequestBody.Name, namedRequestBody.Value)
		}
	}
}

func (b *OpenAPI3Builder) buildFromPaths(paths *openapiv3.Paths) {
	for _, path := range paths.Path {
		b.buildFromNamedPath(path.Name, path.Value)
	}
}

func (b *OpenAPI3Builder) buildFromNamedPath(name string, pathItem *openapiv3.PathItem) {
	for _, method := range []string{"GET", "PUT", "POST", "DELETE", "OPTIONS", "HEAD", "PATCH", "TRACE"} {
		var op *openapiv3.Operation
		switch method {
		case "GET":
			op = pathItem.Get
		case "PUT":
			op = pathItem.Put
		case "POST":
			op = pathItem.Post
		case "DELETE":
			op = pathItem.Delete
		case "OPTIONS":
			op = pathItem.Options
		case "HEAD":
			op = pathItem.Head
		case "PATCH":
			op = pathItem.Patch
		case "TRACE":
			op = pathItem.Trace
		}
		if op != nil {
			m := &Method{
				Operation:   op.OperationId,
				Path:        name,
				Method:      method,
				Name:        sanitizeOperationName(op.OperationId),
				Description: op.Description,
			}
			if m.Name == "" {
				m.Name = generateOperationName(method, name)
			}
			m.ParametersTypeName, m.ResponsesTypeName = b.buildFromNamedOperation(m.Name, op)
			b.model.addMethod(m)
		}
	}
}

func (b *OpenAPI3Builder) buildFromNamedOperation(name string, operation *openapiv3.Operation) (parametersTypeName string, responseTypeName string) {
	// At first, we build the operations input parameters first. This includes parameters (like PATH or QUERY parameters) and a request body
	operationParameters := b.makeType(name + "Parameters")
	operationParameters.Description = operationParameters.Name + " holds parameters to " + name
	for _, paramOrRef := range operation.Parameters {
		fieldInfo := b.buildFromParamOrRef(paramOrRef)
		// For parameters we get the name from the parameter itself. That is why pass in fieldInfo.fieldName
		b.makeFieldAndAppendToType(fieldInfo, operationParameters, fieldInfo.fieldName)
	}

	if operation.RequestBody != nil {
		fInfo := b.buildFromRequestBodyOrRef(name+"RequestBody", operation.RequestBody)
		b.makeFieldAndAppendToType(fInfo, operationParameters, "request_body")
	}

	if len(operationParameters.Fields) > 0 {
		b.model.addType(operationParameters)
		parametersTypeName = operationParameters.Name
	}

	// Secondly, we build the response values for the method.
	if responses := operation.Responses; responses != nil {
		operationResponses := b.makeType(name + "Responses")
		operationResponses.Description = operationResponses.Name + " holds responses of " + name
		for _, namedResponse := range responses.ResponseOrReference {
			fieldInfo := b.buildFromResponseOrRef(convertStatusCodes(namedResponse.Name), namedResponse.Value)
			b.makeFieldAndAppendToType(fieldInfo, operationResponses, namedResponse.Name)
		}
		if len(operationResponses.Fields) > 0 {
			b.model.addType(operationResponses)
			responseTypeName = operationResponses.Name
		}
	}
	return parametersTypeName, responseTypeName
}

func (b *OpenAPI3Builder) buildFromParamOrRef(paramOrRef *openapiv3.ParameterOrReference) (fInfo *FieldInfo) {
	fInfo = &FieldInfo{}
	if param := paramOrRef.GetParameter(); param != nil {
		fInfo = b.buildFromParam(param)
		return fInfo
	} else if ref := paramOrRef.GetReference(); ref != nil {
		fInfo.fieldKind, fInfo.fieldType = FieldKind_REFERENCE, typeForRef(ref.XRef)
		return fInfo
	}
	return nil
}

func (b *OpenAPI3Builder) buildFromParam(parameter *openapiv3.Parameter) (fInfo *FieldInfo) {
	if schemaOrRef := parameter.Schema; schemaOrRef != nil {
		fInfo = b.buildFromSchemaOrReference(parameter.Name, schemaOrRef)
		fInfo.fieldName = parameter.Name
		switch parameter.In {
		case "body":
			fInfo.fieldPosition = Position_BODY
		case "header":
			fInfo.fieldPosition = Position_HEADER
		case "formdata":
			fInfo.fieldPosition = Position_FORMDATA
		case "query":
			fInfo.fieldPosition = Position_QUERY
		case "path":
			fInfo.fieldPosition = Position_PATH
		}
		return fInfo
	}
	return nil
}

func (b *OpenAPI3Builder) buildFromRequestBodyOrRef(name string, reqBodyOrRef *openapiv3.RequestBodyOrReference) (fInfo *FieldInfo) {
	fInfo = &FieldInfo{}
	if requestBody := reqBodyOrRef.GetRequestBody(); requestBody != nil {
		fInfo = b.buildFromRequestBody(name, requestBody)
		return fInfo
	} else if ref := reqBodyOrRef.GetReference(); ref != nil {
		fInfo.fieldKind, fInfo.fieldType = FieldKind_REFERENCE, typeForRef(ref.XRef)
		return fInfo
	}
	return nil
}

func (b *OpenAPI3Builder) buildFromRequestBody(name string, reqBody *openapiv3.RequestBody) (fInfo *FieldInfo) {
	fInfo = &FieldInfo{}
	if reqBody.Content != nil {
		schemaType := b.makeType(name)
		for _, namedMediaType := range reqBody.Content.AdditionalProperties {
			fieldInfo := b.buildFromNamedMediaType(namedMediaType.Name, namedMediaType.Value)
			b.makeFieldAndAppendToType(fieldInfo, schemaType, namedMediaType.Name)
		}
		b.model.addType(schemaType)
		fInfo.fieldKind, fInfo.fieldType = FieldKind_REFERENCE, schemaType.Name
		return fInfo
	}
	return nil
}

func (b *OpenAPI3Builder) buildFromResponseOrRef(name string, responseOrRef *openapiv3.ResponseOrReference) (fInfo *FieldInfo) {
	fInfo = &FieldInfo{}
	if response := responseOrRef.GetResponse(); response != nil {
		fInfo = b.buildFromResponse(name, response)
		return fInfo
	} else if ref := responseOrRef.GetReference(); ref != nil {
		fInfo.fieldKind, fInfo.fieldType = FieldKind_REFERENCE, typeForRef(ref.XRef)
		return fInfo
	}
	return nil
}

func (b *OpenAPI3Builder) buildFromResponse(name string, response *openapiv3.Response) (fInfo *FieldInfo) {
	fInfo = &FieldInfo{}
	if response.Content != nil && response.Content.AdditionalProperties != nil {
		schemaType := b.makeType(name)
		for _, namedMediaType := range response.Content.AdditionalProperties {
			fieldInfo := b.buildFromNamedMediaType(namedMediaType.Name, namedMediaType.Value)
			b.makeFieldAndAppendToType(fieldInfo, schemaType, namedMediaType.Name)
		}
		b.model.addType(schemaType)
		fInfo.fieldKind, fInfo.fieldType = FieldKind_REFERENCE, schemaType.Name
		return fInfo
	}
	log.Printf("Response has no content: %v", response)
	return nil
}

func (b *OpenAPI3Builder) buildFromNamedMediaType(name string, mediaType *openapiv3.MediaType) (fInfo *FieldInfo) {
	if schemaOrRef := mediaType.Schema; schemaOrRef != nil {
		fInfo = b.buildFromSchemaOrReference(name, schemaOrRef)
	}
	return fInfo
}

func (b *OpenAPI3Builder) buildFromSchemaOrReference(name string, schemaOrReference *openapiv3.SchemaOrReference) (fInfo *FieldInfo) {
	fInfo = &FieldInfo{}
	if schema := schemaOrReference.GetSchema(); schema != nil {
		fInfo = b.buildFromSchema(name, schema)
		return fInfo
	} else if ref := schemaOrReference.GetReference(); ref != nil {
		fInfo.fieldKind, fInfo.fieldType = FieldKind_REFERENCE, typeForRef(ref.XRef)
		return fInfo
	}
	return nil
}

func (b *OpenAPI3Builder) buildFromSchema(name string, schema *openapiv3.Schema) (fInfo *FieldInfo) {
	fInfo = &FieldInfo{}
	// Data types according to: https://swagger.io/docs/specification/data-models/data-types/
	switch schema.Type {
	case "":
		fallthrough
	case "object":
		if schema.Properties != nil && schema.Properties.AdditionalProperties != nil {
			schemaType := b.makeType(name)
			for _, namedSchema := range schema.Properties.AdditionalProperties {
				fieldInfo := b.buildFromSchemaOrReference(namedSchema.Name, namedSchema.Value)
				b.makeFieldAndAppendToType(fieldInfo, schemaType, namedSchema.Name)
			}
			b.model.addType(schemaType)
			fInfo.fieldKind, fInfo.fieldType, fInfo.fieldFormat = FieldKind_REFERENCE, schemaType.Name, ""
			return fInfo
		}
	case "array":
		for _, schemaOrRef := range schema.Items.SchemaOrReference {
			arrayFieldInfo := b.buildFromSchemaOrReference("", schemaOrRef)
			fInfo.fieldKind, fInfo.fieldType, fInfo.fieldFormat = FieldKind_ARRAY, arrayFieldInfo.fieldType, arrayFieldInfo.fieldFormat
		}
		return fInfo
	default:
		// We go a scalar value
		fInfo.fieldKind, fInfo.fieldType, fInfo.fieldFormat = FieldKind_SCALAR, schema.Type, schema.Format
		return fInfo
	}
	log.Printf("Could not find field info for schema: %v", schema)
	return nil
}

func (b *OpenAPI3Builder) makeType(name string) *Type {
	t := &Type{
		Name:   name,
		Kind:   TypeKind_STRUCT,
		Fields: make([]*Field, 0),
	}
	return t
}

func (b *OpenAPI3Builder) makeFieldAndAppendToType(info *FieldInfo, schemaType *Type, fieldName string) {
	if info != nil {
		f := &Field{Name: fieldName}
		f.Type, f.Kind, f.Format, f.Position = info.fieldType, info.fieldKind, info.fieldFormat, info.fieldPosition
		schemaType.Fields = append(schemaType.Fields, f)
	}
}

func convertStatusCodes(c string) (statusText string) {
	code, err := strconv.Atoi(c)
	if err == nil {
		statusText = nethttp.StatusText(code)
		if statusText == "" {
			log.Println("Status code " + c + "is not known to net.http.StatusText. This might cause unpredictable behavior.")
			statusText = "unknownStatusCode"
		}
		statusText = strings.Replace(statusText, " ", "_", -1)
	}
	return statusText
}
