#  uniGo 

<details><summary>📦<i>TRIAL ONE</i></summary>

```bash
.
├── app.go
├── build
│  ├── appicon.png
│  ├── darwin
│  │  ├── Info.dev.plist
│  │  └── Info.plist
│  ├── README.md
│  └── windows
│     ├── icon.ico
│     ├── info.json
│     ├── installer
│     │  ├── project.nsi
│     │  └── wails_tools.nsh
│     └── wails.exe.manifest
├── content
├── frontend
│  ├── dist
│  ├── index.html
│  ├── jsconfig.json
│  ├── package.json
│  ├── README.md
│  ├── src
│  │  ├── App.svelte
│  │  ├── assets
│  │  │  ├── fonts
│  │  │  │  ├── nunito-v16-latin-regular.woff2
│  │  │  │  └── OFL.txt
│  │  │  └── images
│  │  │     └── logo-universal.png
│  │  ├── main.js
│  │  ├── style.css
│  │  └── vite-env.d.ts
│  └── vite.config.js
├── go.mod
├── go.sum
├── main.go
├── README.md
├── TREE
└── wails.json
```

---

## Get Started

*We'll focus on:*

1.  **Go Backend (`app.go`)**:
    -   Reading SVG files from `content/svgs/`.
    -   Reading Nerd Font icon definitions (we'll assume a JSON file for this, as parsing font files directly is complex) from `content/nerd-fonts/`.
2.  **Svelte Frontend (`App.svelte`)**:
    - Displaying a sidebar for navigation (SVGs, Nerd Fonts).
    - Displaying a grid of SVG previews.
    - Displaying a grid of Nerd Font icons (using the actual font).
    - Basic search/filter.
    - Copying icon name/SVG content/codepoint.
3.  **Project Structure & Assets**:
    -   We'll slightly adjust the `content` folder structure for clarity.
    -   You'll need to place your Nerd Font file (e.g., `.ttf` or `.otf`) into the `frontend/src/assets/fonts` directory so CSS can load it.

Let's go!

### Update Project Structure (Recommended)**

Modify your `content` directory and `frontend/src/assets/fonts` like this:

```bash
.
├── app.go
├── build
│  └── ... (rest of build)
├── content
│  ├── nerd-fonts           # For Nerd Font metadata
│  │  └── icons.json        # We'll define this format
│  └── svgs                 # For user-provided SVG files
│     └── example.svg       # Add a sample SVG here
├── frontend
│  ├── ...
│  ├── src
│  │  ├── App.svelte
│  │  ├── assets
│  │  │  ├── fonts
│  │  │  │  ├── YourNerdFont.ttf  # <--- PLACE YOUR NERD FONT FILE HERE
│  │  │  │  └── OFL.txt           # (if applicable)
│  │  │  └── images
│  │  │     └── logo-universal.png
│  │  ├── main.js
│  │  ├── style.css
│  │  └── vite-env.d.ts
│  └── vite.config.js
├── go.mod
├── go.sum
├── main.go
├── README.md
├── TREE
└── wails.json
```

### Create Sample Content Files**

> **`content/svgs/example.svg`**:

```html
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="50" height="50">
  <circle cx="50" cy="50" r="40" stroke="black" stroke-width="3" fill="red" />
  <text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle" font-size="12"fill="white">SVG!</text>
</svg>
```

*(Add more SVGs here later)*

*  **`content/nerd-fonts/icons.json`**:
*
This file will map icon names to their Unicode codepoints. You'll need to generate or find sucha JSON for your specific Nerd Font. Many Nerd Font download pages or GitHub repos include acheatsheet or a way to get this data.
Example (using some Font Awesome icons commonly included in Nerd Fonts):

```json
[
  { "name": "nf-fa-apple", "codepoint": "f179" },
  { "name": "nf-fa-android", "codepoint": "f17b" },
  { "name": "nf-fa-windows", "codepoint": "f17a" },
  { "name": "nf-dev-github_badge", "codepoint": "f09b" },
  { "name": "nf-custom-heart", "codepoint": "f004" },
  { "name": "nf-mdi-home", "codepoint": "f2dc" }
]
```

*Important*: The `codepoint` is the hexadecimal Unicode value *without* `U+` or `\u`.

> **`frontend/src/assets/fonts/YourNerdFont.ttf`**:

Download a Nerd Font (e.g., Fira Code Nerd Font, Hack Nerd Font) and place the `.ttf` or `.otf`file here. **Rename it** if needed, and make sure to update the CSS later.

### Go Backend (`app.go`)**

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

**4. Update `main.go`**

Make sure your `main.go` initializes and binds your `App` struct.

```go
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp() // This NewApp() comes from app.go

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Icon Library Previewer",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup, // Use the startup method from app.go
		Bind: []interface{}{
			app, // Bind the entire app struct
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
```

### Svelte Frontend (`frontend/src/`)**

> **`frontend/src/style.css`**:

```css
/* Reset and base styles */
:root {
  font-family: Inter, Avenir, Helvetica, Arial, sans-serif;
  font-size: 16px;
  line-height: 1.5;
  font-weight: 400;
  color-scheme: light dark;
  color: rgba(255, 255, 255, 0.87);
  background-color: #242424;
  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  -webkit-text-size-adjust: 100%;
  --sidebar-width: 250px;
  --header-height: 60px;
  --gap: 1rem;
  --icon-size: 3rem;
  --border-color: #444;
}
body {
  margin: 0;
  display: flex;
  min-width: 320px;
  min-height: 100vh;
  box-sizing: border-box;
}
*, *::before, *::after {
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
  background-color: #1e1e1e;
  padding: var(--gap);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}
.sidebar h2 {
  margin-top: 0;
  font-size: 1.2em;
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
  padding: 0.75em 1em;
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.7);
  text-align: left;
  cursor: pointer;
  border-radius: 4px;
  font-size: 1em;
}
.sidebar li button:hover {
  background-color: #333;
  color: white;
}
.sidebar li button.active {
  background-color: #007acc;
  color: white;
  font-weight: bold;
}
.main-content {
  flex-grow: 1;
  padding: var(--gap);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.search-bar {
  margin-bottom: var(--gap);
  padding: 0.5em;
  width: 100%;
  max-width: 500px;
  background-color: #333;
  border: 1px solid var(--border-color);
  color: white;
  border-radius: 4px;
}
.icon-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: var(--gap);
}
.icon-card {
  background-color: #2a2a2a;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: var(--gap);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.icon-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 4px 8px rgba(0,0,0,0.3);
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
  box-shadow: 0 2px 10px rgba(0,0,0,0.2);
  z-index: 1000;
  opacity: 0;
  transition: opacity 0.5s ease-in-out;
}
.toast.show {
  opacity: 1;
}
/* NERD FONT STYLING */
/*
  IMPORTANT:
  1. Replace 'YourNerdFontName' with a descriptive name for your font.
  2. Replace 'YourNerdFont.ttf' with the ACTUAL FILENAME of your font
     in frontend/src/assets/fonts/
*/
@font-face {
  font-family: 'MyNerdFont'; /* Choose a name */
  src: url('./assets/fonts/YourNerdFont.ttf') format('truetype'); /* UPDATE THIS PATH */
  /* If you have other formats like woff2, add them:
  src: url('./assets/fonts/YourNerdFont.woff2') format('woff2'),
       url('./assets/fonts/YourNerdFont.ttf') format('truetype');
  */
  font-weight: normal;
  font-style: normal;
}
.nerd-font-icon-display {
  font-family: 'MyNerdFont', 'Symbols Nerd Font', monospace; /* Fallback chain */
  font-size: var(--icon-size); /* Adjust as needed */
  line-height: 1;
}
```

> **VERY IMPORTANT**: In the `@font-face` rule:

1.  Change `'MyNerdFont'` to a name you want to use (e.g., `'FiraCodeNerd'`).
2.  Change `url('./assets/fonts/YourNerdFont.ttf')` to match the *exact filename* of the NerdFontyou placed in `frontend/src/assets/fonts/`.
3.  If your font is `.otf`, use `format('opentype')`.

> **`frontend/src/App.svelte`**:

```html
<script>
  import { onMount } from 'svelte';
  import { ListSvgIcons, ListNerdFontIcons } from '../wailsjs/go/main/App'; // Wailsgeneratedbindings
  let svgIcons = [];
  let nerdFontIcons = [];
  let filteredSvgIcons = [];
  let filteredNerdFontIcons = [];
  let isLoading = true;
  let errorMsg = '';
  let currentView = 'svg'; // 'svg' or 'nerdfont'
  let searchTerm = '';
  let toastMessage = '';
  let showToast = false;
  onMount(async () => {
    try {
      const [svgs, nfIcons] = await Promise.all([
        ListSvgIcons(),
        ListNerdFontIcons()
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
        filteredSvgIcons = svgIcons.filter(icon => icon.name.toLowerCase().includes(term));
    } else {
        filteredSvgIcons = [];
    }
    if (nerdFontIcons) {
        filteredNerdFontIcons = nerdFontIcons.filter(icon =>
            icon.name.toLowerCase().includes(term) ||
            icon.codepoint.toLowerCase().includes(term)
        );
    } else {
        filteredNerdFontIcons = [];
    }
  }
  $: if (searchTerm || svgIcons.length || nerdFontIcons.length) { // Re-filter when searchTermordata changes
      filterIcons();
  }
  function setView(view) {
    currentView = view;
    searchTerm = ''; // Reset search on view change
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
      console.error('Failed to copy: ', err);
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
  <aside class="sidebar">
    <h2>Icon Library</h2>
    <input
      type="text"
      class="search-bar"
      placeholder="Search icons..."
      bind:value={searchTerm}
      on:input={filterIcons}
    />
    <ul>
      <li>
        <button class:active={currentView === 'svg'} on:click={() => setView('svg')}>
          SVG Icons ({filteredSvgIcons.length})
        </button>
      </li>
      <li>
        <button class:active={currentView === 'nerdfont'} on:click={() => setView('nerdfont')}>
          Nerd Font Icons ({filteredNerdFontIcons.length})
        </button>
      </li>
    </ul>
  </aside>
  <main class="main-content">
    {#if isLoading}
      <p>Loading icons...</p>
    {:else if errorMsg}
      <p style="color: red;">{errorMsg}</p>
    {:else}
      {#if currentView === 'svg'}
        {#if filteredSvgIcons.length === 0}
          <p>No SVG icons found{searchTerm ? ' matching your search' : (svgIcons.length === 0 ?'in content/svgs/. Add some!' : '')}.</p>
        {:else}
          <div class="icon-grid">
            {#each filteredSvgIcons as icon (icon.path)}
              <div class="icon-card" on:click={() => handleSvgClick(icon)} title="Click tocopyname: {icon.name}">
                <div class="icon-preview">
                  {@html icon.content}
                </div>
                <span class="icon-name">{icon.name}</span>
              </div>
            {/each}
          </div>
        {/if}
      {:else if currentView === 'nerdfont'}
        {#if filteredNerdFontIcons.length === 0}
          <p>No Nerd Font icons found{searchTerm ? ' matching your search' : (nerdFontIconslength=== 0 ? ' in content/nerd-fonts/icons.json or font not loaded. Check console &CSS.' :'')}.</p>
        {:else}
          <div class="icon-grid">
            {#each filteredNerdFontIcons as icon (icon.codepoint)}
              <div class="icon-card" on:click={() => handleNerdFontClick(icon)} title="Clicktocopy character: {icon.name}">
                <div class="icon-preview nerd-font-icon-display">
                  {getNerdFontCharacter(icon.codepoint)}
                </div>
                <span class="icon-name">{icon.name}</span>
                <span class="icon-name" style="font-size: 0.7em; color: #888;">{icon.codepoint<span>
              </div>
            {/each}
          </div>
        {/if}
      {/if}
    {/if}
  </main>
</div>
{#if showToast}
  <div class="toast show">{toastMessage}</div>
{/if}
```

*   **`frontend/wailsjs/go/main/App.d.ts` (Auto-generated - check after `wails dev`)**
    Wails should generate this. If you make changes to Go methods exposed to the frontend, you might need to delete `frontend/wailsjs` and restart `wails dev` for it to regenerate.

**6. Build and Run**

1.  Open your project root in the terminal.
2.  Run `wails dev`.

    *   This will compile the Go backend and start the Vite dev server for the Svelte frontend.
    *   Wails automatically generates the JavaScript bindings in `frontend/wailsjs/`.
3.  Your application window should appear.

**Key Things to Double-Check / Troubleshoot:**

*   **Nerd Font Path in CSS:** The `url(...)` in `frontend/src/style.css` for `@font-face` MUST correctly point to your Nerd Font file within `frontend/src/assets/fonts/`. The font name used in `font-family` in `@font-face` must match the one used in `.nerd-font-icon-display`.
*   **`content/nerd-fonts/icons.json`:** Ensure this file exists, is valid JSON, and its format matches what the Go code expects (`[{"name": "...", "codepoint": "..."}, ...]`).
*   **`content/svgs/`:** Make sure you have some `.svg` files there.
*   **Console Output:** Check both the terminal where `wails dev` is running (for Go errors) and the browser developer console (Ctrl+Shift+I or Cmd+Opt+I, for Svelte/JavaScript errors, especially related to font loading or Wails calls).
*   **Wails Bindings:** If you change function signatures in `app.go` that are bound to the frontend, Wails needs to regenerate the bindings. Sometimes deleting the `frontend/wailsjs` directory and restarting `wails dev` helps.

This comprehensive setup should give you a solid foundation. You can now expand on this by adding more features like:
*   More sophisticated search (e.g., fuzzy search).
*   Tagging icons.
*   User-configurable settings.
*   Different views (list vs. grid).
*   Directly opening SVG file location.

</details>

---

## 📜TRIAL TWO

The main layout issues ("main content won't fill the window" and "grid stuck to one column") are indeed CSS-related, primarily stemming from how heights and widths are being managed in the flex/grid containers, and some conflicting global styles.

Let's break it down:

**Critical CSS Fixes for Layout:**

1.  **`html`, `body`, `#app` Base Styling:**
    *   `text-align: center;` on `html` and `#app` can interfere with block-level children's default width behavior. It's meant for inline content.
    *   Managing `height: 100vh` vs `height: 100%`: `100vh` is viewport height. For children to fill a parent that has `100vh`, they should use `height: 100%`.

2.  **`.main-content` Height:**
    *   You have `height: 100vh;` on `.main-content`. Since its parent (`.app-container`) is already `height: 100vh;`, and `.main-content` is a flex item (`flex-grow: 1`), it should derive its height from the flex layout. The explicit `100vh` can cause overflow or miscalculation, especially when padding is involved.

3.  **`.icon-grid` Sizing:**
    *   For `.icon-grid` to be scrollable and use available space, its parent (`.main-content`) needs to correctly constrain its height. Since `.main-content` is `display: flex; flex-direction: column;`, the `.icon-grid` (with `flex-grow: 1`) should expand to fill the remaining vertical space.
    *   The grid being one column means its *width* is somehow constrained. This often happens if its parent (`.main-content`) isn't getting enough width from *its* parent (`.app-container`).

**Other CSS/HTML Bugs & Improvements:**

4.  **Invalid HTML `<buttondiv>`:**
    *   In `App.svelte`, you have `<buttondiv id="icon-count">`. This is not a valid HTML element. It should be a `<span>` or `<div>` *inside* the `<button>`, or just styled differently.

5.  **Toast Opacity:**
    *   `.toast.show` has `opacity: 0;`. This should be `opacity: 1;` to make it visible.

6.  **Duplicate `:root`:**
    *   You have two `:root` declarations in `style.css`. The second one will override variables from the first if they have the same name. It's better to consolidate them.

7.  **Font Definitions & Usage:**
    *   You define `Nunito` but then in `:root` your `--font-sans` is `"SF Pro Text", "Symbols Nerd Font", sans-serif;`. Make sure you're applying the font you intend. The Nerd Font for icons seems correctly set up with `Symbols Nerd Font`.

8.  **`.sidebar` display:**
    *   `display: block; flex-direction: column;` is unusual. If you want the sidebar to be a flex container for its children (like `h2`, `input`, `ul`), it should be `display: flex; flex-direction: column;`. If not, `flex-direction` has no effect on `display: block`. Given its children, `display: flex; flex-direction: column;` is likely more appropriate.

**Wails `main.go` Options:**

*   Your `Mac.TitleBar` options are set for a standard title bar (`TitlebarAppearsTransparent: false`, `HideTitleBar: false`).
*   However, you have `CSSDragProperty: "drag"` and apply `class="sidebar drag"` in Svelte. This implies you want to drag the window by the sidebar.
*   If you want a frameless window (no native title bar) that you drag by the sidebar:
    *   Set `Frameless: true` in `options.App{}`.
    *   Then, for Mac, you'd typically set `Mac.TitleBar.TitlebarAppearsTransparent: true` and `Mac.TitleBar.FullSizeContent: true`. You might even `HideTitle: true`.
*   If you want to keep the standard OS title bar, then the `CSSDragProperty` on the sidebar won't move the window (it might select text).
*   **For now, I'll assume you want the standard title bar since that's what the explicit options suggest.** The drag property on the sidebar won't hurt but won't move the window. If you want a frameless window, you'll need to adjust `main.go`.

---

**Corrected Code Snippets:**

**1. `frontend\src\style.css` (Key Changes Highlighted)**

```css
/* Consolidated :root and base styles */
:root {
  /* Primary font for UI text - choose one or define fallbacks */
  --font-sans: "Nunito", "SF Pro Text", "Symbols Nerd Font", sans-serif; /* Added Nunito here */
  --font-mono: "SF Mono", "Symbols Nerd Font Mono", monospace;

  font-size: 16px;
  line-height: 1.5;
  font-weight: 400;

  color-scheme: light dark;
  color: rgba(255, 255, 255, 0.87);

  scroll-behavior: smooth;
  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  -webkit-text-size-adjust: 100%;

  --sidebar-width: 200px;
  --header-height: 60px; /* Not currently used, but good to keep if you add a header */
  --gap: 1rem;
  --icon-size: 3rem;
  --border-color: #00000000; /* Transparent border is fine */
}

/* Ensure html, body, and #app take full height and don't interfere with flex/grid */
html {
  height: 100vh;
  background-image: url("assets/images/fabric.png");
  background-color: rgba(25, 25, 25, 0.5); /* This will be under body's background */
  /* text-align: center;  <--- REMOVE for block layouts */
  color: white;
  border-radius: 13px; /* This applies to the HTML element itself */
  overflow: hidden; /* Prevent scrollbars on html if window is sized perfectly */
}

body {
  height: 100%; /* Fill html */
  margin: 0;
  font-family: var(--font-sans);
  background-color: rgba(25, 25, 25, 0.5); /* Overlays html background */
  backdrop-filter: blur(10px);
  color: white;
  display: flex; /* Make body a flex container */
  flex-direction: column; /* Ensure #app can grow if needed */
  overflow: hidden; /* Prevent scrollbars on body */
}

#app {
  height: 100%; /* Fill body */
  width: 100%;
  display: flex; /* Make #app a flex container for .app-container */
  /* text-align: center; <--- REMOVE for block layouts */
}


@font-face {
  font-family: "Nunito";
  font-style: normal;
  font-weight: 400;
  src: local(""),
    url("assets/fonts/nunito-v16-latin-regular.woff2") format("woff2");
}

/* App Layout */
.app-container {
  display: flex;
  width: 100%; /* Fill #app */
  height: 100%; /* Fill #app */
}

.sidebar {
  width: var(--sidebar-width);
  padding: var(--gap);
  border-right: 0px solid var(--border-color); /* This is fine if you want no visible border */
  display: flex; /* Make sidebar a flex container for its children */
  flex-direction: column;
  flex-shrink: 0; /* Prevent sidebar from shrinking */
}

.sidebar h2 {
  margin-top: 0;
  font-size: 1.2em; /* Slightly larger for better readability */
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
  display: flex; /* Use flex to align items inside button */
  justify-content: space-between; /* Pushes count to the right */
  align-items: center; /* Vertically align text and count */
  width: 100%;
  margin-top: 0.5em;
  padding: 0.75em 1em;
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.7);
  text-align: left;
  cursor: pointer;
  border-radius: 4px;
  font-size: 0.9em; /* Adjusted for balance */
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

/* Replaced #icon-count with a class for spans */
.icon-count-badge {
  font-size: 0.8em;
  color: #f5d863aa;
  background-color: rgba(0,0,0,0.2);
  padding: 0.1em 0.4em;
  border-radius: 3px;
}

.main-content {
  flex-grow: 1; /* Allow main content to take remaining width */
  padding: var(--gap);
  display: flex;
  flex-direction: column;
  /* height: 100vh; <--- REMOVE THIS, height comes from flex parent */
  /* width: 100%;   <--- Not strictly needed with flex-grow, but harmless */
  overflow: hidden; /* Important: parent of scrollable area should hide overflow */
  min-width: 0; /* Fixes potential flexbox shrinkage issue with wide children */
}

.search-bar {
  margin-bottom: var(--gap);
  padding: 0.5em;
  width: 100%;
  max-width: 500px; /* Or remove max-width if you want it to span more */
  background-color: #333;
  border: 1px solid var(--border-color);
  color: white;
  border-radius: 7px;
  flex-shrink: 0; /* Prevent search bar from shrinking */
}

.icon-grid {
  display: grid;
  width: 100%; /* Take full width of .main-content */
  /*
    IMPORTANT: For the grid to scroll, its parent (.main-content) defines the
    overall height, and the grid itself takes the remaining space and scrolls.
  */
  flex-grow: 1; /* Make the grid take all available vertical space in .main-content */
  overflow-y: auto; /* Enable vertical scrolling FOR THE GRID */
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); /* Increased minmax slightly */
  gap: var(--gap);
}

.icon-card {
  background-color: #2a2a2a;
  /* width: 175px; <--- REMOVE fixed width; let grid-template-columns handle it */
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
  min-height: 120px; /* Give cards a minimum height */
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
  overflow: hidden;
}

.icon-preview svg {
  max-width: 100%;
  max-height: 100%;
  fill: currentColor;
}

.icon-name {
  font-size: 0.8em;
  word-break: break-all;
  color: #ccc;
  margin-top: auto; /* Pushes name to bottom if card content is sparse */
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
  pointer-events: none; /* So it doesn't intercept clicks when hidden */
}

.toast.show {
  opacity: 1; /* <<< CORRECTED */
}

/* NERD FONT STYLING */
@font-face {
  font-family: "Symbols Nerd Font"; /* This name is used below */
  src: url("./assets/fonts/SymbolsNerdFont-Regular.ttf") format("truetype");
  font-weight: normal;
  font-style: normal;
}

.nerd-font-icon-display {
  font-family: "Symbols Nerd Font", monospace; /* Fallback chain */
  font-size: var(--icon-size);
  line-height: 1;
}
```

**2. `frontend\src\App.svelte` (Changes for icon count display)**

```html
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
      svgIcons = svgs || [];
      nerdFontIcons = nfIcons || [];
      filterIcons();
    } catch (err) {
      console.error("Error loading icons:", err);
      errorMsg = `Failed to load icons: ${err.message || err}`; // Get message from error
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
    filterIcons();
  }

  function setView(view) {
    currentView = view;
    searchTerm = "";
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
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(text);
        displayToast(`${type} copied!`);
      } else {
        displayToast('Clipboard API not available.');
      }
    } catch (err) {
      console.error("Failed to copy: ", err);
      displayToast(`Failed to copy ${type}`);
    }
  }

  function handleSvgClick(icon) {
    copyToClipboard(icon.name, "SVG name");
  }

  function handleNerdFontClick(icon) {
    const character = String.fromCodePoint(parseInt(icon.codepoint, 16));
    copyToClipboard(character, `NerdFont char ${icon.name}`);
  }

  function getNerdFontCharacter(codepoint) {
    return String.fromCodePoint(parseInt(codepoint, 16));
  }
</script>

<div class="app-container">
  <aside class="sidebar drag"> {/* 'drag' property for Wails if window is frameless */}
    <h2>📖 uniGO</h2>
    <i style="margin-bottom: 1em;font-size:13px;color: #aaa;">Powered by <span class="nerd-font-icon-display" style="font-size: 13px;"></span></i> {/* Using span for Wails icon */}
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
          <span>SVG</span> {/* Text in a span */}
          <span class="icon-count-badge">{filteredSvgIcons.length}</span> {/* Count in a styled span */}
        </button>
      </li>
      <li>
        <button
          class:active={currentView === "nerdfont"}
          on:click={() => setView("nerdfont")}
        >
          <span>Nerd Font</span>
          <span class="icon-count-badge">{filteredNerdFontIcons.length}</span>
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
              aria-label="Nerd Font icon: {icon.name}"
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

**Summary of Key Fixes Applied:**
1.  **CSS:**
    *   Removed `text-align: center` from `html` and `#app`.
    *   Set `html`, `body`, `#app` to `height: 100vh` (html) and `height: 100%` (body, app) and `display: flex` to properly propagate dimensions.
    *   Removed explicit `height: 100vh` from `.main-content`; its height is now correctly determined by `flex-grow: 1` within `.app-container`.
    *   Added `min-width: 0` to `.main-content` (a common flexbox fix for nested content).
    *   Added `overflow: hidden` to `html`, `body`, and `.main-content` to control scroll contexts.
    *   Made `.icon-grid` use `flex-grow: 1` and `overflow-y: auto` to ensure it scrolls within the space allocated by `.main-content`.
    *   Removed fixed `width` from `.icon-card`, allowing `grid-template-columns` to define their size.
    *   Corrected `.toast.show { opacity: 1; }`.
    *   Changed `.sidebar` to `display: flex; flex-direction: column;`.
    *   Changed icon count display in buttons to use `<span>` and flex layout on the button.
2.  **Svelte:**
    *   Replaced `<buttondiv>` with `<span>` and classes for styling the count.
    *   Ensured the Wails icon in the sidebar also uses the Nerd Font class if it's a Nerd Font glyph.

After applying these changes, rebuild (`wails dev` or `wails build`):
*   The main content area should now correctly fill the space next to the sidebar.
*   The icon grid should display multiple columns and scroll vertically if there are many icons.

Test thoroughly. Flexbox and Grid layouts can be tricky, and sometimes small changes have big effects, especially concerning overflow and height/width calculations.
