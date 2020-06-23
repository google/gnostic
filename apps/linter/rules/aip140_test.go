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

package rules

import (
	"testing"
)

func TestSnakeCase(t *testing.T) {
	if pass, suggestion := snakeCase("helloWorld"); pass {
		t.Error("Given \"helloWorld\", snakeCase() returned true, expected false")
	} else if suggestion != "hello_world" {
		t.Errorf("Expected suggestion \"hello_world\", received %s instead", suggestion)
	}

	if pass, suggestion := snakeCase("hello_world"); !pass {
		t.Error("Given \"hello_world\", snakeCase() returned false, expected true")
	} else if suggestion != "hello_world" {
		t.Errorf("Expected suggestion \"hello_world\", received %s instead", suggestion)
	}

}

func TestAbbreviation(t *testing.T) {
	if pass, suggestion := abbreviation("configuration"); !pass {
		t.Error("Given \"configuration\", abbreviation() returned false, expected true")
	} else if suggestion != "config" {
		t.Errorf("Expected suggestion \"config\", received %s instead", suggestion)
	}

	if pass, suggestion := abbreviation("identifier"); !pass {
		t.Error("Given \"identifier\", abbreviation() returned false, expected true")
	} else if suggestion != "id" {
		t.Errorf("Expected suggestion \"id\", received %s instead", suggestion)
	}

	if pass, suggestion := abbreviation("information"); !pass {
		t.Error("Given \"information\", abbreviation() returned false, expected true")
	} else if suggestion != "info" {
		t.Errorf("Expected suggestion \"info\", received %s instead", suggestion)
	}

	if pass, suggestion := abbreviation("specification"); !pass {
		t.Error("Given \"specification\", abbreviation() returned false, expected true")
	} else if suggestion != "spec" {
		t.Errorf("Expected suggestion \"spec\", received %s instead", suggestion)
	}

	if pass, suggestion := abbreviation("statistics"); !pass {
		t.Error("Given \"statistics\", abbreviation() returned false, expected true")
	} else if suggestion != "stats" {
		t.Errorf("Expected suggestion \"stats\", received %s instead", suggestion)
	}

	if pass, suggestion := abbreviation("supercalifrag"); pass {
		t.Error("Given \"supercalifrag\", abbreviation() returned true, expected false")
	} else if suggestion != "supercalifrag" {
		t.Errorf("Expected suggestion \"superalifrag\", received %s instead", suggestion)
	}

}

func TestNumbers(t *testing.T) {
	if pass := numbers("90th_percentile"); !pass {
		t.Error("Given \"90th_percentile\", numbers() returned false, expected true")
	}

	if pass := numbers("hello_2nd_world"); !pass {
		t.Error("Given \"hello_2nd_world\", numbers() returned false, expected true")
	}
	if pass := numbers("second"); pass {
		t.Error("Given \"second\", numbers() returned true, expected false")
	}

}

func TestReservedWords(t *testing.T) {
	if pass := reservedWords("catch"); !pass {
		t.Error("Given \"catch\", numbers() returned false, expected true")
	}

	if pass := reservedWords("all_except"); !pass {
		t.Error("Given \"all_except\", reservedWords() returned false, expected true")
	}

	if pass := reservedWords("export"); !pass {
		t.Error("Given \"export\", reservedWords() returned false, expected true")
	}

	if pass := reservedWords("interface"); !pass {
		t.Error("Given \"interface\", reservedWords() returned false, expected true")
	}

	if pass := reservedWords("magic"); pass {
		t.Error("Given \"magic\", reservedWords() returned true, expected false")
	}

}

func TestPrepositions(t *testing.T) {
	if pass := prepositions("written_by"); !pass {
		t.Error("Given \"written_by\", prepositions() returned false, expected true")
	}

	if pass := prepositions("all_except"); !pass {
		t.Error("Given \"all_except\", prepositions() returned false, expected true")
	}

	if pass := prepositions("process_after"); !pass {
		t.Error("Given \"process_after\", prepositions() returned false, expected true")
	}

	if pass := prepositions("between_rocks_by_shore"); !pass {
		t.Error("Given \"between_rocks_by_shore\", prepositions() returned false, expected true")
	}

	if pass := prepositions("no_preps_here"); pass {
		t.Error("Given \"magic\", prepositions() returned true, expected false")
	}

}
