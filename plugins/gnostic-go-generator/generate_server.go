package main

func (renderer *ServiceRenderer) GenerateServer() ([]byte, error) {
	f := NewLineWriter()
	f.WriteLine("// GENERATED FILE: DO NOT EDIT!\n")
	f.WriteLine(``)
	f.WriteLine("package " + renderer.Model.Package)
	f.WriteLine(``)
	imports := []string{
		"encoding/json",
		"errors",
		"net/http",
		"strconv",
		"github.com/gorilla/mux",
	}
	f.WriteLine(``)
	f.WriteLine(`import (`)
	for _, imp := range imports {
		f.WriteLine(`"` + imp + `"`)
	}
	f.WriteLine(`)`)

	f.WriteLine(`func intValue(s string) (v int64) {`)
	f.WriteLine(`	v, _ = strconv.ParseInt(s, 10, 64)`)
	f.WriteLine(`	return v`)
	f.WriteLine(`}`)
	f.WriteLine(``)
	f.WriteLine(`// This package-global variable holds the user-written Provider for API services.`)
	f.WriteLine(`// See the Provider interface for details.`)
	f.WriteLine(`var provider Provider`)
	f.WriteLine(``)
	f.WriteLine(`// These handlers serve API methods.`)
	f.WriteLine(``)

	for _, method := range renderer.Model.Methods {
		f.WriteLine(`// Handler`)
		f.WriteLine(commentForText(method.Description))
		f.WriteLine(`func ` + method.HandlerName + `(w http.ResponseWriter, r *http.Request) {`)
		f.WriteLine(`  var err error`)
		if hasParameters(method) {
			f.WriteLine(`// instantiate the parameters structure`)
			f.WriteLine(`parameters := &` + method.ParametersTypeName + `{}`)
			if method.Method == "POST" {
				f.WriteLine(`// deserialize request from post data`)
				f.WriteLine(`decoder := json.NewDecoder(r.Body)`)
				f.WriteLine(`err = decoder.Decode(&parameters.` + bodyParameterFieldName(method) + `)`)
				f.WriteLine(`if err != nil {`)
				f.WriteLine(`	w.WriteHeader(http.StatusBadRequest)`)
				f.WriteLine(`	w.Write([]byte(err.Error() + "\n"))`)
				f.WriteLine(`	return`)
				f.WriteLine(`}`)
			}
			f.WriteLine(`// get request fields in path and query parameters`)
			if hasPathParameters(method) {
				f.WriteLine(`vars := mux.Vars(r)`)
			}
			if hasFormParameters(method) {
				f.WriteLine(`r.ParseForm()`)
			}
			for _, field := range method.ParametersType.Fields {
				if field.Position == "path" {
					f.WriteLine(`if value, ok := vars["` + field.JSONName + `"]; ok {`)
					f.WriteLine(`	parameters.` + field.FieldName + ` = intValue(value)`)
					f.WriteLine(`}`)
				} else if field.Position == "formdata" {
					f.WriteLine(`if len(r.Form["` + field.JSONName + `"]) > 0 {`)
					f.WriteLine(`	parameters.` + field.FieldName + ` = intValue(r.Form["` + field.JSONName + `"][0])`)
					f.WriteLine(`}`)
				}
			}
		}
		if hasResponses(method) {
			f.WriteLine(`// instantiate the responses structure`)
			f.WriteLine(`responses := &` + method.ResponsesTypeName + `{}`)
		}
		f.WriteLine(`// call the service provider`)
		if hasParameters(method) {
			if hasResponses(method) {
				f.WriteLine(`err = provider.` + method.ProcessorName + `(parameters, responses)`)
			} else {
				f.WriteLine(`err = provider.` + method.ProcessorName + `(parameters)`)
			}
		} else {
			if hasResponses(method) {
				f.WriteLine(`err = provider.` + method.ProcessorName + `(responses)`)
			} else {
				f.WriteLine(`err = provider.` + method.ProcessorName + `()`)
			}
		}
		f.WriteLine(`if err == nil {`)
		if hasResponses(method) {
			if hasFieldNamedOK(method.ResponsesType) {
				f.WriteLine(`if responses.OK != nil {`)
				f.WriteLine(`  // write the normal response`)
				f.WriteLine(`  encoder := json.NewEncoder(w)`)
				f.WriteLine(`  encoder.Encode(responses.OK)`)
				f.WriteLine(`  return`)
				f.WriteLine(`}`)
			}
			if hasFieldNamedDefault(method.ResponsesType) {
				f.WriteLine(`if responses.Default != nil {`)
				f.WriteLine(`  // write the error response`)
				f.WriteLine(`  w.WriteHeader(int(responses.Default.Code))`)
				f.WriteLine(`  encoder := json.NewEncoder(w)`)
				f.WriteLine(`  encoder.Encode(responses.Default)`)
				f.WriteLine(`  return`)
				f.WriteLine(`}`)
			}
			f.WriteLine(``)
			f.WriteLine(``)
			f.WriteLine(``)
			f.WriteLine(``)
		}
		f.WriteLine(`} else {`)
		f.WriteLine(`  w.WriteHeader(http.StatusInternalServerError)`)
		f.WriteLine(`  w.Write([]byte(err.Error() + "\n"))`)
		f.WriteLine(`  return`)
		f.WriteLine(`}`)
		f.WriteLine(`}`)
		f.WriteLine(``)
	}
	f.WriteLine(`// Initialize the API service.`)
	f.WriteLine(`func Initialize(p Provider) {`)
	f.WriteLine(`  provider = p`)
	f.WriteLine(`  var router = mux.NewRouter()`)
	for _, method := range renderer.Model.Methods {
		f.WriteLine(`router.HandleFunc("` + method.Path + `", ` + method.HandlerName + `).Methods("` + method.Method + `")`)
	}
	f.WriteLine(`  http.Handle("/", router)`)
	f.WriteLine(`}`)
	f.WriteLine(``)
	f.WriteLine(`// Provide the API service over HTTP.`)
	f.WriteLine(`func ServeHTTP(address string) error {`)
	f.WriteLine(`  if provider == nil {`)
	f.WriteLine(`    return errors.New("Use ` + renderer.Model.Package + `.Initialize() to set a service provider.")`)
	f.WriteLine(`  }`)
	f.WriteLine(`  return http.ListenAndServe(address, nil)`)
	f.WriteLine(`}`)

	return f.Bytes(), nil
}
