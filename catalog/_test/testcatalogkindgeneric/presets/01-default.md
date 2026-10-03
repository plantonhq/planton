# Default

The canonical torture manifest: one of every generic field class populated —
explicit scalars beside defaulted ones, a nested message, a literal
value-or-reference, a `valueFrom` reference into another instance's outputs,
a map, a repeated field, and a sensitive field. Certification cases load this
preset as their known-good input; if this manifest stops validating or stops
round-tripping, the machinery broke, not the manifest.
