'use client';

/**
 * The visitor's cloud, chosen once for the whole of Home. The first section
 * that shows something different on each cloud carries the picker; every
 * section below it reads the same choice through `useCloud`, so nobody picks
 * twice. The choice is remembered in this browser only (a convenience: the
 * page renders the default when storage is empty or refused).
 *
 * The clouds offered are the ones a section has a real artifact for (each a
 * real run on that cloud's demo company); a cloud joins when its run exists.
 */
import { createContext, useContext, useSyncExternalStore, type ReactNode } from 'react';
import { CLOUDS, DEFAULT_CLOUD, type CloudId } from '@/data/clouds';
import styles from './homepage.module.css';

const STORAGE_KEY = 'planton.home.cloud';
const isCloud = (value: string | null): value is CloudId => CLOUDS.some((c) => c.id === value);

/**
 * The choice as a small external store: this page's pick (kept in memory, so it
 * applies even when storage is refused), else this browser's remembered pick,
 * else the default. The server and the first paint render the default.
 */
let picked: CloudId | null = null;
const listeners = new Set<() => void>();
const subscribe = (listener: () => void) => {
  listeners.add(listener);
  return () => listeners.delete(listener);
};
const snapshot = (): CloudId => {
  if (picked) return picked;
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (isCloud(saved)) return saved;
  } catch {
    // Storage refused (a private window, blocked site data): the default stands.
  }
  return DEFAULT_CLOUD;
};
const choose = (next: CloudId) => {
  picked = next;
  try {
    localStorage.setItem(STORAGE_KEY, next);
  } catch {
    // Remembering is a convenience; the choice still applies on this page.
  }
  listeners.forEach((listener) => listener());
};

const CloudContext = createContext<{ cloud: CloudId; choose: (cloud: CloudId) => void }>({
  cloud: DEFAULT_CLOUD,
  choose,
});

export function CloudChoiceProvider({ children }: { children: ReactNode }) {
  const cloud = useSyncExternalStore(subscribe, snapshot, () => DEFAULT_CLOUD);
  return <CloudContext.Provider value={{ cloud, choose }}>{children}</CloudContext.Provider>;
}

export const useCloud = () => useContext(CloudContext);

/** The picker: one radio per cloud, drawn as a row of segments. */
export function CloudPicker({ label }: { label: string }) {
  const { cloud, choose } = useCloud();
  return (
    <fieldset className={styles.cloudPicker}>
      <legend>{label}</legend>
      <div>
        {CLOUDS.map((c) => (
          <label key={c.id} data-checked={c.id === cloud || undefined}>
            <input
              type="radio"
              name="home-cloud"
              value={c.id}
              checked={c.id === cloud}
              onChange={() => choose(c.id)}
            />
            {c.name}
          </label>
        ))}
      </div>
    </fieldset>
  );
}
