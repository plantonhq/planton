/**
 * The deck engine and its slide kit. A presentation is a list of SlideConfig
 * rendered by Deck; the slide components themselves compose the primitives
 * here. A hand-written deck lists its slide components directly; a deck
 * rendered from data binds its content to generic slides with bindSlide.
 * Meeting decks live in `@/components/meetings`, persona decks in
 * `@/components/decks`. Nothing in this folder knows who is in the room.
 */
export { Deck } from './Deck';
export type { DeckProps, SlideConfig, SlideComponentProps } from './Deck';
export { bindSlide } from './bind-slide';
export { SlideFrame } from './slide-frame';
export * from './primitives';
