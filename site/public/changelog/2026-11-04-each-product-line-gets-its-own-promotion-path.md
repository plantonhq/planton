---
title: "Each Product Line Gets Its Own Promotion Path"
date: 2026-11-04
category: breaking
tags:
  - console
  - platform
  - mcp
excerpt: "Each product line can now have its own Promotion Path, and a stage can hold several environments side by side. Pushes to trigger branches now deploy only environments on a path: an environment left out of your old order, or every environment if you never set one, gets no deployments from them. Mapped branches deploy as before. Check Environments → Promotion."
author:
  - name: Swarup Donepudi
    title: Founder
---

Until now, your environments sat in one order across the whole organization. That breaks down as soon as you run more than one product line. With one order for everyone, marking one team's production Promotion Only also stopped pushes to every environment after it, other teams' included. And nothing could say "after dev, both demo and prod come next." Now each product line with environments of its own can have its own **Promotion Path**: an ordered list of stages, where a stage can hold several environments side by side.

## Check Your Environments Today

A push to a service's trigger branch, the push that builds once and walks your environments, now deploys only environments that sit on a path. Anything under **Not on a Path** on **Environments → Promotion** gets no deployments from those pushes, and Promote into it is refused, until you add it to a path. The next push then reaches it, or you can deploy a build there directly. A new environment starts on no path too. Branches you mapped to an environment keep deploying it, as before.

- **If your organization had an order, it's now your first path.** Environments you never placed aren't on it, so pushes that used to reach them after the ordered ones no longer do. Each environment you had placed keeps its place, one per stage, and Promotion Only starts where it did.
- **Your product lines share that path until you split them.** Until then, one line's Promotion Only still stops the others' pushes. To give a line its own path, create one with **New Path** and **Move To** its environments there. Promotion Only belongs to each path, so when you move a production environment onto another path, put it in that path's Promotion Only stage, which each stage's **Promotion Only** switch sets: anywhere else, pushes start deploying it, and so does a branch mapped to it. The confirmation names every environment pushes would start deploying. Lines that share dev, staging and prod stay on one path.
- **If it had no order, it has no paths.** Pushes to trigger branches used to walk your environments alphabetically; now they deploy nothing, and Promote is refused, until you add a path.
- **New organizations start with no paths.** **New Path** on the Promotion page puts your environments in the order a build travels.

## How Pushes and Promote Work Now

A push can deploy in two ways. A branch you mapped to an environment deploys exactly that environment, whether or not it sits on a path, unless it's in a Promotion Only stage. A push to a service's trigger branch builds once and walks the Promotion Paths:

- **A push walks every path that holds an environment the service deploys to,** stage by stage, stopping where the path says the rest is reached only by promotion.
- **An environment on no path is never deployed by a push.** The run says so, naming the environment and how to reach it: add it to a path, or deploy the build there yourself.
- **Promote stays on its path.** Promoting into an environment on another path, or on no path, is refused before anything runs. The refusal says which path each environment sits on, and what to do instead. When the stage before holds several environments, you choose which one Promote carries from.
- **Siblings are promoted on their own.** With demo and prod both next after dev, a promotion waiting for approval at prod never holds up demo.
- **Promotion Only is set once per path,** from a stage onward. It covers that stage and every stage after it, including stages you add later, so adding a stage never lets pushes reach further than you meant. As before, an environment there changes only by promotion: a branch mapped to it doesn't deploy it, and a new mapping into it is refused.

A path is made of environments, and an environment sits on one path at most. So a product line gets its own path when it has its own environments; lines that deploy to the same dev, staging and prod share one path. Environments that hold only shared infrastructure usually stay off the paths, so pushes never redeploy them.

## Where You See It

**Environments → Promotion** draws each path, its stages and the environments that share a stage. You can add a path with **New Path**, move an environment with **Move To**, and remove a path with **Delete Path**. A change that alters what pushes deploy asks you to confirm first, and says which environments pushes will start or stop deploying.

Changing paths takes an Organization Owner or Infrastructure Admin, along with the right to update every environment the change affects; on a protected environment, even an Organization Owner needs that right granted on the environment itself. If you can see only some of your organization's environments, the page shows the part of each path you can see and tells you so, instead of showing an error.

A service's Delivery Path draws one row per path it sits on. An Infra Stack lives on one path, and its Delivery Path shows its part of that path; promoting a stack never leaves it.

[Deployment Environments](/docs/ci-cd/deployment-environments#promotion-paths-the-order-a-push-walks) in the docs covers how a push walks your paths, step by step.

## Your Coding Agent and the Terminal

The Planton MCP server reads your Promotion Paths and changes them as you, with `get_promotion_path_set`, `set_promotion_path` and `delete_promotion_path`. The tools tell your agent which environments a change would start or stop deploying and ask it to confirm with you first; it can only make the changes your own access allows. After your agent creates an environment, it can place it on a path with `set_promotion_path`, confirming with you first. The Planton Assistant and every connected coding agent are held to the same rules you are.

`planton promotion-path` and `planton env create --path` arrive with the next CLI release.

## Why This Matters

- **Each team's line can be its own.** Once a line's environments are on a path of their own, its Promotion Only stage no longer stops another team's pushes, and a promotion can't cross into another path.
- **Fan-out is first class.** "After dev, both demo and prod come next" is one stage, and each sibling moves at its own pace.
- **No silent skips.** An environment on no path is never deployed by a push, and the run names it and says why.
- **You see the effect before you save.** A change to a path tells you which environments pushes will start or stop reaching.
