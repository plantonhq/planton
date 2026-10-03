// Command catalog-logo-sheet renders one provider's kind logos as a
// contact sheet -- every registered kind's logo.svg at the sizes the console
// actually draws an icon at, on a light and a dark paper wash, the kind folder
// under each -- so a logo set can be judged glyph by glyph before it is kept.
// It is the visual half of the pkg/cataloglogo gate: the gate proves no two
// kinds share a drawing and every file names its provenance; this sheet is where a
// person decides whether each glyph reads as its concept, distinct from its
// siblings, at the smallest size the product renders it. It is a REPORT, not a
// gate: run it from the repository root any time
// (`go run ./tools/catalog-logo-sheet/ -provider cloudflare`), open the HTML
// it writes in a browser, and judge.
//
// Why these sizes: the console's diagram card sets its icon on a 34px plinth,
// the compact drawn forms (the worker chip, the globe) carry it at 26px, the
// vault door at 22px, and an attachment plate at 18px -- so 18px is the size
// that decides, and 48px is kept for looking at detail. Why an <img>: the
// console renders a logo as an image element from a CDN URL, never inline,
// so a stylesheet cannot reach into it; the sheet renders it the same way, and
// a glyph that leans on `currentColor` or an outer stylesheet fails here as it
// would in the product. Why a neutral wash by default: the console's plinth is
// the kind's family color, which only the platform knows, so the sheet's
// built-in washes are neutral and the family-colored judgment belongs to the
// platform's specimen sheet. When a kind has no card yet (a logo set judged
// before its provider's diagram identities exist), the platform supplies its
// family washes through -washes instead -- the color knowledge stays on the
// platform's side; this tool only knows how to paint what it is handed.
package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
)

const perRow = 9

// wash is one background the glyphs are judged on: the paper the console's
// plinth stands on in one theme, with a tint over it -- neutral by default,
// or a family accent at the plinth's alpha when the platform supplies one.
type wash struct {
	name  string
	paper string
	tint  string
	label string
}

var defaultWashes = []wash{
	{name: "light", paper: "#ffffff", tint: "rgba(96, 96, 96, 0.10)", label: "#444"},
	{name: "dark", paper: "#121212", tint: "rgba(200, 200, 200, 0.16)", label: "#bbb"},
}

// washSyntax is what -washes accepts: washes separated by ";", each
// "<name>=<paper>/<tint>". A CSS color may contain commas and parentheses
// (rgba), so the separators are ones no color uses.
const washSyntax = `-washes takes "<name>=<paper>/<tint>" entries separated by ";", e.g. "light=#ffffff/rgba(96,96,96,.10);dark=#121212/rgba(200,200,200,.16)"`

type entry struct {
	kindDir string
	dataURI string
	missing bool
}

func main() {
	provider := flag.String("provider", "", "provider whose logo set to render (catalog directory or enum name, e.g. cloudflare, digital_ocean)")
	out := flag.String("out", "", "HTML file to write (default /tmp/catalog-logo-sheet-<provider>.html)")
	sizes := flag.String("sizes", "18,26,34,48", "comma-separated pixel sizes to render each glyph at")
	washSpec := flag.String("washes", "", "backgrounds to judge on instead of the neutral light and dark pair; "+washSyntax)
	flag.Parse()

	if *provider == "" {
		fail("-provider is required: the catalog directory name of one provider (e.g. cloudflare)")
	}
	prov := catalogkindreflect.ProviderFromString(*provider)
	if prov == catalogkind.CatalogProvider_catalog_provider_unspecified {
		fail(fmt.Sprintf("unknown provider %q -- use a catalog directory name such as %s", *provider, strings.Join(providerDirNames(), ", ")))
	}
	providerDir := catalogkindreflect.ProviderDirName(prov)
	if _, err := os.Stat(filepath.Join("catalog", providerDir)); err != nil {
		fail(fmt.Sprintf("catalog/%s not found -- run this from the repository root", providerDir))
	}
	px, err := parseSizes(*sizes)
	if err != nil {
		fail(err.Error())
	}
	washes := defaultWashes
	if *washSpec != "" {
		if washes, err = parseWashes(*washSpec); err != nil {
			fail(err.Error())
		}
	}
	if *out == "" {
		*out = filepath.Join(os.TempDir(), "catalog-logo-sheet-"+providerDir+".html")
	}

	entries := collect(prov, providerDir)
	if len(entries) == 0 {
		fail(fmt.Sprintf("no registered kinds carry provider %s", providerDir))
	}
	if err := os.WriteFile(*out, []byte(render(providerDir, entries, px, washes)), 0o644); err != nil {
		fail(err.Error())
	}

	missing := 0
	for _, e := range entries {
		if e.missing {
			missing++
		}
	}
	names := make([]string, 0, len(washes))
	for _, w := range washes {
		names = append(names, w.name)
	}
	fmt.Printf("%s: %d kinds on the sheet (%d without a logo file) at %s px on %s -> %s\n", providerDir, len(entries), missing, *sizes, strings.Join(names, ", "), *out)
}

// collect walks the kind registry the way the gate does, so the sheet shows
// exactly the kinds the product serves for this provider.
func collect(prov catalogkind.CatalogProvider, providerDir string) []entry {
	metaByKind := catalogkindreflect.KindToKindMetaMap()
	var entries []entry
	for _, kind := range catalogkindreflect.KindsList() {
		if kind == catalogkind.CatalogKind_unspecified {
			continue
		}
		meta := metaByKind[kind]
		if meta == nil || meta.GetProvider() != prov {
			continue
		}
		kindDir := strings.ToLower(kind.String())
		data, err := os.ReadFile(filepath.Join("catalog", providerDir, kindDir, "logo.svg"))
		if err != nil {
			entries = append(entries, entry{kindDir: kindDir, missing: true})
			continue
		}
		entries = append(entries, entry{
			kindDir: kindDir,
			dataURI: "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(data),
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].kindDir < entries[j].kindDir })
	return entries
}

// render writes one section per (size, wash): a grid of plinth tiles with the
// glyph centered on each and the kind folder beneath, nine to a row.
func render(providerDir string, entries []entry, px []int, washes []wash) string {
	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>")
	b.WriteString(html.EscapeString(providerDir))
	b.WriteString(" logo sheet</title><style>")
	b.WriteString("body{margin:0;font-family:ui-monospace,Menlo,monospace;font-size:10px}")
	b.WriteString("section{padding:20px 24px}")
	b.WriteString("h2{font:600 13px/1 system-ui,sans-serif;margin:0 0 14px}")
	// minmax(0,1fr) and min-width:0 let a cell shrink below its long kind name,
	// so a nine-column row never overflows the page and drops its last column.
	b.WriteString(".grid{display:grid;grid-template-columns:repeat(" + strconv.Itoa(perRow) + ",minmax(0,1fr));gap:14px 10px}")
	b.WriteString(".cell{display:flex;flex-direction:column;align-items:center;gap:6px;min-width:0}")
	b.WriteString(".tile{display:flex;align-items:center;justify-content:center;border-radius:10px;border:1px solid rgba(128,128,128,.35)}")
	b.WriteString(".name{max-width:100%;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;opacity:.85}")
	b.WriteString(".missing{font:600 9px system-ui,sans-serif;color:#c33}")
	b.WriteString("</style></head><body>")
	for _, size := range px {
		for _, w := range washes {
			tile := size + 14 // the console's plinth padding around its icon
			fmt.Fprintf(&b, "<section style=\"background:%s;color:%s\"><h2>%s -- %dpx on %s</h2><div class=\"grid\">", w.paper, w.label, html.EscapeString(providerDir), size, w.name)
			for _, e := range entries {
				b.WriteString("<div class=\"cell\">")
				fmt.Fprintf(&b, "<div class=\"tile\" style=\"width:%dpx;height:%dpx;background:%s\">", tile, tile, w.tint)
				if e.missing {
					b.WriteString("<span class=\"missing\">none</span>")
				} else {
					fmt.Fprintf(&b, "<img src=\"%s\" width=\"%d\" height=\"%d\" alt=\"\">", e.dataURI, size, size)
				}
				b.WriteString("</div>")
				fmt.Fprintf(&b, "<div class=\"name\" title=\"%s\">%s</div>", html.EscapeString(e.kindDir), html.EscapeString(strings.TrimPrefix(e.kindDir, providerDir)))
				b.WriteString("</div>")
			}
			b.WriteString("</div></section>")
		}
	}
	b.WriteString("</body></html>")
	return b.String()
}

func parseSizes(s string) ([]int, error) {
	var px []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("-sizes must be positive pixel counts separated by commas, got %q", part)
		}
		px = append(px, n)
	}
	if len(px) == 0 {
		return nil, fmt.Errorf("-sizes names no size")
	}
	return px, nil
}

// parseWashes reads the -washes flag (washSyntax). The label color is chosen
// from the paper: dark text on a light paper, light text on a dark one, so a
// caller names only the two things the console's plinth is made of.
func parseWashes(spec string) ([]wash, error) {
	var ws []wash
	for _, part := range strings.Split(spec, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, colors, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("wash %q has no \"=\"; %s", part, washSyntax)
		}
		paper, tint, ok := strings.Cut(colors, "/")
		name, paper, tint = strings.TrimSpace(name), strings.TrimSpace(paper), strings.TrimSpace(tint)
		if !ok || name == "" || paper == "" || tint == "" {
			return nil, fmt.Errorf("wash %q must be <name>=<paper>/<tint>; %s", part, washSyntax)
		}
		ws = append(ws, wash{name: name, paper: paper, tint: tint, label: labelFor(paper)})
	}
	if len(ws) == 0 {
		return nil, fmt.Errorf("-washes names no wash; %s", washSyntax)
	}
	return ws, nil
}

// labelFor picks the caption color that reads on a paper given as a six-digit
// hex color; anything else (a named or functional color) gets the dark label,
// since the console's papers are hex.
func labelFor(paper string) string {
	hex := strings.TrimPrefix(paper, "#")
	if len(hex) != 6 {
		return "#444"
	}
	rgb, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return "#444"
	}
	r, g, b := float64(rgb>>16&0xff), float64(rgb>>8&0xff), float64(rgb&0xff)
	if 0.2126*r+0.7152*g+0.0722*b < 128 {
		return "#bbb"
	}
	return "#444"
}

func providerDirNames() []string {
	var names []string
	for _, p := range catalogkindreflect.ProvidersList() {
		if p == catalogkind.CatalogProvider_catalog_provider_unspecified {
			continue
		}
		names = append(names, catalogkindreflect.ProviderDirName(p))
	}
	sort.Strings(names)
	return names
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "catalog-logo-sheet: "+msg)
	os.Exit(1)
}
