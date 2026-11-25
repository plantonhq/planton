/** The hero diagram's clock: four seconds a phase, then a hold on the finished
 * state. Scenes sample time from here and never read wall time or randomness,
 * so every frame is reproducible. */
export const PHASE_SECONDS = 4;
export const HOLD_SECONDS = 4;
export const FPS = 30;
export const duration = (story: { phases: readonly unknown[] }) =>
  story.phases.length * PHASE_SECONDS + HOLD_SECONDS;
