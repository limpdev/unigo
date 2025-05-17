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