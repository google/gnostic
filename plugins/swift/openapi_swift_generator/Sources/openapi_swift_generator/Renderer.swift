
import Stencil


extension String {
  func capitalizingFirstLetter() -> String {
    let first = String(characters.prefix(1)).capitalized
    let other = String(characters.dropFirst())
    return first + other
  }

  mutating func capitalizeFirstLetter() {
    self = self.capitalizingFirstLetter()
  }
}


class ServiceType {
  var name : String = ""
  var fields : [ServiceTypeField] = []
}

class ServiceTypeField {
  var name : String = ""
  var typeName : String = ""
  var jsonName : String = ""
  var position: String = "" // "body", "header", "formdata", "query", or "path"
}

class ServiceMethod {
  var name               : String = ""
  var path               : String = ""
  var method             : String = ""
  var description        : String = ""
  var handlerName        : String = ""
  var processorName      : String = ""
  var clientName         : String = ""
  var resultTypeName     : String = ""
  var parametersTypeName : String = ""
  var responsesTypeName  : String = ""
  var parametersType     : ServiceType?
  var responsesType      : ServiceType?
}

func propertyNameForResponseCode(_ code:String) -> String {
  switch code {
  case "200":
    return "ok"
  case "default":
    return "error"
  default:
    return code
  }
}

func typeForRef(_ ref : String) -> String {
  let parts = ref.components(separatedBy:"/")
  return parts.last!.capitalizingFirstLetter()
}

func typeForSchema(_ schema : Openapi_V2_Schema) -> String {
  let ref = schema.ref
  if ref != "" {
    return typeForRef(ref)
  }
  if schema.hasType {
    let types = schema.type.value
    let format = schema.format
    if types.count == 1 && types[0] == "string" {
      return "String"
    }
    if types.count == 1 && types[0] == "integer" && format == "int32" {
      return "Int32"
    }
    if types.count == 1 && types[0] == "array" && schema.hasItems {
      // we have an array.., but of what?
      let items = schema.items.schema
      if items.count == 1 && items[0].ref != "" {
        return "[]" + typeForRef(items[0].ref)
      }
    }
  }
  // this function is incomplete... so return a string representing anything that we don't handle
  return "\(schema)"
}

func typeForName(_ name : String, _ format : String) -> String {
  switch name {
  case "integer":
    if format == "int32" {
      return "Int32"
    } else if format == "int64" {
      return "Int64"
    } else {
      return "Int"
    }
  default:
    return name.capitalizingFirstLetter()
  }
}

class ServiceRenderer {
  private var templateEnvironment : Environment

  private var name : String = ""
  private var package: String = ""
  private var types : [ServiceType] = []
  private var methods : [ServiceMethod] = []

  public init(document : Openapi_V2_Document) {
    let ext = Extension()
    templateEnvironment = Environment(loader:InternalLoader(), extensions:[ext])
    loadService(document:document)
  }

  private func loadServiceTypeFromParameters(_ name:String,
                                             _ parameters:[Openapi_V2_ParametersItem])
    -> ServiceType {
      let t = ServiceType()
      t.name = name.capitalizingFirstLetter() + "Parameters"
      for parametersItem in parameters {
        let f = ServiceTypeField()
        f.typeName = "\(parametersItem)"

        switch parametersItem.oneof {
        case .parameter(let parameter):
          switch parameter.oneof {
          case .bodyParameter(let bodyParameter):
            f.name = bodyParameter.name
            if bodyParameter.hasSchema {
              f.typeName = typeForSchema(bodyParameter.schema)
              f.position = "body"
            }
          case .nonBodyParameter(let nonBodyParameter):
            switch (nonBodyParameter.oneof) {
            case .headerParameterSubSchema(let headerParameter):
              f.name = headerParameter.name
              f.position = "header"
            case .formDataParameterSubSchema(let formDataParameter):
              f.name = formDataParameter.name
              f.position = "formdata"
            case .queryParameterSubSchema(let queryParameter):
              f.name = queryParameter.name
              f.position = "query"
            case .pathParameterSubSchema(let pathParameter):
              f.name = pathParameter.name
              f.position = "path"
              f.typeName = typeForName(pathParameter.type, pathParameter.format)
            default:
              Log("?")
            }
          default:
            Log("?")
          }
        case .jsonReference: // (let reference):
          Log("?")
        default:
          Log("?")
        }
        t.fields.append(f)
      }
      if t.fields.count > 0 {
        self.types.append(t)
      }
      return t
  }

  private func loadServiceTypeFromResponses(_ m:ServiceMethod,
                                            _ name:String,
                                            _ responses:Openapi_V2_Responses)
    -> ServiceType {
      let t = ServiceType()
      t.name = name.capitalizingFirstLetter() + "Responses"
      for responseCode in responses.responseCode {
        let f = ServiceTypeField()
        f.name = propertyNameForResponseCode(responseCode.name)
        f.jsonName = ""
        switch responseCode.value.oneof {
        case .response(let response):
          let schema = response.schema
          switch schema.oneof {
          case .schema(let schema):
            f.typeName = typeForSchema(schema) + "?"
            t.fields.append(f)
            if f.name == "ok" {
              m.resultTypeName = typeForSchema(schema)
            }
          default:
            Log("unknown")
          }
        default:
          Log("unknown")
        }
      }
      if t.fields.count > 0 {
        self.types.append(t)
      }
      return t
  }

  private func loadOperation(_ operation : Openapi_V2_Operation,
                             method : String,
                             path : String) {
    let m = ServiceMethod()
    m.name = operation.operationId
    m.path = path
    m.method = method
    m.description = operation.description_p
    m.handlerName = "handle" + m.name
    m.processorName = "" + m.name
    m.clientName = m.name
    m.parametersType = loadServiceTypeFromParameters(m.name, operation.parameters)
    if m.parametersType != nil {
      m.parametersTypeName = m.parametersType!.name
    }
    m.responsesType = loadServiceTypeFromResponses(m, m.name, operation.responses)
    if m.responsesType != nil {
      m.responsesTypeName = m.responsesType!.name
    }
    self.methods.append(m)
  }

  private func loadService(document : Openapi_V2_Document) {
    // collect service type descriptions
    for pair in document.definitions.additionalProperties {
      let t = ServiceType()
      let schema = pair.value
      for pair2 in schema.properties.additionalProperties {
        let f = ServiceTypeField()
        f.name = pair2.name
        f.typeName = typeForSchema(pair2.value)
        f.jsonName = pair2.name
        t.fields.append(f)
      }
      t.name = pair.name.capitalizingFirstLetter()
      self.types.append(t)
    }
    // collect service method descriptions
    for pair in document.paths.path {
      let v = pair.value
      if v.hasGet {
        loadOperation(v.get, method:"GET", path:pair.name)
      }
      if v.hasPost {
        loadOperation(v.post, method:"POST", path:pair.name)
      }
      if v.hasPut {
        loadOperation(v.put, method:"PUT", path:pair.name)
      }
      if v.hasDelete {
        loadOperation(v.delete, method:"DELETE", path:pair.name)
      }
    }
  }

  public func generate(filenames : [String], response : inout Openapi_Plugin_V1_Response) throws {
    let context = ["renderer": self]

    for filename in filenames {
      let clientcode = try templateEnvironment.renderTemplate(name:filename,
                                                              context:context)
      if let data = stripMarkers(clientcode).data(using:.utf8) {
        var clientfile = Openapi_Plugin_V1_File()
        clientfile.name = filename
        clientfile.data = data
        response.files.append(clientfile)
      }
    }
  }
}
