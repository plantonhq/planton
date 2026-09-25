'use client';

import { useRef, useState } from 'react';
import { OVERVIEW_VIDEO as video } from '@/data/homepage-video';
import { trackOverviewVideo } from '@/lib/demo-analytics';
import styles from './overview-video.module.css';

/** Native media owns playback, seeking, fullscreen, and keyboard behavior.
 * There is no observer, timer, hover handler, or custom player to hydrate. */
export function OverviewVideoPlayer() {
  const [failed, setFailed] = useState(false);
  const started = useRef(false), completed = useRef(false);
  const source = `${video.base}/${video.id}-720p.mp4`;
  return <>
    <video
      className={styles.player}
      id="homepage-overview-video"
      aria-label={video.label}
      aria-describedby="overview-video-description"
      width={1280}
      height={720}
      controls
      playsInline
      preload="none"
      poster={video.poster}
      src={source}
      onPlay={() => {
        setFailed(false);
        if (!started.current) {
          started.current = true;
          trackOverviewVideo('overview_video_start');
        }
      }}
      onEnded={() => {
        if (!completed.current) {
          completed.current = true;
          trackOverviewVideo('overview_video_complete');
        }
      }}
      onError={() => setFailed(true)}
    >
      <a href={source}>Watch the Planton overview video</a>
    </video>
    {failed && <p className={styles.error} role="alert">The video could not load. <a href={source}>Open the video directly</a>, or read the transcript below.</p>}
  </>;
}
