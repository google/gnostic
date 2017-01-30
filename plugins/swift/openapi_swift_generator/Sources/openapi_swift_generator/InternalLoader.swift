/*
 *
 * Copyright 2017, Google Inc.
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     * Redistributions of source code must retain the above copyright
 * notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above
 * copyright notice, this list of conditions and the following disclaimer
 * in the documentation and/or other materials provided with the
 * distribution.
 *     * Neither the name of Google Inc. nor the names of its
 * contributors may be used to endorse or promote products derived from
 * this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

import Stencil
import Foundation

// A class for loading Stencil templates from compiled-in representations

public class InternalLoader: Loader {
  private var templates: [String:String]

  public init() {
    self.templates = loadTemplates()
  }

  public func loadTemplate(name: String, environment: Environment) throws -> Template {
    if let encoding = templates[name],
      let data = Data(base64Encoded: encoding, options:[]),
      let template = String(data:data, encoding:.utf8) {
      return environment.templateClass.init(templateString: template,
                                            environment: environment,
                                            name: name)
    } else {
      throw TemplateDoesNotExist(templateNames: [name], loader: self)
    }
  }

  public func loadTemplate(names: [String], environment: Environment) throws -> Template {
    for name in names {
      if let encoding = templates[name],
        let data = Data(base64Encoded: encoding, options:[]),
        let template = String(data:data, encoding:.utf8) {
        return environment.templateClass.init(templateString: template,
                                              environment: environment,
                                              name: name)
      }
    }
    throw TemplateDoesNotExist(templateNames: names, loader: self)
  }
}
