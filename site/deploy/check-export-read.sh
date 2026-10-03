#!/usr/bin/env bash
# Refuse a deploy before it happens unless wrangler reads every entry of the export.
#
# The first deploy uploaded the 77 top-level files and none of the pages beneath
# them: Yarn had fallen back to Plug'n'Play (this package's .yarnrc.yml was not in
# the sparse checkout yet), and PnP's patched fs ignores readdir's `recursive`.
# Nothing failed; pages answered 404. This compares wrangler's own count, from a
# dry run, with what is on disk, so any such gap stops the job instead.
#
#   check-export-read.sh <wrangler args identifying the Worker, e.g. --env="">
#
# Run from site/deploy, after the export is in ../out.
set -euo pipefail

want="$(find ../out -mindepth 1 | wc -l | tr -d ' ')"
read_line="$(yarn wrangler deploy --dry-run --config ../wrangler.jsonc "$@" 2>&1 | grep -E 'Read [0-9]+ files? from the assets directory' || true)"
read_count="$(grep -oE 'Read [0-9]+' <<<"$read_line" | grep -oE '[0-9]+' || echo 0)"

if [[ "$read_count" != "$want" ]]; then
  echo "::error::wrangler reads ${read_count} of the export's ${want} entries; refusing to deploy a partial site"
  exit 1
fi
echo "wrangler reads all ${want} entries of the export"
