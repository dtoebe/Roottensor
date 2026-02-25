package httpserver

import (
	"html/template"
	"path/filepath"
)

type Templates struct {
	index *template.Template
}

func LoadTemplates(root string) (*Templates, error) {
	layoutPath := filepath.Join(root, "layout.html")

	layout, err := template.ParseFiles(layoutPath)
	if err != nil {
		return nil, err
	}

	return &Templates{
		index: layout,
	}, nil
}
