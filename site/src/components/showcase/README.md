# `components/showcase`

How Planton shows the product — the desktop app and the CLI as the *same thing,
two ways*.

- `WindowChrome` — shared macOS-style title bar (traffic lights + title).
- `Terminal` = `WindowChrome` + `TerminalLine[]`. Renders a real, read-only
  terminal from **structured data** (`TerminalLineData`), not a screenshot, so it
  stays crisp and editable.
- `AppFrame` — a desktop window frame. Priority: a real `screenshot` if given,
  else rendered `children` (honest rendered content — e.g. the architecture
  graph), else a labeled placeholder. Never a fabricated screenshot.
- `ArchitectureGraph` — the read-only resource graph body (no window chrome);
  drop it inside an `AppFrame` (which supplies the window). Rendered from data.
- `ShowcaseTabs` — the reusable two-tab **Desktop / Terminal** component
  (built on the Radix `ui/tabs` primitive), composing `AppFrame` + `Terminal`.
  The desktop tab takes a `screenshot`, or `media` (rendered content), or falls
  back to a placeholder.

Terminal content is data: to change a demo, edit the `TerminalLineData[]`, not JSX.
