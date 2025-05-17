<script>
	import { onMount } from "svelte";
	import { ListSvgIcons, ListNerdFontIcons } from "../wailsjs/go/main/App"; // Wails generated bindings
	import { optimize } from "svgo"; // Correct import for SVGO
	import { loadConfig } from "svgo";

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
				name: "preset-default",
				params: {
					overrides: {
						// Example: disable a specific plugin from the preset if needed
						// removeViewBox: false,
						// Or enable/configure specific plugins if you don't use preset-default
					},
				},
			},
			// If you prefer to list plugins manually (more control, more verbose):
			"removeDoctype",
			// 'removeXMLProcInst',
			"removeComments",
			// 'removeMetadata',
			// 'removeEditorsNSData',
			// 'cleanupAttrs',
			// 'mergeStyles',
			// 'inlineStyles',
			"minifyStyles",
			"cleanupIDs", // SVGO uses 'cleanupIDs'
			// { name: 'convertShapeToPath', params: { convertArcs: true } }, // Fixed typo from 'Shapt'
			"cleanupNumericValues",
			// 'convertColors',
			"removeUnknownsAndDefaults",
			// 'removeNonInheritableGroupAttrs',
			"removeUselessStrokeAndFill",
			// 'removeViewBox', // Be careful, often desired
			// 'cleanupEnableBackground',
			"removeHiddenElems",
			// 'removeEmptyText',
			// 'convertPathData',
			// 'convertTransform',
			// 'removeEmptyAttrs',
			// 'removeEmptyContainers',
			// 'mergePaths',
			"removeUnusedNS",
			"sortAttrs",
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
		if (!svgString || typeof svgString !== "string") {
			// console.warn("Invalid SVG string passed to optimizer.");
			return svgString || ""; // Return original or empty if invalid
		}
		try {
			const result = optimize(svgString, svgoConfig);
			return result.data;
		} catch (error) {
			console.error("SVGO Optimization Error for a an icon:", error, "\nOriginal SVG:\n", svgString.substring(0, 200) + "...");
			return svgString; // Return original SVG string on error
		}
	}

	onMount(async () => {
		isLoading = true; // Ensure loading state is true at the start
		errorMsg = ""; // Clear previous errors
		try {
			const [svgsResponse, nfIconsResponse] = await Promise.all([ListSvgIcons(), ListNerdFontIcons()]);

			const rawSvgs = svgsResponse || [];
			// Optimize SVG content as it's loaded
			svgIcons = rawSvgs.map((icon) => ({
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
			filteredSvgIcons = svgIcons.filter((icon) => icon.name.toLowerCase().includes(term));
		} else {
			filteredSvgIcons = [];
		}

		if (nerdFontIcons) {
			filteredNerdFontIcons = nerdFontIcons.filter(
				(icon) => icon.name.toLowerCase().includes(term) || (icon.codepoint && icon.codepoint.toLowerCase().includes(term)), // Ensure codepoint exists
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
				displayToast("Clipboard API not available in this context.");
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
		if (!codepoint) return "";
		return String.fromCodePoint(parseInt(codepoint, 16));
	}
</script>

<!-- Your HTML template remains the same -->
<div class="app-container">
	<aside class="sidebar drag">
		<h2>📖 uniGO</h2>
		<i style="margin-bottom: 1em;font-size:13px;color: #aaa;">Powered by <span class="nerd-font-icon-display" style="font-size: 13px;"></span></i>
		<input type="text" class="search-bar" placeholder="Search icons..." bind:value={searchTerm} on:input={filterIcons} />
		<ul>
			<li>
				<button class:active={currentView === "svg"} on:click={() => setView("svg")}>
					<span>SVG</span>
					<span class="icon-count-badge">{filteredSvgIcons.length}</span>
				</button>
			</li>
			<li>
				<button class:active={currentView === "nerdfont"} on:click={() => setView("nerdfont")}>
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
					{searchTerm ? " matching your search" : svgIcons.length === 0 ? " in content/svgs/. Add some!" : ""}.
				</p>
			{:else}
				<div class="icon-grid">
					{#each filteredSvgIcons as icon (icon.path)}
						<!-- svelte-ignore a11y-click-events-have-key-events -->
						<div class="icon-card" aria-label="SVG icon: {icon.name}" on:click={() => handleSvgClick(icon)} title="Click to copy optimized SVG content: {icon.name}">
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
					{searchTerm ? " matching your search" : nerdFontIcons.length === 0 ? " in content/nerd-fonts/icons.json or font not loaded. Check console & CSS." : ""}.
				</p>
			{:else}
				<div class="icon-grid">
					{#each filteredNerdFontIcons as icon (icon.codepoint)}
						<!-- svelte-ignore a11y-click-events-have-key-events -->
						<div class="icon-card" aria-label="Nerd Font icon: {icon.name}" on:click={() => handleNerdFontClick(icon)} title="Click to copy character: {icon.name}">
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
