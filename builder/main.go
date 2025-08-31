package builder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
)

type IBuilder interface {
	Prepare() error
	BuildServer() error
	BuildClient() error
}

type BuildContext struct {
	location string
	rtConfig *RTConfig
	apiMetas []*APIMeta
	dbMetas  []*DBTableMeta
	output   *RTOutputConfig
}

func Build() error {
	rtConfig, err := LoadRTConfig()
	if err != nil {
		log.Panicf("Failed to load capi config: %v", err)
	}

	projectDir := filepath.Dir(rtConfig.GetFilePath())
	log.Printf("capi: project dir: %s\n", projectDir)
	log.Printf("capi: meta file: %s\n", rtConfig.GetFilePath())

	apiMetas := []*APIMeta{}
	dbMetas := []*DBTableMeta{}

	for _, output := range rtConfig.Outputs {
		walkErr := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			var header struct {
				Version string `json:"version"`
			}

			switch filepath.Ext(path) {
			case ".json", ".yaml", ".yml":
				if err := UnmarshalConfig(path, &header); err != nil {
					return nil // Not a rt meta file, just ignore.  continue walking
				} else if slices.Contains(APIVersions, header.Version) {
					if apiConfig, err := LoadAPIMeta(path); err != nil {
						return err
					} else {
						apiMetas = append(apiMetas, apiConfig)
						return nil
					}
				} else if slices.Contains(DBVersions, header.Version) {
					if dbConfig, err := LoadDBTableMeta(path); err != nil {
						return err
					} else {
						dbMetas = append(dbMetas, dbConfig)
						return nil
					}
				} else {
					return nil
				}
			default:
				return nil
			}
		})

		if walkErr != nil {
			return fmt.Errorf("error walking project directory: %w", walkErr)
		}

		var builder IBuilder

		context := BuildContext{
			location: MainLocation,
			rtConfig: rtConfig,
			apiMetas: apiMetas,
			dbMetas:  dbMetas,
			output:   output,
		}

		switch output.Language {
		case "go":
			builder = &GoBuilder{
				BuildContext: context,
			}
		case "typescript":
			builder = &TypescriptBuilder{
				BuildContext: context,
			}
		default:
			return fmt.Errorf("unsupported language: %s", context.output.Language)
		}

		if err := builder.Prepare(); err != nil {
			return err
		}

		switch context.output.Kind {
		case "server":
			if err := builder.BuildServer(); err != nil {
				return err
			}
		case "client":
			if err := builder.BuildClient(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported kind: %s", context.output.Kind)
		}
	}

	return nil
}
