/**
 * THE single source of positioning truth for the website.
 *
 * What Planton is, is said in one sentence (`umbrella.sentence`), with no
 * analogy and no category label. It was chosen by cold readers spanning the
 * knowledge spectrum, from a non-technical founder to a platform engineer,
 * and it changes only when real conversations show it confuses people. The
 * line under it (`umbrella.firstAction`) says what to do first. The story's
 * chapter 2 (`company/marketing/positioning/the-planton-story.md` in the
 * company repository, mirrored in `story.ts`) carries both.
 *
 * The two hubs keep their names and one plain line each. They wear no
 * analogy: an analogy coined for one hub once escaped its scope and spent
 * six months describing the whole product.
 *
 * Any sentence on the site that states what Planton IS must read from
 * here — the same law src/data/pricing.ts applies to prices.
 *
 * Vocabulary: "Infra Chart" is the product noun; "template" is the plain
 * word that explains it. Pair them on first mention, and never capitalize
 * "Template" as if it were a product name.
 */

export const POSITIONING = {
  /** What Planton is, in one sentence, and the first thing a visitor does. */
  umbrella: {
    sentence:
      'Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs.',
    firstAction: 'Start by pasting one prompt into your coding agent.',
    /**
     * The mission, in the founder's words (2026-09-11). Platforms that hide
     * the cloud trade control for convenience; Planton's bet is that you can
     * keep the account, the policy, the state, and the bill and still have
     * the convenience. It is the stance behind every page, and it is a
     * contrast with hidden-cloud platforms -- NOT the argument a page makes
     * to an engineer who already has a coding agent and a cloud CLI, because
     * that engineer already has both convenience and control. For them the
     * argument is what Planton adds to the agent (see `desktop`).
     */
    mission: 'Convenience without losing control.',
  },

  /**
   * The desktop distribution. Free for individuals, including commercial
   * use, because Planton makes its money when a team adopts it -- the page
   * says so as a business model, not a badge. The argument leads with what
   * Planton adds to the coding agent the reader already uses; control is the
   * second beat; competitors are described, never named; and the page
   * concedes that for a one-off bucket the agent alone wins.
   */
  desktop: {
    name: 'Planton Desktop',
    line: 'Your coding agent can already create cloud infrastructure. Planton makes it verifiable, recorded, and reusable — in your account, on your laptop, free for individuals.',
    whatItAdds: [
      'A typed catalog, not memory: schema lookups and validation run offline, so a wrong field fails before it touches your cloud.',
      'Verified before created: the monthly cost with its coverage stated, and the least-privilege permission policy derived from what is composed.',
      'A record, not a transcript: every deploy is an Infra Job with a live log and a revision history; state sits under a path you can list.',
      'Secrets the agent never reads: encrypted locally, key in your OS keychain, resolved on the runner at the moment of use.',
      'Push-to-deploy from your laptop: it watches GitHub, builds in a pod, deploys to your cloud, and writes the status back.',
      'What you built becomes a template: an Infra Chart redeploys into the next environment; agent commands do not compose.',
      'When you become a team, nothing is redone: the same manifests and model on planton.ai or your own cluster.',
    ],
    concession: 'For a one-off bucket, the agent alone wins. Planton earns its place on anything you will still be running in a month.',
  },

  /** The two hubs: a name and one plain line each, never an analogy. */
  infraHub: {
    name: 'Infra Hub',
    line: 'Every infrastructure change is a plan that shows its monthly cost and the exact permissions it needs before anything is created. Only the plan you reviewed is applied, and what you built can be saved as an Infra Chart your team redeploys.',
  },
  serviceHub: {
    name: 'Service Hub',
    line: 'Connect a GitHub repository and every push is built on Planton’s runners and deployed into your own Kubernetes, Google Cloud Run or Amazon ECS, with results written back to GitHub. No pipeline YAML, no Dockerfile required.',
  },
} as const;
