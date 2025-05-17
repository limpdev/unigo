package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// // DESIGNATED ASSET *HANDLER*
// type Loader struct {
// 	http.Handler
// }

// func iconLoader() *Loader {
// 	return &Loader{}
// }

// func (h *Loader) ServeHTTP(res http.ResponseWriter, req *http.Request) {
// 	var err error
// 	reqIcon := strings.TrimPrefix(req.URL.Path, "/")
// 	println("REQ: Requesting icon:", reqIcon)
// 	iconData, err := os.ReadFile(reqIcon)
// 	if err != nil {
// 		res.WriteHeader(http.StatusBadRequest)
// 		res.Write([]byte(fmt.Sprintf("Couldn't load file %s", reqIcon)))
// 	}
// 	res.Write(iconData)
// }

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SvgIcon represents a single SVG icon
type SvgIcon struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Path    string `json:"path"`
}

// NerdFontIcon represents a single Nerd Font icon
type NerdFontIcon struct {
	Name      string `json:"name"`
	Codepoint string `json:"codepoint"` // Hex codepoint string
}

// ListSvgIcons scans the content/svgs directory for SVG files
func (a *App) ListSvgIcons() ([]SvgIcon, error) {
	svgDir := "content/svgs" // Relative to project root
	files, err := os.ReadDir(svgDir)
	if err != nil {
		if os.IsNotExist(err) {
			// If directory doesn't exist, return empty list, not an error for the frontend
			return []SvgIcon{}, nil
		}
		return nil, fmt.Errorf("failed to read SVG directory %s: %w", svgDir, err)
	}

	var icons []SvgIcon
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".svg") {
			filePath := filepath.Join(svgDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				// Log error but continue processing other files
				fmt.Printf("Error reading SVG file %s: %v\n", filePath, err)
				continue
			}
			icons = append(icons, SvgIcon{
				Name:    strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())),
				Content: string(content),
				Path:    filePath,
			})
		}
	}
	return icons, nil
}

// ListNerdFontIcons loads Nerd Font icon definitions from content/nerd-fonts/icons.json
func (a *App) ListNerdFontIcons() ([]NerdFontIcon, error) {
	jsonPath := "content/nerd-fonts/icons.json" // Relative to project root
	file, err := os.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, return empty list
			return []NerdFontIcon{}, nil
		}
		return nil, fmt.Errorf("failed to read Nerd Font JSON %s: %w", jsonPath, err)
	}

	var icons []NerdFontIcon
	err = json.Unmarshal(file, &icons)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Nerd Font JSON %s: %w", jsonPath, err)
	}
	return icons, nil
}
