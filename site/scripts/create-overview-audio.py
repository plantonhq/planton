"""Local audio proof: macOS speech plus an original, low-level ambient score.

No service credentials or third-party music. This is deliberately an audio
preview; review the synthetic voice before promoting a new CDN video version.
The existing silent MP4 remains the authoritative picture master.
"""
import argparse
import array
import json
import math
import os
from pathlib import Path
import subprocess
import sys
import wave

parser = argparse.ArgumentParser()
parser.add_argument('--output', required=True)
parser.add_argument('--voice', default='Samantha')
args = parser.parse_args()
site = Path(__file__).resolve().parent.parent
output = Path(args.output).resolve()
output.mkdir(parents=True, exist_ok=True)
chapters = json.loads((site / 'video/overview-narration.json').read_text())
ffmpeg = next((site / 'node_modules/@remotion').glob('compositor-darwin-*/ffmpeg'))
env = {**os.environ, 'DYLD_LIBRARY_PATH': str(ffmpeg.parent)}

def run(*command):
    subprocess.run([str(c) for c in command], check=True, env=env, cwd=ffmpeg.parent)

rate = 44100
length = 60 * rate
voice = array.array('f', [0.0]) * length
subtitle = ['WEBVTT', '']
def timestamp(t):
    return f'00:{int(t)//60:02}:{t%60:06.3f}'

for i, chapter in enumerate(chapters):
    script = output / f'voice-{i+1}.txt'
    script.write_text(chapter['text'])
    decoded = output / f'voice-{i+1}.wav'
    available = chapter['end'] - chapter['start']
    # Adjust speaking rate, never cut a sentence or stretch the recorded voice.
    speaking_rate = 156
    for attempt in range(4):
        run('/usr/bin/say', '--file-format=WAVE', '--data-format=LEI16@44100', '-v', args.voice, '-r', speaking_rate, '-f', script, '-o', decoded)
        with wave.open(str(decoded)) as wav:
            pcm = array.array('h', wav.readframes(wav.getnframes()))
        duration = len(pcm) / rate
        if duration <= available:
            break
        speaking_rate = math.ceil(speaking_rate * duration / available) + 2
    if duration > available:
        raise RuntimeError(f'Chapter {i+1} overruns its scene')
    peak = max(abs(n) for n in pcm) or 1
    gain = 0.63 / peak
    start = round(chapter['start'] * rate)
    for j, sample in enumerate(pcm):
        voice[start+j] = sample * gain
    end = chapter['start'] + duration
    subtitle.extend([str(i+1), f"{timestamp(chapter['start'])} --> {timestamp(end)}", chapter['text'], ''])
    print(f'Chapter {i+1}: {duration:.2f}s / {available:.2f}s; {speaking_rate} words/minute', flush=True)

(output / 'overview-narration.vtt').write_text('\n'.join(subtitle))

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
    for chapter in chapters:
        if chapter['start']-.2 <= t <= chapter['end']+.25:
            duck = 0.48
            break
    fade = min(1,t/2) * min(1,max(0,(60-t)/2.5))
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
print(f'Mix peak: {20*math.log10(peak):.1f} dBFS; 60 seconds, stereo. Original music with speech ducking.')
