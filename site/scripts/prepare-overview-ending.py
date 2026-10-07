"""Hold a logo frame, then fade to its background without changing earlier timing.

Requires a full FFmpeg build with trim, tpad, and fade filters. The minimal
Remotion binary intentionally omits those filters. Never overwrite the input.
"""
import argparse
import math
from pathlib import Path
import subprocess
import tempfile

parser = argparse.ArgumentParser()
parser.add_argument('--input', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
parser.add_argument('--ffmpeg', required=True, help='Path to a full FFmpeg binary')
parser.add_argument('--hold-at', type=float, default=44, help='End of original picture to retain, seconds')
parser.add_argument('--duration', type=float, default=47)
parser.add_argument('--fade', type=float, default=1)
parser.add_argument('--background', default='0xf0eeeb')
args = parser.parse_args()
if args.input.resolve() == args.output.resolve():
    parser.error('Output must not overwrite the original picture')
if not all(math.isfinite(v) for v in [args.hold_at, args.duration, args.fade]) or not 0 < args.hold_at < args.duration - args.fade < args.duration:
    parser.error('Require 0 < hold-at < duration - fade < duration')
args.output.parent.mkdir(parents=True, exist_ok=True)
filters = (
    f'tpad=stop_mode=clone:stop_duration={args.duration - args.hold_at},'
    f'fade=t=out:st={args.duration - args.fade}:d={args.fade}:color={args.background}'
)
# Finish the trim in a lossless intermediate first. A trim EOF inside the same
# graph can prevent tpad from emitting its ending on some FFmpeg versions.
with tempfile.TemporaryDirectory(prefix='planton-ending-') as temporary:
    trimmed = Path(temporary) / 'trimmed.mp4'
    subprocess.run([
        args.ffmpeg, '-y', '-v', 'error', '-i', str(args.input.resolve()),
        '-map', '0:v:0', '-an', '-t', str(args.hold_at), '-c:v', 'libx264',
        '-crf', '0', '-pix_fmt', 'yuv420p', '-r', '30', str(trimmed),
    ], check=True)
    subprocess.run([
        args.ffmpeg, '-y', '-v', 'error', '-i', str(trimmed),
        '-map', '0:v:0', '-vf', filters, '-an', '-c:v', 'libx264', '-crf', '16',
        '-pix_fmt', 'yuv420p', '-r', '30', '-t', str(args.duration),
        '-movflags', '+faststart', str(args.output.resolve()),
    ], check=True)
print(f'Prepared {args.duration:g}s picture: original timing to {args.hold_at:g}s, logo hold, {args.fade:g}s fade.')
