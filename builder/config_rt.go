package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type RTOutputConfig struct {
	Kind       string `json:"kind" required:"true"`
	Language   string `json:"language" required:"true"`
	Dir        string `json:"dir" required:"true"`
	GoModule   string `json:"goModule"`
	GoPackage  string `json:"goPackage"`
	HttpEngine string `json:"httpEngine"`
}

type RTConfig struct {
	Listen       string            `json:"listen"`
	Outputs      []*RTOutputConfig `json:"outputs"`
	DB           *DBConfig         `json:"db"`
	__filepath__ string
}

func (c *RTConfig) GetFilePath() string {
	return c.__filepath__
}

func LoadRTConfig() (*RTConfig, error) {
	configPath := ""

	if len(os.Args) > 1 {
		if fileInfo, err := os.Stat(os.Args[1]); err == nil && !fileInfo.IsDir() {
			configPath = os.Args[1]
		}
	}

	if configPath == "" {
		// 在当前目录下，依次寻找 .capi.json .capi.yaml .capi.yml
		searchFiles := []string{"./.capi.json", "./.capi.yaml", "./.capi.yml"}
		for _, file := range searchFiles {
			if fileInfo, err := os.Stat(file); err == nil && !fileInfo.IsDir() {
				configPath = file
				break
			}
		}
	}

	if !filepath.IsAbs(configPath) {
		if absPath, err := filepath.Abs(configPath); err != nil {
			return nil, fmt.Errorf("failed to convert config path to absolute path: %v", err)
		} else {
			configPath = absPath
		}
	}

	var config RTConfig

	if err := UnmarshalConfig(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	projectDir := filepath.Dir(configPath)

	for i, output := range config.Outputs {
		var err error
		config.Outputs[i].Dir, err = ParseProjectDir(output.Dir, projectDir)
		if err != nil {
			return nil, fmt.Errorf("failed to parse output dir: %w", err)
		}

		if !filepath.IsAbs(config.Outputs[i].Dir) {
			config.Outputs[i].Dir = filepath.Join(projectDir, config.Outputs[i].Dir)
		}

		if config.Outputs[i].GoModule != "" && config.Outputs[i].GoPackage == "" {
			goModuleArr := strings.Split(config.Outputs[i].GoModule, "/")
			config.Outputs[i].GoPackage = goModuleArr[len(goModuleArr)-1]
		}
	}

	config.__filepath__ = configPath

	return &config, nil
}
