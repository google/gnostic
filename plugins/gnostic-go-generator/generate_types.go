package main

func (renderer *ServiceRenderer) GenerateTypes() ([]byte, error) {
	f := NewLineWriter()
	f.WriteLine(`// GENERATED FILE: DO NOT EDIT!`)
	f.WriteLine(``)
	f.WriteLine(`package ` + renderer.Model.Package)
	f.WriteLine(`// Types used by the API.`)
	for _, modelType := range renderer.Model.Types {
		f.WriteLine(`// ` + modelType.Description)
		if modelType.Kind == "struct" {
			f.WriteLine(`type ` + modelType.Name + ` struct {`)
			for _, field := range modelType.Fields {
				f.WriteLine(field.FieldName + ` ` + goType(field.Type) + jsonTag(field))
			}
			f.WriteLine(`}`)
		} else {
			f.WriteLine(`type ` + modelType.Name + ` ` + modelType.Kind)
		}
	}
	return f.Bytes(), nil
}

func jsonTag(field *ServiceTypeField) string {
	if field.JSONName != "" {
		return " `json:" + `"` + field.JSONName + `"` + "`"
	}
	return ""
}
