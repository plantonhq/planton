/**
 * The deck engine and its slide kit. A presentation is a list of SlideConfig
 * rendered by Deck; the slide components themselves compose the primitives
 * here. Meeting decks live in `@/components/meetings`; persona decks will sit
 * beside them. Nothing in this folder knows who is in the room.
 */
export { Deck } from './Deck';
export type { DeckProps, SlideConfig, SlideComponentProps } from './Deck';
export * from './primitives';
