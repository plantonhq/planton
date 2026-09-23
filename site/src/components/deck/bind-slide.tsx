/**
 * A slide is a component with no props, because a meeting deck hand-writes
 * every slide. A deck rendered from data (a persona deck) uses one generic
 * slide component many times with different content, so it binds the content
 * to the component here and hands the engine an ordinary SlideConfig. The
 * bound component is created once, when the deck's slide list is built at
 * module scope, so its identity is stable across renders and the engine's
 * cross-fade never remounts a slide because the presenter toggled the notes.
 *
 * `props` is typed against the component, so a slide cannot be bound to
 * content it does not accept; `meta` is what the engine needs to know about
 * the slide (its hash id, the name the dots show, the notes).
 */
import type { ComponentType } from 'react';
import type { SlideComponentProps, SlideConfig } from './Deck';

export function bindSlide<P extends object>(
  component: ComponentType<P & SlideComponentProps>,
  props: P,
  meta: Omit<SlideConfig, 'component'>,
): SlideConfig {
  const Slide = component;
  const Bound: ComponentType<SlideComponentProps> = ({ notesVisible }) => <Slide {...props} notesVisible={notesVisible} />;
  Bound.displayName = `Bound(${Slide.displayName ?? Slide.name ?? 'Slide'})`;
  return { ...meta, component: Bound };
}
