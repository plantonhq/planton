import { OVERVIEW_VIDEO as video } from '@/data/homepage-video';
import { OverviewVideoPlayer } from './OverviewVideoPlayer';
import styles from './overview-video.module.css';

export function OverviewVideo() {
  return <section className={styles.section} aria-labelledby="overview-video-title">
    <div className={styles.heading}>
      <h2 id="overview-video-title">{video.label}</h2>
      <p id="overview-video-description">{video.description}</p>
    </div>
    <OverviewVideoPlayer />
  </section>;
}
