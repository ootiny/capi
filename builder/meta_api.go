package builder

import (
	"fmt"
	"strings"
)

type APIDefinitionAttributeMeta struct {
	Name        string `json:"name" required:"true"`
	Type        string `json:"type" required:"true"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type APIDefinitionMeta struct {
	Description string                        `json:"description"`
	Attributes  []*APIDefinitionAttributeMeta `json:"attributes"`
}

type APIActionParameterMeta struct {
	Name        string `json:"name" required:"true"`
	Type        string `json:"type" required:"true"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type APIActionReturnMeta struct {
	Type        string `json:"type" required:"true"`
	Description string `json:"description"`
}

type APIActionMeta struct {
	Description string                    `json:"description"`
	Method      string                    `json:"method" required:"true"`
	Parameters  []*APIActionParameterMeta `json:"parameters"`
	Return      *APIActionReturnMeta      `json:"return"`
}

type APIMeta struct {
	Version      string                        `json:"version" required:"true"`
	Namespace    string                        `json:"namespace" required:"true"`
	Description  string                        `json:"description"`
	Definitions  map[string]*APIDefinitionMeta `json:"definitions" required:"true"`
	Actions      map[string]*APIActionMeta     `json:"actions"`
	__filepath__ string
}

func (c *APIMeta) GetFilePath() string {
	return c.__filepath__
}

func LoadAPIMeta(filePath string) (*APIMeta, error) {
	var meta APIMeta

	if err := UnmarshalConfig(filePath, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse meta file: %w", err)
	}

	meta.__filepath__ = filePath

	return &meta, nil
}

type APIMetaNode struct {
	name      string
	namespace string
	meta      *APIMeta
	children  map[string]*APIMetaNode
}

func MakeAPIConfigTree(configlist []*APIMeta) *APIMetaNode {
	nsMap := map[string]*APIMeta{}
	for _, meta := range configlist {
		nsMap[meta.Namespace] = meta
	}

	buildMap := map[string]*APIMetaNode{}

	for _, meta := range nsMap {
		nsArr := strings.Split(meta.Namespace, ".")

		if len(nsArr) > 1 && (nsArr[0] == "API" || nsArr[0] == "DB") {
			for i := range nsArr {
				partNS := strings.Join(nsArr[:i+1], ".")
				if _, ok := buildMap[partNS]; !ok {
					buildMap[partNS] = &APIMetaNode{
						name:      nsArr[i],
						namespace: partNS,
						meta:      nil,
						children:  map[string]*APIMetaNode{},
					}
				}
			}

			buildMap[meta.Namespace].meta = meta
		}
	}

	// 建立父子关系
	for namespace, node := range buildMap {
		nsArr := strings.Split(namespace, ".")
		if len(nsArr) > 1 {
			// 找到父节点的namespace
			parentNS := strings.Join(nsArr[:len(nsArr)-1], ".")
			if parentNode, ok := buildMap[parentNS]; ok {
				// 将当前节点添加到父节点的children中
				parentNode.children[node.name] = node
			}
		}
	}

	return &APIMetaNode{
		name:      "",
		namespace: "",
		meta:      nil,
		children: map[string]*APIMetaNode{
			"API": buildMap["API"],
			"DB":  buildMap["DB"],
		},
	}
}
