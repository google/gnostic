
import Stencil

func TemplateExtensions() -> Extension {
  let ext = Extension()

  ext.registerFilter("hasParameters") { (value: Any?, arguments: [Any?]) in
    let method : ServiceMethod = arguments[0] as! ServiceMethod
    return method.parametersType != nil
  }
  ext.registerFilter("hasResponses") { (value: Any?, arguments: [Any?]) in
    let method : ServiceMethod = arguments[0] as! ServiceMethod
    return method.responsesType != nil
  }
  ext.registerFilter("clientParametersDeclaration") { (value: Any?, arguments: [Any?]) in
    let method : ServiceMethod = arguments[0] as! ServiceMethod
    var result = ""
    if let parametersType = method.parametersType {
      for field in parametersType.fields {
        if result != "" {
          result += ", "
        }
        result += field.name + " : " + field.typeName
      }
    }
    return result
  }
  ext.registerFilter("clientReturnDeclaration") { (value: Any?, arguments: [Any?]) in
    let method : ServiceMethod = arguments[0] as! ServiceMethod
    var result = ""
    if let resultTypeName = method.resultTypeName {
      result = " -> " + resultTypeName
    }
    return result
  }
  ext.registerFilter("protocolParametersDeclaration") { (value: Any?, arguments: [Any?]) in
    let method : ServiceMethod = arguments[0] as! ServiceMethod
    var result = ""
    if let parametersTypeName = method.parametersTypeName {
      result = "parameters : " + parametersTypeName
    }
    return result
  }
  ext.registerFilter("protocolReturnDeclaration") { (value: Any?, arguments: [Any?]) in
    let method : ServiceMethod = arguments[0] as! ServiceMethod
    var result = ""
    if let responsesTypeName = method.responsesTypeName {
      result = "-> " + responsesTypeName
    }
    return result
  }
  ext.registerFilter("parametersTypeFields") { (value: Any?, arguments: [Any?]) in
    let method : ServiceMethod = value as! ServiceMethod
    if let parametersType = method.parametersType {
      return parametersType.fields
    } else {
      return []
    }
  }
  ext.registerFilter("kituraPath") { (value: Any?, arguments: [Any?]) in
    let method : ServiceMethod = value as! ServiceMethod
    var path = method.path
    if let parametersType = method.parametersType {
      for field in parametersType.fields {
        if field.position == "path" {
          let original = "{" + field.jsonName + "}"
          let replacement = ":" + field.jsonName
          path = path.replacingOccurrences(of:original, with:replacement)
        }
      }
    }
    return path
  }
  return ext
}
