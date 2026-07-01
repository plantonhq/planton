/**
 * The planton.app landing sections, in three movements (see app/page.tsx):
 *   1. Skimmer top    — Hero, CatalogStats, TrustStrip
 *   2. The conversation — Interlude (prologue) → Conversation → Interlude (bridge) → Origin
 *   3. Feature showcase — Features, Payoff, WhereItFits, Horizon, Catch, FinalCta
 *
 * Shared primitives: Interlude (centered narration), FeatureRow (alternating
 * feature row), Line (a toned/spoken line), Section/Speaker/InlineCode. Product
 * visuals come from components/showcase. Copy is lifted from the approved
 * messaging spec — change wording deliberately.
 *
 * Convention: numbered StoryStations belong to the conversation ONLY; the other
 * movements use their own layout so the page stops reading as one long dialogue.
 */

// Movement 1
export { Hero } from "./Hero";
export { CatalogStats } from "./CatalogStats";
export { TrustStrip } from "./TrustStrip";

// Movement 2
export { Interlude } from "./Interlude";
export { Conversation } from "./Conversation";
export { Origin } from "./Origin";

// Movement 3
export { Features } from "./Features";
export { Payoff } from "./Payoff";
export { WhereItFits } from "./WhereItFits";
export { Horizon } from "./Horizon";
export { Catch } from "./Catch";
export { FinalCta } from "./FinalCta";

// Shared primitives
export { Line } from "./Line";
export { FeatureRow } from "./FeatureRow";
