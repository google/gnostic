// Copyright 2020 Google LLC. All Rights Reserved.
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

import 'dart:io';
import '../lib/generated/OpenAPIv2.pb.dart' as openapiv2;
import '../lib/generated/OpenAPIv3.pb.dart' as openapiv3;
import 'package:protobuf/protobuf.dart';

//
// This is a simple program that accepts a list of protobuf-encoded OpenAPI descriptions
// as arguments and generates and prints high-level summaries of their contents.
//
void main(List<String> args) {
  args.forEach((filename) {
    File file = new File(filename);
    if (!(handleAsOpenAPIv2(file) || handleAsOpenAPIv3(file))) {
      print(
          "Unknown format. Inputs must be binary protobuf-encoded OpenAPI descriptions.");
    }
  });
}

bool handleAsOpenAPIv2(File file) {
  try {
    openapiv2.Document doc =
        new openapiv2.Document.fromBuffer(file.readAsBytesSync());
    final summary = summarizeOpenAPIv2Document(doc);
    summary.dump();
    return true;
  } on InvalidProtocolBufferException {
    return false;
  } catch (error) {
    print(error);
    return false;
  }
}

bool handleAsOpenAPIv3(File file) {
  try {
    openapiv3.Document doc =
        new openapiv3.Document.fromBuffer(file.readAsBytesSync());
    final summary = summarizeOpenAPIv3Document(doc);
    summary.dump();
    return true;
  } on InvalidProtocolBufferException {
    return false;
  } catch (error) {
    print(error);
    return false;
  }
}

class Summary {
  String title = "";
  String description = "";
  String version = "";
  int schemaCount = 0;
  int pathCount = 0;
  int getCount = 0;
  int postCount = 0;
  int putCount = 0;
  int deleteCount = 0;
  List<String> tags = List();
  Summary() {}

  dump() {
    print("");
    print("title   ${title}");
    print("version ${version}");
    print("tags    ${tags}");
    print("paths   ${pathCount} " +
        "(get:${getCount} post:${postCount} put:${putCount} delete:${deleteCount})");
    print("schemas ${schemaCount}");
  }
}

Summary summarizeOpenAPIv2Document(openapiv2.Document doc) {
  final s = Summary();
  if (doc.hasInfo()) {
    if (doc.info.hasTitle()) {
      s.title = doc.info.title;
    }
    if (doc.info.hasDescription()) {
      s.description = doc.info.description;
    }
    if (doc.info.hasVersion()) {
      s.version = doc.info.version;
    }
  }
  doc.paths.path.forEach((pair) {
    s.pathCount++;
    final path = pair.value;
    if (path.hasGet()) s.getCount++;
    if (path.hasPost()) s.postCount++;
    if (path.hasPut()) s.putCount++;
    if (path.hasDelete()) s.deleteCount++;
  });
  doc.definitions.additionalProperties.forEach((pair) {
    s.schemaCount++;
  });
  doc.tags.forEach((tag) {
    if (!s.tags.contains(tag.name)) {
      s.tags.add(tag.name);
    }
  });
  s.tags.sort();
  return s;
}

Summary summarizeOpenAPIv3Document(openapiv3.Document doc) {
  final s = Summary();
  if (doc.hasInfo()) {
    if (doc.info.hasTitle()) {
      s.title = doc.info.title;
    }
    if (doc.info.hasDescription()) {
      s.description = doc.info.description;
    }
    if (doc.info.hasVersion()) {
      s.version = doc.info.version;
    }
  }
  doc.paths.path.forEach((pair) {
    s.pathCount++;
    final path = pair.value;
    if (path.hasGet()) s.getCount++;
    if (path.hasPost()) s.postCount++;
    if (path.hasPut()) s.putCount++;
    if (path.hasDelete()) s.deleteCount++;
  });
  doc.components.schemas.additionalProperties.forEach((pair) {
    s.schemaCount++;
  });
  doc.tags.forEach((tag) {
    if (!s.tags.contains(tag.name)) {
      s.tags.add(tag.name);
    }
  });
  s.tags.sort();
  return s;
}
