# Architecture Patterns

This folder holds authored architecture patterns: named compositions of two
or more catalog kinds, with validated manifest wiring and the trade-offs
behind each choice. This is NOT a provider directory -- it sits beside the
catalog's generated indexes because patterns span providers.

Patterns are the catalog's judgment layer for anything bigger than one kind:
where a kind's `GUIDE.md` teaches wisdom about that kind alone, a pattern
teaches how kinds compose -- and what each composition looks like on the
architecture diagram the platform renders.

## The patterns

| Pattern | The truth it owns |
|---|---|
| [namespace-ownership.md](namespace-ownership.md) | Who owns a namespace: a dedicated KubernetesNamespace component vs the `createNamespace` flag — including the sole-tenant infra case where the flag is the normal shape |
| [operator-prerequisite.md](operator-prerequisite.md) | Custom-resource kinds need their controller watching their namespace; the dependency is registry metadata, draws no diagram edge, and operators disagree on default watch scope |
| [observability-stack.md](observability-stack.md) | "Give me observability": the assembled stack (collector, Loki, Tempo, kube-prometheus-stack, Grafana) vs the all-in-one (Signoz + composed ClickHouse), with the fully wired example |
| [stateful-kind-disaster-recovery.md](stateful-kind-disaster-recovery.md) | "Can I get it back?" for the PostgreSQL database, the MongoDB replica set, and the OpenBao vault: the store by reference, keyless where the cloud allows and a token pair on R2 (never keyless), one path per live instance, the restore as a second declared instance, and the one step that stays with the operator |

## Reading a pattern

One file per pattern. Frontmatter declares the kinds the pattern composes,
so tooling (and you) can find every pattern touching a kind:

```
rg -l "KubernetesNamespace" _patterns/    # every pattern composing a kind
```

Each pattern states the problem, the composition choices with their
consequences (state ownership, diagram shape, customization surface), and
complete manifests you can adapt.

## Writing a pattern

Patterns are contributed through pull requests -- the file you edit here is
exactly the file agents read, in the repo, the release archive, and the
packaged skill alike. Embedded complete manifests must validate against
their schemas, and declared kind names must resolve: catalog checks enforce
both, so a pattern can age gracefully but can never be silently
schema-wrong. Authoring standards live in the repository's contribution
docs.
