
import Stencil

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

class ServiceRenderer {
  var templateEnvironment : Environment

  var name : String = ""
  var package: String = ""
  var types : [ServiceType] = []
  var methods : [ServiceMethod] = []


  init(document : Openapi_V2_Document) {
    let ext = Extension()
    templateEnvironment = Environment(loader: InternalLoader(),
                                      extensions:[ext])
  }


  func generate(filenames : [String], response : inout Openapi_Plugin_V1_Response) throws {


    let context = ["123": 123]

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
