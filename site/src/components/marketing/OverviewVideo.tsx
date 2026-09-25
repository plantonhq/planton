import { OVERVIEW_VIDEO as video, OVERVIEW_CHAPTERS } from '@/data/homepage-video';
import { OverviewVideoPlayer } from './OverviewVideoPlayer';
import styles from './overview-video.module.css';

/** Server-rendered copy and transcript stay available without JavaScript. */
export function OverviewVideo() {
  return <section className={styles.section} aria-labelledby="overview-video-title">
    <div className={styles.heading}>
      <h2 id="overview-video-title">{video.label}</h2>
      <p id="overview-video-description">{video.description}</p>
    </div>
    <OverviewVideoPlayer />
    <div className={styles.meta}>
      <span>60 seconds · No sound needed</span>
      <a href={`${video.base}/${video.id}-1080p.mp4`}>Open Full-HD Video <span aria-hidden="true">↗</span></a>
    </div>
    <details className={styles.transcript}>
      <summary>Read the Video Transcript</summary>
      <ol>
        {OVERVIEW_CHAPTERS.map(chapter => <li key={chapter.start}>
          <span className={styles.time}>0:{String(chapter.start).padStart(2, '0')}</span>
          <div><h3>{chapter.title}</h3><p>{chapter.transcript}</p></div>
        </li>)}
      </ol>
    </details>
  </section>;
}
