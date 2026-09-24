import { CONTROL_COPY as C } from '@/data/homepage-experience';
import styles from './experience.module.css';
/** Readable examples distinguish team controls from platform placement. */
export function ControlOwnership() {
  return (
    <div className={styles.controlGrid}>
      <div className={styles.controlCard}>
        <p className={styles.kicker}>{C.label}</p>
        <ol className={styles.controlSteps}>
          {C.steps.map((step, i) => (
            <li key={step.title}>
              <span className={styles.step}>{i + 1}</span>
              <div>
                <h3>{step.title}</h3>
                <strong>{step.value}</strong>
                <p>{step.text}</p>
              </div>
            </li>
          ))}
        </ol>
      </div>
      <div className={styles.controlCard}>
        <p className={styles.kicker}>{C.placement}</p>
        <div className={styles.choices}>
          {C.choices.map((item) => (
            <div key={item.title}>
              <h3>{item.title}</h3>
              <p>{item.text}</p>
            </div>
          ))}
        </div>
        <div className={styles.cloud}>
          <strong>{C.boundary}</strong>
          <p>{C.boundaryText}</p>
        </div>
        <p className={styles.ownershipNote}>{C.note}</p>
      </div>
    </div>
  );
}
