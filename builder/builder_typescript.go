package builder

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
)

func toTypeScriptType(location string, currentPackage string, name string) (string, string) {
	name = strings.TrimSpace(name)

	switch name {
	case "String":
		return "string", ""
	case "Float64":
		return "number", ""
	case "Int64":
		return "number", ""
	case "Bool":
		return "boolean", ""
	case "Bytes":
		return "string", ""
	default:
		// if name is List<innter>, then return []inner
		if strings.HasPrefix(name, "List<") && strings.HasSuffix(name, ">") {
			innerType := name[5 : len(name)-1]
			ret, pkg := toTypeScriptType(location, currentPackage, innerType)
			return fmt.Sprintf("%s[]", ret), pkg
		} else if strings.HasPrefix(name, "Map<") && strings.HasSuffix(name, ">") {
			innerType := name[4 : len(name)-1] // Remove "Map<" and ">"
			ret, pkg := toTypeScriptType(location, currentPackage, innerType)
			return fmt.Sprintf("{ [key: string]: %s }", ret), pkg
		} else if strings.HasPrefix(name, DBPrefix) || strings.HasPrefix(name, APIPrefix) {
			nameArr := strings.Split(name, "@")
			if len(nameArr) == 2 {
				pkgName := NamespaceToFolder(location, nameArr[0])

				if pkgName == currentPackage {
					return nameArr[1], ""
				} else {
					pkg := fmt.Sprintf("import * as %s from \"../%s\"", pkgName, pkgName)
					return pkgName + "." + nameArr[1], pkg
				}
			} else {
				return name, ""
			}
		} else {
			return name, ""
		}
	}
}

type TypescriptBuilder struct{}

func (p *TypescriptBuilder) buildClientBaseFiles(ctx *BuildContext) (map[string]string, error) {
	systemDir := filepath.Join(ctx.output.Dir, "system")
	if engineContent, err := assets.ReadFile("assets/typescript/utils.ts"); err != nil {
		return nil, fmt.Errorf("failed to read assets file: %v", err)
	} else {
		return map[string]string{
			filepath.Join(systemDir, "utils.ts"): string(engineContent),
		}, nil
	}
}

func (p *TypescriptBuilder) BuildServer(ctx *BuildContext) (map[string]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (p *TypescriptBuilder) BuildClient(ctx *BuildContext) (map[string]string, error) {
	ret, err := p.buildClientBaseFiles(ctx)
	if err != nil {
		return nil, err
	}

	metas := []*APIMeta{}
	metas = append(metas, ctx.apiMetas...)
	for _, dbMeta := range ctx.dbMetas {
		if apiMeta, err := dbMeta.ToAPIMeta(); err != nil {
			return nil, err
		} else {
			metas = append(metas, apiMeta)
		}
	}

	rootNode := MakeAPIConfigTree(metas)
	if rootNode == nil {
		// no api found
		return ret, nil
	}

	if apiNode := rootNode.children["API"]; apiNode != nil {
		if fileMap, err := p.buildClientWithMetaNode(ctx, apiNode); err != nil {
			return nil, err
		} else {
			maps.Copy(ret, fileMap)
		}
	}

	if dbNode := rootNode.children["DB"]; dbNode != nil {
		if fileMap, err := p.buildClientWithMetaNode(ctx, dbNode); err != nil {
			return nil, err
		} else {
			maps.Copy(ret, fileMap)
		}
	}

	return ret, nil
}

func (p *TypescriptBuilder) buildClientWithMetaNode(ctx *BuildContext, metaNode *APIMetaNode) (map[string]string, error) {
	ret := map[string]string{}

	if metaNode.namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	currentPackage := NamespaceToFolder(ctx.location, metaNode.namespace)

	imports := []string{}

	defines := []string{}

	actions := []string{}

	// needImportFetchJson := false

	// definitions
	if metaNode.meta != nil {
		for name, define := range metaNode.meta.Definitions {
			if len(define.Attributes) > 0 {
				attributes := []string{}
				fullDefineName := metaNode.meta.Namespace + "@" + name
				for _, attribute := range define.Attributes {
					attrType, pkg := toTypeScriptType(ctx.location, currentPackage, attribute.Type)
					if pkg != "" {
						imports = append(imports, pkg)
					}

					attributes = append(attributes, fmt.Sprintf(
						"  %s: %s;",
						attribute.Name,
						attrType,
					))
				}

				defines = append(defines, fmt.Sprintf(
					"// definition: %s",
					fullDefineName,
				))
				defines = append(defines, fmt.Sprintf(
					"export interface %s {\n%s\n}\n",
					name,
					strings.Join(attributes, "\n"),
				))

			}
		}
	}

	// actions
	if metaNode.meta != nil && len(metaNode.meta.Actions) > 0 {
		imports = append(imports, "import { fetchJson } from \"../system/utils\";")
		for name, action := range metaNode.meta.Actions {
			if len(action.Parameters) > 0 {
				attributes := []string{}
				dataAttrs := []string{}
				fullActionName := metaNode.meta.Namespace + ":" + name
				method := strings.ToUpper(action.Method)
				for _, attribute := range action.Parameters {
					attrType, pkg := toTypeScriptType(ctx.location, currentPackage, attribute.Type)
					if pkg != "" {
						imports = append(imports, pkg)
					}

					attributes = append(attributes, fmt.Sprintf(
						"%s: %s",
						attribute.Name,
						attrType,
					))
					if attribute.Required {
						dataAttrs = append(dataAttrs, attribute.Name)
					}
				}

				returnType, pkg := toTypeScriptType(ctx.location, currentPackage, action.Return.Type)
				if pkg != "" {
					imports = append(imports, pkg)
				}
				if returnType == "" {
					returnType = "void"
				}

				actionStr := fmt.Sprintf("\t// action: %s\n", fullActionName)
				actionStr += fmt.Sprintf("\tasync %s(%s): Promise<%s> {\n", name, strings.Join(attributes, ", "), returnType)
				actionStr += fmt.Sprintf("\t\treturn fetchJson(this.url, \"%s\", \"%s\", { %s })\n", fullActionName, method, strings.Join(dataAttrs, ", "))
				actionStr += "\t}\n"

				actions = append(actions, actionStr)
			}
		}
	}

	// children
	childrenDefineContent := ""
	childrenConstructorContent := ""
	for name, child := range metaNode.children {
		tagetPackage := NamespaceToFolder(ctx.location, child.namespace)
		if tagetPackage != currentPackage {
			if metaNode.namespace == "API" {
				imports = append(imports, fmt.Sprintf("import * as %s from \"./%s\";", tagetPackage, tagetPackage))
			} else {
				imports = append(imports, fmt.Sprintf("import * as %s from \"../%s\";", tagetPackage, tagetPackage))
			}
		}

		childrenDefineContent += fmt.Sprintf("\tpublic %s: %s.__Main__;\n", name, tagetPackage)
		childrenConstructorContent += fmt.Sprintf("\t\tthis.%s = new %s.__Main__(url);\n", name, tagetPackage)
	}

	importsContent := ""
	if len(imports) > 0 {
		imports = slices.Compact(imports)
		importsContent = strings.Join(imports, "\n") + "\n"
	}

	defineContent := ""
	if len(defines) > 0 {
		defineContent = strings.Join(defines, "\n")
	}

	actionContent := ""
	if len(actions) > 0 || childrenDefineContent != "" {
		if metaNode.namespace == "API" {
			actionContent = "export class Client {\n"
		} else {
			actionContent = "export class __Main__ {\n"
		}
		if metaNode.meta != nil {
			actionContent += "\tprivate url: string;\n"
		}
		actionContent += childrenDefineContent
		actionContent += "\tconstructor(url: string) {\n"
		if metaNode.meta != nil {
			actionContent += "\t\tthis.url = url;\n"
		}
		actionContent += childrenConstructorContent
		actionContent += "\t}\n"
		actionContent += "\n"
		actionContent += strings.Join(actions, "\n")
		actionContent += "}\n"
	}

	// build children
	for _, child := range metaNode.children {
		if fileMap, err := p.buildClientWithMetaNode(ctx, child); err != nil {
			return nil, err
		} else {
			maps.Copy(ret, fileMap)
		}
	}

	if strings.HasPrefix(metaNode.namespace, "DB") && metaNode.meta == nil {
		// do nothing
	} else if metaNode.namespace == "API" {
		ret[filepath.Join(ctx.output.Dir, "index.ts")] = fmt.Sprintf(
			"%s%s%s",
			importsContent,
			defineContent,
			actionContent,
		)
	} else {
		// write file
		ret[filepath.Join(ctx.output.Dir, currentPackage, "index.ts")] = fmt.Sprintf(
			"%s%s%s",
			importsContent,
			defineContent,
			actionContent,
		)
	}

	return ret, nil
}
