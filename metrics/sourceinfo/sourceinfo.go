package sourceinfo

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"gopkg.in/yaml.v3"
)

func findNode(node *yaml.Node, keyIndex int, maxDepth int, keys []string) (*yaml.Node, error) {
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		if keyIndex+1 == maxDepth && val.Value == keys[maxDepth] {
			return val, nil
		}
		if key.Value == keys[keyIndex] {
			switch val.Kind {
			case yaml.SequenceNode:
				nextKeyIndex, err := strconv.Atoi(keys[keyIndex+1])
				if err != nil {
					return nil, err
				}
				return findNode(val.Content[nextKeyIndex], keyIndex+2, maxDepth, keys)
			default:
				return findNode(val, keyIndex+1, maxDepth, keys)
			}
		} else {
			continue
		}

	}
	return &yaml.Node{}, nil
}

// FindNode returns a line number for a field within a yaml file given the path and yaml file.
func FindNode(filename string, keys []string, token string) (*yaml.Node, error) {
	data, _ := ioutil.ReadFile(filename)

	var node yaml.Node
	err := yaml.Unmarshal(data, &node)
	if err != nil {
		fmt.Printf("%+v", err)
	}
	keys = append(keys, token)
	return findNode(node.Content[0], 0, len(keys)-1, keys)
}
