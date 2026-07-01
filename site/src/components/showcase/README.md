# `components/showcase`

How Planton shows the product — the desktop app and the CLI as the *same thing,
two ways*.

- `WindowChrome` — shared macOS-style title bar (traffic lights + title).
- `Terminal` = `WindowChrome` + `TerminalLine[]`. Renders a real, read-only
  terminal from **structured data** (`TerminalLineData`), not a screenshot, so it
  stays crisp and editable.
- `AppFrame` — a desktop window frame. Shows a real screenshot when provided;
  until keynote captures exist it shows an honest placeholder (never a fabricated
  screenshot).
- `ShowcaseTabs` — the reusable two-tab **Desktop / Terminal** component
  (built on the Radix `ui/tabs` primitive), composing `AppFrame` + `Terminal`.

Terminal content is data: to change a demo, edit the `TerminalLineData[]`, not JSX.
