import Foundation

let TEMPLATES = "Templates"

var s = ""
s += "// GENERATED: DO NOT EDIT\n"
s += "//\n"
s += "// This file contains base64 encodings of templates used for Swift OpenAPI code generation.\n"
s += "//\n"
s += "func loadTemplates() -> [String:String] {\n"
s += "  return [\n"

let filenames = try FileManager.default.contentsOfDirectory(atPath:TEMPLATES)
for filename in filenames {
  if filename.hasSuffix(".tmpl") {
    let fileURL = URL(fileURLWithPath:TEMPLATES + "/" + filename)
    let filedata = try Data(contentsOf:fileURL)
    let encoding = filedata.base64EncodedString()
    var templatename = filename
    if let extRange = templatename.range(of: ".tmpl") {
      templatename.replaceSubrange(extRange, with: "")
    }
    s += "    \"" + templatename + "\": \"" + encoding + "\",\n"
  }
}

s += "  ]\n"
s += "}\n"
print(s)
