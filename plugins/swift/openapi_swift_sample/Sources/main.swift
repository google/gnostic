// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import Foundation

func printDocument(document:Openapi_V2_Document,
                   name:String,
                   version:String) -> String {
  var code = CodePrinter()
  code.print("READING \(name) (\(version))\n")
  code.print("Swagger: \(document.swagger)\n")
  code.print("Host: \(document.host)\n")
  code.print("BasePath: \(document.basePath)\n")
  if document.hasInfo {
    code.print("Info:\n")
    code.indent()
    if document.info.title != "" {
      code.print("Title: \(document.info.title)\n")
    }
    if document.info.description_p != "" {
      code.print("Description: \(document.info.description_p)\n")
    }
    if document.info.version != "" {
      code.print("Version: \(document.info.version)\n")
    }
    code.outdent()
  }
  code.print("Paths:\n")
  code.indent()
  for pair in document.paths.path {
    let v = pair.value
    if v.hasGet {
      code.print("GET \(pair.name)\n")
    }
    if v.hasPost {
      code.print("POST \(pair.name)\n")
    }
  }
  code.outdent()
  return code.content
}

func main() throws {
  var response = Openapic_V1_PluginResponse()
  let rawRequest = try Stdin.readall()
  let request = try Openapic_V1_PluginRequest(protobuf: rawRequest)
  for wrapper in request.wrapper {
    let document = try Openapi_V2_Document(protobuf:wrapper.value)
    response.text.append(printDocument(document:document, name:wrapper.name, version:wrapper.version))
  }
  let serializedResponse = try response.serializeProtobuf()
  Stdout.write(bytes: serializedResponse)
}

try main()
