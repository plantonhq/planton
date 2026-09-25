# Homepage overview film

`HomepageOverview` is an offline Remotion composition, not a browser dependency.
Chapter timing and transcript live in `src/data/homepage-video.ts`; geometry lives
in `overview-layout.ts`. The film reuses the site's canonical icon registry,
dark palette, and length-sampled connector animation. The real architecture
export is an unretouched image with a bounded camera transform.

```sh
yarn export:overview --output=/tmp/planton-overview-review
node scripts/check-overview-export.mjs /tmp/planton-overview-review --silent
```

Use `--stills` for an early composition review. The exporter writes 1080p and
720p H.264 files, representative frames, a contact sheet, and an optimized WebP
poster. Review the entire film before release, including the approval wait and
both camera transitions. Keep generated MP4s outside Git. Copy only the poster
to `public/_site/images/product/homepage-overview.webp`.

Publish videos with explicit object uploads to:
`s3://planton-assets/videos/homepage/<version>/homepage-overview-{1080p,720p}.mp4`.
Use `Content-Type: video/mp4` and `Cache-Control: public, max-age=31536000, immutable`.
Never use the deletion-managed `site/` mirror for video. Never overwrite a
published version; bump `OVERVIEW_VIDEO.version` and its base URL for new bytes.
Verify the public CDN returns 206 and Content-Range for a byte-range GET before
releasing the player. The 720p file is the default website source; the 1080p
master is offered as a direct link.

The player uses native controls and no preload/autoplay. The HTML transcript and
direct MP4 links remain usable without JavaScript. Analytics count the first
play and first completion once per page view, including replays in that view;
only the public video ID and version are recorded.

## Narrated preview

Eric is the selected ElevenLabs voice. The narration script lives in
`src/data/homepage-video-story.json`; the same text drives the HTML transcript,
voice generation, and English captions. Its windows fit the scene boundaries.

```sh
python3 scripts/create-overview-audio.py --key-file /path/to/private-key \
  --output /tmp/planton-eric-review --picture-dir /tmp/planton-overview-review
```

The key is read into memory and sent only to ElevenLabs. Generated chapters are
cached by voice, script, and settings, avoiding duplicate charges on reruns.
Overlong chapters stop the process for editorial correction rather than being
sped up or truncated. The original ambient score ducks under speech. The final
mix is loudness-normalized and added to both picture masters without re-encoding
their video streams. Copy the generated WebVTT file to `public/_site/videos/homepage-overview-en.vtt`.
Run `check-overview-export.mjs` without `--silent` on the narrated deliverables. Review narration
and pronunciation before publishing a new immutable CDN version.
