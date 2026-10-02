# Media Recommendations

## Use Case

A "recommended for you" engine over a media catalog that optimizes for conversion -- a viewer watching at least half of what it recommends -- and trains continuously on view, play, and home-page events.

## When to Use

- Video, music, podcast, or article catalogs with user events flowing in
- A home-page or "watch next" rail personalized per viewer
- When you want Google's model rather than your own ranking

## What This Creates

- A RECOMMENDATION engine in `global` over the referenced MEDIA store (enrolled in `SOLUTION_TYPE_RECOMMENDATION`), in `default_collection`
- The `recommended-for-you` model with a `cvr` objective on `watch-percentage` at 0.5, queried with home-page context, in `TRAINING` state

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `mediaRecommendationEngineConfig.type` | `recommended-for-you` | `others-you-may-like`, `more-like-this`, or `most-popular-items` (with `mostPopularConfig.timeWindowDays`). |
| `mediaRecommendationEngineConfig.optimizationObjective` | `cvr` | `ctr` to optimize clicks instead of watch time; drop the objective config with it. |
| `mediaRecommendationEngineConfig.optimizationObjectiveConfig.targetField` | `watch-percentage` | `watch-time` with a value in seconds. |
| `mediaRecommendationEngineConfig.trainingState` | `TRAINING` | `PAUSED` to hold the model and stop training cost. |
| `industryVertical` | `MEDIA` | `GENERIC` for a non-media catalog (then omit the media config). |

A recommendation engine has no collection id and no KMS key of its own (the spec walls both), and it takes at most one data store. Predictions bill per thousand; training bills while `TRAINING`.
