# Homepage overview films

## Current film: imported picture with Eric narration

The current homepage uses the supplied silent, 1080p, 30 fps MP4, with a
47-second finish. Original scene timing runs through 44 seconds; the logo then
holds until 46 seconds and fades to its warm background over the final second.
The closing narration ends near 43.9 seconds, leaving a quiet three-second landing.
Narration windows and transcript are maintained in
`src/data/homepage-video-story.json`; player metadata, duration, and versioned
poster/caption paths are in `src/data/homepage-video.ts`.

```sh
# Prepare the longer ending with a full FFmpeg binary (trim, tpad, fade).
python3 scripts/prepare-overview-ending.py \
  --input /path/to/planton-movie-1.mp4 --output /path/to/picture-47s.mp4 \
  --ffmpeg /path/to/ffmpeg --hold-at 44 --duration 47 --fade 1
# ELEVENLABS_API_KEY is read from the environment; --key-file is also supported.
python3 scripts/create-overview-audio.py \
  --picture-file /path/to/picture-47s.mp4 --duration 47 \
  --speed 0.82 --sentence-pause 0.35 --music-fade 3 \
  --output /path/to/planton-movie-1-narrated
node scripts/check-overview-export.mjs /path/to/planton-movie-1-narrated
OVERVIEW_VIDEO_FILE=/path/to/planton-movie-1-narrated/homepage-overview-720p.mp4 \
  node scripts/check-overview-video.mjs
```

Eric is the selected ElevenLabs voice (`eleven_multilingual_v2`), generated at
0.82 speaking speed with 0.35-second pauses between sentences. Chapters may
override `sentencePause`; the closing phrases use their natural pauses.
The shorter script leaves room for unhurried delivery within each scene.
Credentials
and generated media stay outside Git. Requests are cached by voice, text, and
settings. Overlong chapters stop generation for editorial correction; do not
speed up or truncate speech. `narration-timing.json` records actual chapter
lengths. Captions use actual narration start/end times.

The original ambient score smoothly ducks around actual speech, then fades
over the final three seconds. The mix is normalized to -16 LUFS with a -1.5 dBTP target.
Preparing the ending re-encodes the picture at CRF 16. Audio assembly copies
that prepared 1080p picture stream unchanged; the 720p delivery copy is scaled and encoded as H.264.
Both receive
stereo AAC audio and fast-start metadata.

Copy `overview-narration.vtt` into the versioned caption path in the player
metadata. Create a 1280x720 WebP poster from a representative frame and save it
to the matching versioned poster path. Keep previous posters and captions for
rollback. Review the entire film, narration, and pronunciation before release.

## Publish and verify

Upload the MP4s explicitly to:
`s3://planton-assets/videos/homepage/<version>/homepage-overview-{1080p,720p}.mp4`.
Use `Content-Type: video/mp4` and
`Cache-Control: public, max-age=31536000, immutable`.
Never use the deletion-managed `site/` asset mirror. Never overwrite a published
version: change both the version and base URL for different bytes. Verify CDN
byte-range GETs return 206 with Content-Range before releasing the website.

The native player uses the 720p copy, with a direct Full-HD link, no preload,
and no autoplay. Its transcript and direct links work without JavaScript.
Analytics record first play/completion once per page view, with ID and version.
The `release.site` workflow builds and deploys the website to Cloudflare Workers.
Confirm deployment success and live playback before reporting completion.

## Original 60-second Remotion composition

The original animation remains an offline export, independent of the current
homepage. Its frozen story and metadata live in `video/legacy-overview-story.json`
and `video/legacy-overview-data.ts`. Geometry remains in `overview-layout.ts`.

```sh
yarn export:overview --output=/tmp/planton-legacy-overview
node scripts/check-overview-export.mjs /tmp/planton-legacy-overview --legacy --silent
python3 scripts/create-overview-audio.py \
  --story video/legacy-overview-story.json --duration 60 \
  --picture-dir /tmp/planton-legacy-overview --output /tmp/planton-legacy-narrated
node scripts/check-overview-export.mjs /tmp/planton-legacy-narrated --legacy
```

Use `--stills` with the exporter for composition review. It produces 1080p and
720p picture masters, representative frames, a contact sheet, and a WebP poster.
These legacy exports do not change or publish the current homepage movie.
