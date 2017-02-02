// Copyright 2017 Google Inc. All Rights Reserved.
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
import Stencil

// Code templates use "//-" prefixes to comment-out template operators
// to keep them from interfering with Swift code formatting tools.
// Use this to remove them after templates have been expanded.
func stripMarkers(_ code:String) -> String {
  let inputLines = code.components(separatedBy:"\n")

  var outputLines : [String] = []
  for line in inputLines {
    if line.contains("//-") {
      let removed = line.replacingOccurrences(of:"//-", with:"")
      if (removed.trimmingCharacters(in:CharacterSet.whitespaces) != "") {
        outputLines.append(removed)
      }
    } else {
      outputLines.append(line)
    }
  }
  return outputLines.joined(separator:"\n")
}

func Log(_ message : String) {
  FileHandle.standardError.write((message + "\n").data(using:.utf8)!)
}

func main() throws {

  let filenames = ["client.swift", "service.swift", "types.swift"]

  let ext = Extension()
  let templateEnvironment = Environment(loader: InternalLoader(),
                                        extensions:[ext])

  var response = Openapi_Plugin_V1_Response()
  let rawRequest = try Stdin.readall()
  let request = try Openapi_Plugin_V1_Request(protobuf: rawRequest)
  let wrapper = request.wrapper
  let document = try Openapi_V2_Document(protobuf:wrapper.value)
  let name = wrapper.name
  let version = wrapper.version

  let context = ["123": 123]

  for filename in filenames {
    let clientcode = try templateEnvironment.renderTemplate(name:filename,
                                                            context: context)
    if let data = stripMarkers(clientcode).data(using:.utf8) {
      var clientfile = Openapi_Plugin_V1_File()
      clientfile.name = filename
      clientfile.data = data
      response.files.append(clientfile)
    }
  }

  let serializedResponse = try response.serializeProtobuf()
  Stdout.write(bytes: serializedResponse)
}

try main()
