package main

import (
	"text/template"
)

// template support functions

func hasOKField(s *ServiceType) bool {
	return s.hasFieldNamed("OK")
}

func parameterList(m *ServiceMethod) string {
	result := ""
	for i, field := range m.ParametersType.Fields {
		if i > 0 {
			result += ", "
		}
		result += field.JSONName + " " + field.Type
	}
	return result
}

func bodyParameterName(m *ServiceMethod) string {
	for _, field := range m.ParametersType.Fields {
		if field.Position == "body" {
			return field.JSONName
		}
	}
	return ""
}

func bodyParameterFieldName(m *ServiceMethod) string {
	for _, field := range m.ParametersType.Fields {
		if field.Position == "body" {
			return field.Name
		}
	}
	return ""
}

func helpers() template.FuncMap {
	return template.FuncMap{
		"HASOK":                  hasOKField,
		"parameterList":          parameterList,
		"bodyParameterName":      bodyParameterName,
		"bodyParameterFieldName": bodyParameterFieldName,
	}
}
