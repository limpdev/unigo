## `app.go`

```go
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
```

## `frontend\dist\index.html`

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta content="width=device-width, initial-scale=1.0" name="viewport" />
    <title>uniGo</title>
    <script type="module" crossorigin src="/assets/index.09f92cde.js"></script>
    <link rel="stylesheet" href="/assets/index.72e3db0e.css">
  </head>
  <body>
    <div id="app"></div>

  </body>
</html>
```

## `frontend\index.html`

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta content="width=device-width, initial-scale=1.0" name="viewport" />
    <title>uniGo</title>
  </head>
  <body>
    <div id="app"></div>
    <script src="./src/main.js" type="module"></script>
  </body>
</html>
```

## `frontend\src\App.svelte`

```svelte
<script>
  import { onMount } from "svelte";
  import { ListSvgIcons, ListNerdFontIcons } from "../wailsjs/go/main/App"; // Wails generated bindings

  let svgIcons = [];
  let nerdFontIcons = [];
  let filteredSvgIcons = [];
  let filteredNerdFontIcons = [];
  let isLoading = true;
  let errorMsg = "";
  let currentView = "svg"; // 'svg' or 'nerdfont'
  let searchTerm = "";
  let toastMessage = "";
  let showToast = false;

  onMount(async () => {
    try {
      const [svgs, nfIcons] = await Promise.all([
        ListSvgIcons(),
        ListNerdFontIcons(),
      ]);
      svgIcons = svgs || []; // Ensure it's an array if Go returns null
      nerdFontIcons = nfIcons || [];
      filterIcons();
    } catch (err) {
      console.error("Error loading icons:", err);
      errorMsg = `Failed to load icons: ${err}`;
    } finally {
      isLoading = false;
    }
  });

  function filterIcons() {
    const term = searchTerm.toLowerCase();
    if (svgIcons) {
      filteredSvgIcons = svgIcons.filter((icon) =>
        icon.name.toLowerCase().includes(term)
      );
    } else {
      filteredSvgIcons = [];
    }

    if (nerdFontIcons) {
      filteredNerdFontIcons = nerdFontIcons.filter(
        (icon) =>
          icon.name.toLowerCase().includes(term) ||
          icon.codepoint.toLowerCase().includes(term)
      );
    } else {
      filteredNerdFontIcons = [];
    }
  }

  $: if (searchTerm || svgIcons.length || nerdFontIcons.length) {
    // Re-filter when searchTerm or data changes
    filterIcons();
  }

  function setView(view) {
    currentView = view;
    searchTerm = ""; // Reset search on view change
    filterIcons();
  }

  function displayToast(message) {
    toastMessage = message;
    showToast = true;
    setTimeout(() => {
      showToast = false;
    }, 2000);
  }

  async function copyToClipboard(text, type) {
    try {
      await navigator.clipboard.writeText(text);
      displayToast(`${type} copied!`);
    } catch (err) {
      console.error("Failed to copy: ", err);
      displayToast(`Failed to copy ${type}`);
    }
  }

  function handleSvgClick(icon) {
    // Example: copy SVG name. Could also copy content or path.
    copyToClipboard(icon.name, "SVG name");
    // To copy SVG content: copyToClipboard(icon.content, "SVG content");
  }

  function handleNerdFontClick(icon) {
    // Copy icon name, or codepoint, or the character itself
    const character = String.fromCodePoint(parseInt(icon.codepoint, 16));
    copyToClipboard(character, `NerdFont char ${icon.name}`);
    // To copy codepoint: copyToClipboard(icon.codepoint, "Codepoint");
  }

  function getNerdFontCharacter(codepoint) {
    return String.fromCodePoint(parseInt(codepoint, 16));
  }
</script>

<div class="app-container">
  <aside class="sidebar drag">
    <h2>📖 uniGO</h2>
    <i style="margin-bottom: 1em;font-size:13px;color: #aaa;">Powered by </i>
    <input
      type="text"
      class="search-bar"
      placeholder="Search icons..."
      bind:value={searchTerm}
      on:input={filterIcons}
    />
    <ul>
      <li>
        <button
          class:active={currentView === "svg"}
          on:click={() => setView("svg")}
        >
          SVG <buttondiv id="icon-count">{filteredSvgIcons.length}</buttondiv>
        </button>
      </li>
      <li>
        <button
          class:active={currentView === "nerdfont"}
          on:click={() => setView("nerdfont")}
        >
          Nerd Font <buttondiv id="icon-count">{filteredNerdFontIcons.length}</buttondiv>
        </button>
      </li>
    </ul>
  </aside>

  <main class="main-content">
    {#if isLoading}
      <p>Loading icons...</p>
    {:else if errorMsg}
      <p style="color: red;">{errorMsg}</p>
    {:else if currentView === "svg"}
      {#if filteredSvgIcons.length === 0}
        <p>
          No SVG icons found
          {searchTerm
            ? " matching your search"
            : svgIcons.length === 0
              ? " in content/svgs/. Add some!"
              : ""}.
        </p>
      {:else}
        <div class="icon-grid">
          {#each filteredSvgIcons as icon (icon.path)}
            <div
              class="icon-card"
              aria-label="SVG icon: {icon.name}"
              on:click={() => handleSvgClick(icon)}
              title="Click to copy name: {icon.name}"
            >
              <div class="icon-preview">
                {@html icon.content}
              </div>
              <span class="icon-name">{icon.name}</span>
            </div>
          {/each}
        </div>
      {/if}
    {:else if currentView === "nerdfont"}
      {#if filteredNerdFontIcons.length === 0}
        <p>
          No Nerd Font icons found
          {searchTerm
            ? " matching your search"
            : nerdFontIcons.length === 0
              ? " in content/nerd-fonts/icons.json or font not loaded. Check console & CSS."
              : ""}.
        </p>
      {:else}
        <div class="icon-grid">
          {#each filteredNerdFontIcons as icon (icon.codepoint)}
            <div
              class="icon-card"
              aria-label="SVG icon: {icon.name}"
              on:click={() => handleNerdFontClick(icon)}
              title="Click to copy character: {icon.name}"
            >
              <div class="icon-preview nerd-font-icon-display">
                {getNerdFontCharacter(icon.codepoint)}
              </div>
              <span class="icon-name">{icon.name}</span>
              <span class="icon-name" style="font-size: 0.7em; color: #888;">
                {icon.codepoint}
              </span>
            </div>
          {/each}
        </div>
      {/if}
    {/if}
  </main>
</div>

{#if showToast}
  <div class="toast show">{toastMessage}</div>
{/if}
```

## `frontend\src\style.css`

```css
:root {
  --font-sans: "SF Pro Text", "Symbols Nerd Font", sans-serif;
  --font-mono: "SF Mono", "Symbols Nerd Font Mono", monospace;
}

html {
  background-image: url("assets/images/fabric.png");
  background-color: rgba(25, 25, 25, 0.5);
  text-align: center;
  color: white;
  border-radius: 13px;
  height: 100vh;
}

body {
  background-color: rgba(25, 25, 25, 0.5);
  backdrop-filter: blur(10px);
  margin: 0;
  color: white;
  font-family: var(--font-sans);
  display: block;
  height: 100vh;
}

@font-face {
  font-family: "Nunito";
  font-style: normal;
  font-weight: 400;
  src: local(""),
    url("assets/fonts/nunito-v16-latin-regular.woff2") format("woff2");
}

#app {
  height: 100vh;
  text-align: center;
}

/* Reset and base styles */
:root {
  font-size: 16px;
  line-height: 1.5;
  font-weight: 400;
  color-scheme: light dark;
  color: rgba(255, 255, 255, 0.87);
  /* background-color: #242424; */
  scroll-behavior: smooth;
  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  -webkit-text-size-adjust: 100%;
  --sidebar-width: 200px;
  --header-height: 60px;
  --gap: 1rem;
  --icon-size: 3rem;
  --border-color: #00000000;
}
body {
  margin: 0;
  display: flex;
  min-width: 320px;
  min-height: 100vh;
  box-sizing: border-box;
}
*,
*::before,
*::after {
  box-sizing: inherit;
}
/* App Layout */
.app-container {
  display: flex;
  width: 100%;
  height: 100vh;
}
.sidebar {
  width: var(--sidebar-width);
  /* background-color: #1e1e1e; */
  padding: var(--gap);
  border-right: 0px solid var(--border-color);
  display: block;
  flex-direction: column;
}
.sidebar h2 {
  margin-top: 0;
  font-size: 1em;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 0.5em;
  margin-bottom: 1em;
}
.sidebar ul {
  list-style: none;
  padding: 0;
  margin: 0;
}
.sidebar li button {
  display: block;
  width: 100%;
  margin-top: 0.5em;
  padding: 0.75em 1em;
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.7);
  text-align: left;
  cursor: pointer;
  border-radius: 4px;
  font-size: 0.8em;
  transition: all 300ms ease;
}
.sidebar li button:hover {
  transform: translateX(5px);
  background-color: #333;
  color: white;
}
.sidebar li button.active {
  background-color: #007accaa;
  color: white;
  font-weight: bold;
}
#icon-count {
  font-size: 0.8em;
  color: #f5d863aa;
  float: right;
}
.main-content {
  display: flex;
  flex-grow: 1;
  padding: var(--gap);
  /* overflow-y: auto; */
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100%;
}
.search-bar {
  margin-bottom: var(--gap);
  padding: 0.5em;
  width: 100%;
  max-width: 500px;
  background-color: #333;
  border: 1px solid var(--border-color);
  color: white;
  border-radius: 7px;
}
.icon-grid {
  display: grid;
  width: 100%;
  overflow-y: auto;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: var(--gap);
  flex-grow: 1;
}
.icon-card {
  background-color: #2a2a2a;
  width: 175px;
  border: 1px solid var(--border-color);
  border-radius: 7px;
  padding: var(--gap);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  cursor: pointer;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}
.icon-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
}
.icon-preview {
  width: var(--icon-size);
  height: var(--icon-size);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.5em;
  overflow: hidden; /* Prevents large SVGs from breaking layout */
}
.icon-preview svg {
  max-width: 100%;
  max-height: 100%;
  fill: currentColor; /* For SVGs that use currentColor */
}
.icon-name {
  font-size: 0.8em;
  word-break: break-all;
  color: #ccc;
}
.toast {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background-color: #333;
  color: white;
  padding: 10px 20px;
  border-radius: 5px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
  z-index: 1000;
  opacity: 0;
  transition: opacity 0.5s ease-in-out;
}
.toast.show {
  opacity: 0;
}
/* NERD FONT STYLING */
/*
  IMPORTANT:
  1. Replace 'YourNerdFontName' with a descriptive name for your font.
  2. Replace 'YourNerdFont.ttf' with the ACTUAL FILENAME of your font
     in frontend/src/assets/fonts/
*/
@font-face {
  font-family: "Symbols Nerd Font"; /* Choose a name */
  src: url("./assets/fonts/SymbolsNerdFont-Regular.ttf") format("truetype"); /* UPDATE THIS PATH */
  /* If you have other formats like woff2, add them:
  src: url('./assets/fonts/YourNerdFont.woff2') format('woff2'),
       url('./assets/fonts/YourNerdFont.ttf') format('truetype');
  */
  font-weight: normal;
  font-style: normal;
}
.nerd-font-icon-display {
  font-family: "Symbols Nerd Font", monospace; /* Fallback chain */
  font-size: var(--icon-size); /* Adjust as needed */
  line-height: 1;
}
```

## `main.go`

```go
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp() // This NewApp() comes from app.go

	// Create application with options
	err := wails.Run(&options.App{
		Title:                    "uniGO",
		Width:                    768,
		Height:                   1024,
		MinWidth:                 600,
		MinHeight:                800,
		Fullscreen:               false,
		CSSDragProperty:          "drag",
		EnableDefaultContextMenu: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:        app.startup, // Use the startup method from app.go
		Bind: []interface{}{
			app, // Bind the entire app struct
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:  true,
			CSSDropProperty: "dropper",
			CSSDropValue:    "drop",
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "uniGO",
				Message: "MIT",
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			BackdropType:                      windows.Acrylic,
			DisablePinchZoom:                  true,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
```
