## `app.go`

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

// DESIGNATED ASSET *HANDLER*
type Loader struct {
	http.Handler
}

func iconLoader() *Loader {
	return &Loader{}
}

func (h *Loader) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var err error
	reqIcon := strings.TrimPrefix(req.URL.Path, "/")
	println("REQ: Requesting icon:", reqIcon)
	iconData, err := os.ReadFile(reqIcon)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(fmt.Sprintf("Couldn't load file %s", reqIcon)))
	}
	res.Write(iconData)
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

## `archive\uniGo.go`

```go
// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/unicode/runenames" // For character names
)

// CharacterInfo holds data about a single Unicode character
type CharacterInfo struct {
	Char       string `json:"char"`
	CodePoint  string `json:"codePoint"` // Hex representation like U+0041
	Name       string `json:"name"`
	Category   string `json:"category"`
	CategoryAb string `json:"categoryAb"` // Abbreviation like Lu
	// BlockName string `json:"blockName"` // Block info is harder to get reliably without external data
}

// APIResponse structures the JSON response for the characters endpoint
type APIResponse struct {
	Characters   []CharacterInfo `json:"characters"`
	TotalItems   int             `json:"totalItems"`
	CurrentPage  int             `json:"currentPage"`
	ItemsPerPage int             `json:"itemsPerPage"`
	TotalPages   int             `json:"totalPages"`
}

// MetadataResponse structures the JSON response for metadata
type MetadataResponse struct {
	Categories map[string]string `json:"categories"` // Map Abbreviation -> Full Name
	Blocks     []string          `json:"blocks"`
}

var (
	allCharacters []CharacterInfo
	categories    map[string]string // Map Abbreviation -> Full Name
	dataMutex     sync.RWMutex
)

// Map of Unicode categories with their full names
var categoryNames = map[string]string{
	"Lu": "Uppercase Letter",
	"Ll": "Lowercase Letter",
	"Lt": "Titlecase Letter",
	"Lm": "Modifier Letter",
	"Lo": "Other Letter",
	"Mn": "Nonspacing Mark",
	"Mc": "Spacing Mark",
	"Me": "Enclosing Mark",
	"Nd": "Decimal Number",
	"Nl": "Letter Number",
	"No": "Other Number",
	"Pc": "Connector Punctuation",
	"Pd": "Dash Punctuation",
	"Ps": "Open Punctuation",
	"Pe": "Close Punctuation",
	"Pi": "Initial Punctuation",
	"Pf": "Final Punctuation",
	"Po": "Other Punctuation",
	"Sm": "Math Symbol",
	"Sc": "Currency Symbol",
	"Sk": "Modifier Symbol",
	"So": "Other Symbol",
	"Zs": "Space Separator",
	"Zl": "Line Separator",
	"Zp": "Paragraph Separator",
	"Cc": "Control",
	"Cf": "Format",
	"Cs": "Surrogate",
	"Co": "Private Use",
	"Cn": "Unassigned",
}

// loadUnicodeData pre-populates the character list
func loadUnicodeData() {
	log.Println("Loading Unicode data...")
	dataMutex.Lock()
	defer dataMutex.Unlock()

	allCharacters = []CharacterInfo{}
	categories = make(map[string]string)
	addedCategories := make(map[string]bool)

	// Iterate through a relevant range (e.g., BMP 0x0000 to 0xFFFF)
	for r := rune(0); r <= 0xFFFF; r++ {
		if !unicode.IsPrint(r) || unicode.IsControl(r) || (unicode.IsSpace(r) && r != ' ') {
			// Skip non-printable, control chars (except space)
			// Add more exclusion logic if needed (e.g., surrogates)
			if r >= 0xD800 && r <= 0xDFFF { // Skip surrogate pairs
				continue
			}

			// Skip private use area for general browsing
			if r >= 0xE000 && r <= 0xF8FF {
				continue
			}

			// Skip combining marks unless you specifically want them displayed standalone
			// if unicode.In(r, unicode.Mn, unicode.Me, unicode.Mc) {
			// continue // Uncomment to skip combining marks
			// }

			if r == '\uFFFD' { // Skip replacement character
				continue
			}

			// Skip characters known to cause issues or be unrenderable in many contexts
			if r == 0xAD { // Soft hyphen
				continue
			}
			if r >= 0x2060 && r <= 0x206F { // General punctuation invisible operators
				continue
			}
			if r >= 0xFFF9 && r <= 0xFFFB { // Interlinear annotation anchors etc
				continue
			}
			continue // Default skip if not printable or otherwise undesirable
		}

		name := runenames.Name(r)
		if name == "" || strings.Contains(name, "<") { // Skip reserved/private use/control names
			continue
		}

		// Get character category
		catAb := getCategoryAbbreviation(r)
		catName := categoryNames[catAb]
		if catName == "" {
			catName = "Unknown Category"
		}

		info := CharacterInfo{
			Char:       string(r),
			CodePoint:  fmt.Sprintf("U+%04X", r),
			Name:       name,
			Category:   catName,
			CategoryAb: catAb,
		}
		allCharacters = append(allCharacters, info)

		// Collect unique categories
		if !addedCategories[catAb] {
			categories[catAb] = catName
			addedCategories[catAb] = true
		}
	}

	log.Printf("Loaded %d characters and %d categories.", len(allCharacters), len(categories))
}

// getCategoryAbbreviation returns the two-letter Unicode category for a rune
func getCategoryAbbreviation(r rune) string {
	// Check for specific category ranges
	switch {
	case unicode.IsLetter(r):
		if unicode.IsUpper(r) {
			return "Lu"
		} else if unicode.IsLower(r) {
			return "Ll"
		} else if unicode.IsTitle(r) {
			return "Lt"
		} else {
			// Further differentiate between Lm and Lo if needed
			return "Lo" // Default to "Other Letter"
		}
	case unicode.IsDigit(r):
		return "Nd"
	case unicode.IsPunct(r):
		// This is simplified - ideally you'd distinguish between different punctuation types
		return "Po"
	case unicode.IsSymbol(r):
		// This is simplified - ideally you'd distinguish between different symbol types
		if strings.Contains(runenames.Name(r), "CURRENCY") {
			return "Sc"
		}
		return "So"
	case unicode.IsSpace(r):
		return "Zs"
	case unicode.IsControl(r):
		return "Cc"
	default:
		return "Cn" // Unassigned as fallback
	}
}

// handleCharacters serves the character data based on query parameters
func handleCharacters(w http.ResponseWriter, r *http.Request) {
	dataMutex.RLock() // Use read lock for concurrent reads
	defer dataMutex.RUnlock()

	query := r.URL.Query()
	search := strings.ToLower(strings.TrimSpace(query.Get("search")))
	categoryFilter := query.Get("category") // Expecting Category Abbreviation (e.g., "Lu")

	pageStr := query.Get("page")
	limitStr := query.Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100 // Default limit
	}

	// Filter characters
	filtered := make([]CharacterInfo, 0, len(allCharacters))
	for _, charInfo := range allCharacters {
		match := true

		// Search filter (checks name, codepoint)
		if search != "" {
			nameLower := strings.ToLower(charInfo.Name)
			codeLower := strings.ToLower(charInfo.CodePoint)
			// Basic substring search, could be improved (e.g., word boundary)
			if !strings.Contains(nameLower, search) && !strings.Contains(codeLower, search) && charInfo.Char != search {
				match = false
			}
		}

		// Category filter
		if match && categoryFilter != "" && charInfo.CategoryAb != categoryFilter {
			match = false
		}

		if match {
			filtered = append(filtered, charInfo)
		}
	}

	// Apply pagination
	totalItems := len(filtered)
	totalPages := (totalItems + limit - 1) / limit
	if page > totalPages && totalPages > 0 {
		page = totalPages // Adjust page if it's out of bounds
	}

	start := (page - 1) * limit
	end := start + limit
	if start > totalItems {
		start = totalItems
	}
	if end > totalItems {
		end = totalItems
	}

	paginatedChars := []CharacterInfo{}
	if start < end { // Ensure indices are valid
		paginatedChars = filtered[start:end]
	}

	// Prepare response
	resp := APIResponse{
		Characters:   paginatedChars,
		TotalItems:   totalItems,
		CurrentPage:  page,
		ItemsPerPage: limit,
		TotalPages:   totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleMetadata serves category (and potentially block) lists
func handleMetadata(w http.ResponseWriter, r *http.Request) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	// Sort category names alphabetically for the dropdown
	sortedCategories := make(map[string]string)
	catKeys := make([]string, 0, len(categories))
	for k := range categories {
		catKeys = append(catKeys, k)
	}
	sort.Slice(catKeys, func(i, j int) bool {
		// Sort by full name for user-friendliness
		return categories[catKeys[i]] < categories[catKeys[j]]
	})
	for _, k := range catKeys {
		sortedCategories[k] = categories[k]
	}

	resp := MetadataResponse{
		Categories: sortedCategories,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding metadata JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// serveHTML serves the index.html file
func serveHTML(w http.ResponseWriter, r *http.Request) {
	// Basic security: Prevent path traversal
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "uniGo.html") // Assume index.html is in the same directory
}

func main() {
	// Load data once on startup
	loadUnicodeData()

	// --- HTTP Handlers ---
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/api/characters", handleCharacters)
	http.HandleFunc("/api/metadata", handleMetadata)

	// --- Start Server ---
	port := "6969"
	log.Printf("Starting server on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

## `frontend\src\App.svelte`

```svelte
<script>
  import { onMount } from "svelte";
  import { ListSvgIcons, ListNerdFontIcons } from "../wailsjs/go/main/App"; // Wails generated bindings
  import { optimize } from "svgo"; // Correct import for SVGO
  import { loadConfig } from 'svgo';

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

  // SVGO configuration (you can customize this further)
  const svgoConfig = {
    multipass: true, // Recommended for better optimization
    plugins: [
      // Using preset-default is a good starting point as it includes many common optimizations.
      {
        name: 'preset-default',
        params: {
          overrides: {
            // Example: disable a specific plugin from the preset if needed
            // removeViewBox: false,
            // Or enable/configure specific plugins if you don't use preset-default
          },
        },
      },
      // If you prefer to list plugins manually (more control, more verbose):
      'removeDoctype',
      // 'removeXMLProcInst',
      'removeComments',
      // 'removeMetadata',
      // 'removeEditorsNSData',
      // 'cleanupAttrs',
      // 'mergeStyles',
      // 'inlineStyles',
      'minifyStyles',
      'cleanupIDs', // SVGO uses 'cleanupIDs'
      // { name: 'convertShapeToPath', params: { convertArcs: true } }, // Fixed typo from 'Shapt'
      'cleanupNumericValues',
      // 'convertColors',
      'removeUnknownsAndDefaults',
      // 'removeNonInheritableGroupAttrs',
      'removeUselessStrokeAndFill',
      // 'removeViewBox', // Be careful, often desired
      // 'cleanupEnableBackground',
      'removeHiddenElems',
      // 'removeEmptyText',
      // 'convertPathData',
      // 'convertTransform',
      // 'removeEmptyAttrs',
      // 'removeEmptyContainers',
      // 'mergePaths',
      'removeUnusedNS',
      'sortAttrs'
      // 'sortDefsChildren',
      // 'removeTitle', // Often good to remove for icon libraries
      // 'removeDesc', // Often good to remove
    ],
  };

  /**
   * Optimizes a single SVG string using SVGO.
   * @param {string} svgString The raw SVG content.
   * @returns {string} The optimized SVG string, or the original if optimization fails.
   */
  function optimizeSvgContent(svgString) {
    if (!svgString || typeof svgString !== 'string') {
      // console.warn("Invalid SVG string passed to optimizer.");
      return svgString || ""; // Return original or empty if invalid
    }
    try {
      const result = optimize(svgString, svgoConfig);
      return result.data;
    } catch (error) {
      console.error("SVGO Optimization Error for a an icon:", error, "\nOriginal SVG:\n", svgString.substring(0, 200)+"...");
      return svgString; // Return original SVG string on error
    }
  }

  onMount(async () => {
    isLoading = true; // Ensure loading state is true at the start
    errorMsg = "";    // Clear previous errors
    try {
      const [svgsResponse, nfIconsResponse] = await Promise.all([
        ListSvgIcons(),
        ListNerdFontIcons(),
      ]);

      const rawSvgs = svgsResponse || [];
      // Optimize SVG content as it's loaded
      svgIcons = rawSvgs.map(icon => ({
        ...icon,
        content: optimizeSvgContent(icon.content), // Optimize here
      }));

      nerdFontIcons = nfIconsResponse || [];
      filterIcons(); // Initial filter after loading and optimizing
    } catch (err) {
      console.error("Error loading icons:", err);
      errorMsg = `Failed to load icons: ${err.message || err}`;
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
          (icon.codepoint && icon.codepoint.toLowerCase().includes(term)) // Ensure codepoint exists
      );
    } else {
      filteredNerdFontIcons = [];
    }
  }

  // Reactive statement to re-filter when search term or source data changes
  $: if (searchTerm || svgIcons.length || nerdFontIcons.length) {
    // This condition might cause filterIcons to run before svgIcons/nerdFontIcons are fully populated
    // It's generally okay because filterIcons handles potentially empty arrays.
    // Consider calling filterIcons explicitly after data mutations if needed.
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
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(text);
        displayToast(`${type} copied!`);
      } else {
         displayToast('Clipboard API not available in this context.');
      }
    } catch (err) {
      console.error("Failed to copy: ", err);
      displayToast(`Failed to copy ${type}. Check console.`);
    }
  }

  function handleSvgClick(icon) {
    copyToClipboard(icon.content, "Optimized SVG content"); // Now copies optimized SVG
  }

  function handleNerdFontClick(icon) {
    const character = String.fromCodePoint(parseInt(icon.codepoint, 16));
    copyToClipboard(character, `NerdFont char ${icon.name}`);
  }

  function getNerdFontCharacter(codepoint) {
    if (!codepoint) return '';
    return String.fromCodePoint(parseInt(codepoint, 16));
  }
</script>

<!-- Your HTML template remains the same -->
<div class="app-container">
  <aside class="sidebar drag">
    <h2>📖 uniGO</h2>
    <i style="margin-bottom: 1em;font-size:13px;color: #aaa;">Powered by <span class="nerd-font-icon-display" style="font-size: 13px;"></span></i>
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
          <span>SVG</span>
          <span class="icon-count-badge">{filteredSvgIcons.length}</span>
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
              title="Click to copy optimized SVG content: {icon.name}"
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

## `frontend\src\style.css`

```css
/* Consolidated :root and base styles */

* {
  scrollbar-width: none; /* hides scrollbar */
}

/* Chrome, Edge, Safari */
*::-webkit-scrollbar {
  width: 0;
  height: 0;
}

:root {
  /* Primary font for UI text - choose one or define fallbacks */
  --font-sans: "SF Pro Text", "Symbols Nerd Font", sans-serif; /* Added Nunito here */
  --font-mono: "SFMono Nerd Font", "Symbols Nerd Font", monospace;

  font-size: 16px;
  line-height: 1.2;
  font-weight: 400;

  color-scheme: light dark;
  color: rgba(255, 255, 255, 0.87);

  scroll-behavior: smooth;
  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  -webkit-text-size-adjust: 100%;

  --sidebar-width: 180px;
  --header-height: 60px; /* Not currently used, but good to keep if you add a header */
  --gap: 0.8rem;
  --icon-size: 3rem;
  --border-color: #00000000; /* Transparent border is fine */
}

/* Ensure html, body, and #app take full height and don't interfere with flex/grid */
html {
  height: 100vh;
  background-image: url("assets/images/fabric.png");
  /* background-color: rgba(25, 25, 25, 0.5); This will be under body's background */
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
  font-family: "Symbols Nerd Font";
  font-style: normal;
  font-weight: 400;
  src: local(""),
    url("frontend/src/assets/fonts/SymbolsNerdFont-Regular.ttf") format("ttf");
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
  padding: 0.5em;
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
  border: 0px solid var(--border-color);
  border-radius: 7px;
  padding: var(--gap);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  cursor: pointer;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
  height: 90px; /* Give cards a minimum height */
  width: 110px;
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

//go:embed build/appicon.png
var appIcon []byte

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
			Assets:  assets,
			Handler: iconLoader(),
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
				Icon:    appIcon,
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
			// BackdropType:                      windows.Acrylic,
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
