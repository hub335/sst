package python

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sst/sst/v3/internal/fs"
	"github.com/sst/sst/v3/pkg/project/common"
)

func Generate(root string, links common.Links) error {
	projects := fs.FindDown(root, "pyproject.toml")
	files := []io.Writer{}
	for _, project := range projects {
		path := filepath.Join(filepath.Dir(project), "sst.pyi")
		file, _ := os.Create(path)
		files = append(files, file)
	}
	if len(files) == 0 {
		return nil
	}
	properties := map[string]interface{}{}
	for name, link := range links {
		properties[name] = link.Properties
	}

	properties["App"] = map[string]interface{}{
		"name":  "str",
		"stage": "str",
	}

	writer := io.MultiWriter(files...)
	writer.Write([]byte("# Automatically generated by SST\n"))
	writer.Write([]byte("# pylint: disable=all\n"))
	writer.Write([]byte("from typing import Any\n"))
	writer.Write([]byte("\nclass Resource:\n"))
	writer.Write([]byte(infer(properties, "    ") + "\n"))
	for _, file := range files {
		file.(io.WriteCloser).Close()
	}
	return nil
}

func infer(input map[string]interface{}, indentArgs ...string) string {
	indent := ""
	if len(indentArgs) > 0 {
		indent = indentArgs[0]
	}
	var builder strings.Builder
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := input[key]
		switch v := value.(type) {
		case map[string]interface{}:
			// For nested maps, create a new class definition
			builder.WriteString(indent + "class " + key + ":\n")
			builder.WriteString(infer(v, indent+"    "))
		default:
			// Write the field directly if it's not a nested map
			builder.WriteString(indent + key + ": " + inferType(v) + "\n")
		}
	}
	return builder.String()
}

func inferType(value interface{}) string {
	switch value.(type) {
	case string:
		return "str"
	case int:
		return "int"
	case float64, float32:
		return "float"
	case bool:
		return "bool"
	case map[string]interface{}:
		return "dict"
	default:
		return "Any"
	}
}
