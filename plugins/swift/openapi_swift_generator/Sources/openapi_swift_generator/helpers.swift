//
//  helpers.swift
//  
//
//  Created by Tim Burks on 2/2/17.
//
//

import Foundation

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

