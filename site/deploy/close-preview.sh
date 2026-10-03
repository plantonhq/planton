#!/usr/bin/env bash
# Close one website preview: upload a one-page version under the preview's alias, so
# review-<n>.<preview zone> (and the version address behind it) says the preview closed
# instead of serving the pull request's last build. Cloudflare keeps an alias until 1000
# newer ones push it out and offers no way to delete one, so overwriting it is how a
# preview is destroyed. Idempotent: closing a closed preview uploads the same page again.
#
#   close-preview.sh <pull request number>
#
# Run from site/deploy with CLOUDFLARE_API_TOKEN set to the preview key (it reaches the
# preview Worker alone).
set -euo pipefail

number="${1:?usage: close-preview.sh <pull request number>}"
[[ "$number" =~ ^[1-9][0-9]*$ ]] || { echo "not a pull request number: $number" >&2; exit 2; }

page="$(mktemp -d)"
trap 'rm -rf "$page"' EXIT

cat >"$page/index.html" <<HTML
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex">
<title>Preview closed · Planton</title>
<style>
:root { color-scheme: light dark; }
body { margin: 0; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 24px 16px; font: 15px/1.55 ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif; }
main { max-width: 480px; }
h1 { font-size: 22px; font-weight: 600; margin: 0 0 8px; }
p { margin: 0; opacity: 0.75; }
</style>
</head>
<body>
<main>
<h1>This preview has closed</h1>
<p>It showed pull request #${number}, which is merged or closed. The website is at <a href="https://planton.ai">planton.ai</a>.</p>
</main>
</body>
</html>
HTML
cp "$page/index.html" "$page/404.html"
# Every path answers the page; the marker tells the weekly sweep this preview is already closed.
printf '/*\n  X-Robots-Tag: noindex\n  X-Planton-Preview: closed\n' >"$page/_headers"

yarn wrangler versions upload --config ../wrangler.jsonc --env preview \
  --assets "$page" --preview-alias "review-${number}" --message "closed: pull request #${number}"
