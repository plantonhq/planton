"""ElevenLabs narration with an original, low-level ambient score.

Keep credentials and generated media outside Git. Cache each completed request
so a timing correction never regenerates or bills unchanged chapters. The silent
MP4 remains the picture master; this script does not publish anything.
"""
import argparse
import array
import json
import math
import hashlib
import os
from pathlib import Path
import subprocess
import sys
import wave
import urllib.request
import urllib.error
from xml.sax.saxutils import escape

parser = argparse.ArgumentParser()
parser.add_argument('--output', required=True)
parser.add_argument('--key-file', help='Optional private key file; otherwise use ELEVENLABS_API_KEY')
parser.add_argument('--story', type=Path, help='Chapter JSON; defaults to the current homepage story')
parser.add_argument('--duration', type=float, help='Exact picture duration; defaults to the final chapter end')
parser.add_argument('--picture-file', type=Path, help='Imported silent 1080p H.264 MP4; produces both delivery sizes')
parser.add_argument('--speed', type=float, default=1.0, help='ElevenLabs speaking speed (0.7 to 1.2); never time-stretch the output')
parser.add_argument('--music-fade', type=float, default=2.5, help='Final music fade duration in seconds')
parser.add_argument('--sentence-pause', type=float, default=0, help='Explicit pause between sentences in seconds (0 to 3)')
parser.add_argument('--voice-id', default='cjVigY5qzO86Huf0OWal', help='Eric, the selected narration voice')
parser.add_argument('--picture-dir', help='Optional directory containing the silent 1080p and 720p masters')
args = parser.parse_args()
site = Path(__file__).resolve().parent.parent
output = Path(args.output).resolve()
output.mkdir(parents=True, exist_ok=True)
story = json.loads((args.story or site / 'src/data/homepage-video-story.json').read_text())
duration_seconds = args.duration if args.duration is not None else max(c['end'] for c in story)
if not math.isfinite(duration_seconds) or duration_seconds <= 0:
    parser.error('Duration must be positive and finite')
if not math.isfinite(args.music_fade) or not 0 < args.music_fade <= duration_seconds:
    parser.error('Music fade must be positive and no longer than the film')
if args.picture_file and args.picture_dir:
    parser.error('Use only one of --picture-file and --picture-dir')
if not 0.7 <= args.speed <= 1.2 or not 0 <= args.sentence_pause <= 3:
    parser.error('Speed must be 0.7–1.2 and sentence pause 0–3 seconds')
previous_end = 0
for chapter in story:
    if not (previous_end <= chapter['voiceStart'] < chapter['voiceEnd'] <= duration_seconds):
        parser.error('Narration windows must be ordered, non-overlapping, and within the picture')
    previous_end = chapter['voiceEnd']
chapters = [dict(start=c['voiceStart'], end=c['voiceEnd'], text=c['transcript'], pause=c.get('sentencePause', args.sentence_pause)) for c in story]
if any(not 0 <= c['pause'] <= 3 for c in chapters):
    parser.error('Chapter sentence pauses must be 0–3 seconds')
key = Path(args.key_file).expanduser().read_text().strip() if args.key_file else os.environ.get('ELEVENLABS_API_KEY', '').strip()
if not key:
    parser.error('Set ELEVENLABS_API_KEY or supply --key-file')
if key.startswith('ELEVENLABS_API_KEY='):
    key = key.split('=', 1)[1].strip().strip('\"\'')
ffmpeg = next((site / 'node_modules/@remotion').glob('compositor-*/ffmpeg'))
ffprobe = ffmpeg.with_name('ffprobe')
env = {**os.environ, 'DYLD_LIBRARY_PATH': str(ffmpeg.parent)}

def run(*command):
    subprocess.run([str(c) for c in command], check=True, env=env, cwd=ffmpeg.parent)

if args.picture_file:
    args.picture_file = args.picture_file.expanduser().resolve()
    result = subprocess.run([str(ffprobe), '-v', 'error', '-show_format', '-show_streams', '-of', 'json', str(args.picture_file)], check=True, env=env, cwd=ffmpeg.parent, capture_output=True, text=True)
    info = json.loads(result.stdout)
    picture = next(s for s in info['streams'] if s['codec_type'] == 'video')
    if (picture['codec_name'], picture['width'], picture['height'], picture['pix_fmt'], picture['avg_frame_rate']) != ('h264', 1920, 1080, 'yuv420p', '30/1'):
        parser.error('Imported picture must be 1920x1080 H.264 yuv420p at 30 fps')
    if abs(float(info['format']['duration']) - duration_seconds) > .05:
        parser.error('Imported picture duration differs from the requested narration duration')
    if any(s['codec_type'] == 'audio' for s in info['streams']):
        parser.error('Imported picture already has audio; supply a silent picture master')

rate = 44100
length = round(duration_seconds * rate)
voice = array.array('f', [0.0]) * length
subtitle = ['WEBVTT', '']
measurements = []
def timestamp(t):
    milliseconds = round(t * 1000)
    hours, milliseconds = divmod(milliseconds, 3600000)
    minutes, milliseconds = divmod(milliseconds, 60000)
    return f'{hours:02}:{minutes:02}:{milliseconds/1000:06.3f}'

for i, chapter in enumerate(chapters):
    script = output / f'voice-{i+1}.txt'
    script.write_text(chapter['text'])
    decoded = output / f'voice-{i+1}.wav'
    available = chapter['end'] - chapter['start']
    speech_text = chapter['text']
    if chapter['pause']:
        speech_text = escape(speech_text).replace('. ', f'. <break time="{chapter["pause"]}s" /> ')
    request = {
        'text': speech_text, 'model_id': 'eleven_multilingual_v2',
        'voice_settings': {'stability': 0.5, 'similarity_boost': 0.75, 'style': 0.15, 'use_speaker_boost': True},
        'seed': 42,
    }
    if args.speed != 1:
        request['voice_settings']['speed'] = args.speed
    fingerprint = hashlib.sha256(json.dumps([args.voice_id, request], sort_keys=True).encode()).hexdigest()[:16]
    encoded = output / f'voice-{i+1}-{fingerprint}.mp3'
    if not encoded.exists():
        url = f'https://api.elevenlabs.io/v1/text-to-speech/{args.voice_id}?output_format=mp3_44100_128'
        req = urllib.request.Request(url, data=json.dumps(request).encode(), headers={'xi-api-key': key, 'Content-Type': 'application/json'}, method='POST')
        try:
            with urllib.request.urlopen(req, timeout=90) as response:
                audio = response.read()
            temporary = encoded.with_suffix('.partial')
            temporary.write_bytes(audio)
            temporary.replace(encoded)
        except urllib.error.HTTPError as error:
            raise RuntimeError(f'ElevenLabs HTTP {error.code}: {error.read().decode().replace(key, "[redacted]")}') from None
    run(ffmpeg, '-y', '-v', 'error', '-i', encoded, '-ar', rate, '-ac', 1, '-c:a', 'pcm_s16le', decoded)
    with wave.open(str(decoded)) as wav:
        pcm = array.array('h', wav.readframes(wav.getnframes()))
    duration = len(pcm) / rate
    if duration > available:
        raise RuntimeError(f'Chapter {i+1}: {duration:.2f}s exceeds {available:.2f}s. Revise the script; do not rush or truncate the voice.')
    peak = max(abs(n) for n in pcm) or 1
    gain = 0.63 / peak
    start = round(chapter['start'] * rate)
    for j, sample in enumerate(pcm):
        voice[start+j] = sample * gain
    end = chapter['start'] + duration
    subtitle.extend([str(i+1), f"{timestamp(chapter['start'])} --> {timestamp(end)}", chapter['text'], ''])
    measurements.append(dict(chapter=i+1, start=chapter['start'], end=end, available=available, text=chapter['text'], cache=encoded.name))
    print(f'Chapter {i+1}: {duration:.2f}s / {available:.2f}s', flush=True)

(output / 'overview-narration.vtt').write_text('\n'.join(subtitle))
(output / 'narration-timing.json').write_text(json.dumps(measurements, indent=2) + '\n')

# An original C-major / A-minor ambient progression. Slow pad attacks, widely
# spaced soft notes, no percussion. Music ducks beneath each spoken chapter.
chords = [(48,55,59,64), (45,52,55,60), (41,48,52,57), (43,50,55,59)]
def hz(note): return 440 * 2 ** ((note-69)/12)
frequencies = [[hz(n) for n in chord] for chord in chords]
stereo = array.array('h')
music_only = array.array('h')
peak = 0.0
for n in range(length):
    t = n / rate
    segment = int(t // 8)
    local = t % 8
    chord = frequencies[segment % 4]
    pad = sum(math.sin(2*math.pi*f*t) + 0.14*math.sin(2*math.pi*f*2*t) for f in chord) / 4
    envelope = min(1, local/1.3) * min(1, (8-local)/1.8)
    note_time = t % 2
    note = chord[int(t//2) % 4] * 2
    bell = math.sin(2*math.pi*note*note_time) * math.exp(-note_time*2.9) * min(1,note_time/.02)
    duck = 1.0
    for chapter in measurements:
        # Smooth attack/release around actual speech, including the final line.
        envelope_voice = min(1, max(0, (t - chapter['start'] + .2) / .2),
                             max(0, (chapter['end'] + .6 - t) / .6))
        duck = min(duck, 1 - .52 * envelope_voice)
    fade = min(1,t/2) * min(1,max(0,(duration_seconds-t)/args.music_fade))
    bed = (0.038*pad*envelope + 0.017*bell) * duck * fade
    pan = 0.12 * math.sin(2*math.pi*.07*t)
    for side in (1-pan,1+pan):
        music = bed * side
        value = voice[n] + music
        peak = max(peak,abs(value))
        stereo.append(round(max(-1,min(1,value))*32767))
        music_only.append(round(music*32767))

for name, samples in [('overview-mix.wav',stereo),('overview-music.wav',music_only)]:
    if sys.byteorder != 'little': samples.byteswap()
    with wave.open(str(output/name),'wb') as wav:
        wav.setnchannels(2)
        wav.setsampwidth(2)
        wav.setframerate(rate)
        wav.writeframes(samples.tobytes())
print(f'Mix peak: {20*math.log10(peak):.1f} dBFS; {duration_seconds:g} seconds, stereo. Original music with speech ducking.')

if args.picture_dir or args.picture_file:
    # Normalize once so both sizes have identical narration and music. Copy the
    # picture stream unchanged; audio revisions do not require a Remotion render.
    mastered = output / 'overview-master.wav'
    run(ffmpeg, '-y', '-v', 'error', '-i', output / 'overview-mix.wav',
        '-af', 'loudnorm=I=-16:TP=-1.5:LRA=7', '-ar', rate, mastered)
    for size in ('1080p', '720p'):
        name = f'homepage-overview-{size}.mp4'
        picture = args.picture_file or Path(args.picture_dir).resolve() / name
        encoding = ['-c:v', 'copy']
        if args.picture_file and size == '720p':
            encoding = ['-vf', 'scale=1280:720', '-c:v', 'libx264', '-crf', '21', '-pix_fmt', 'yuv420p']
        run(ffmpeg, '-y', '-v', 'error', '-i', picture,
            '-i', mastered, '-map', '0:v:0', '-map', '1:a:0', *encoding,
            '-c:a', 'aac', '-b:a', '192k', '-t', duration_seconds, '-movflags', '+faststart', output / name)
        print(f'Exported {name}', flush=True)
