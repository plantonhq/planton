# DigitalOcean Database Kafka Schema -- Operational Guide

What experience with this component teaches that the field reference cannot.

## Every change drops ALL prior versions -- the kind's loudest warning

The provider has no update path: changing the definition (or type, or subject name) destroys the subject and re-registers the new document as version 1. Consumers that pin older schema versions lose them the moment the replacement lands. If your consumers rely on registry-mediated compatibility across versions, do NOT evolve schemas through this resource -- evolve them through your producers' registry client (which appends versions), and use this resource only to declare the founding schema.

## The registry needs a General Purpose Kafka plan

DigitalOcean's schema registry exists only on General Purpose (dedicated-CPU: `gd-*`, `c2-*`, `m3-*`) Kafka clusters. Point this resource at a Basic-plan cluster and every registry call answers `412 schema registry is disabled for this cluster`; asking that cluster to turn the registry on (the `schema_registry` toggle in the cluster's advanced configuration) answers `422 schema registry not supported for current plan`. A fresh General Purpose cluster comes with the registry enabled (measured 2026-09-17). If a General Purpose cluster somehow has it off, enable it once in the control panel (Settings -> Advanced Configuration) or with `PATCH /v2/databases/{id}/config` `{"config":{"schema_registry":true}}` -- neither the DigitalOceanDatabaseCluster kind nor the Terraform provider carries that toggle today.

## JSON schemas are canonicalized for you; protobuf is not

The registry stores every Avro and JSON Schema definition in canonical form -- object keys sorted, no whitespace -- and the provider stores that canonical text verbatim on every read. Left alone, that means a schema written the natural way (`type` first, then `name`, then `fields`) would look "changed" on every refreshed plan and, because every field is create-only, would propose the destroy-and-drop above forever. Both provisioners therefore render Avro and JSON definitions into the registry's canonical form before sending, so you can write the JSON in any key order and with any whitespace and a re-apply never proposes a change. The one thing that still counts as a change is a real change: a different field, type, default, or doc string (proven live 2026-09-17 with a human-ordered Avro record under the apply-twice check). A definition that is not valid JSON fails at plan time, not at the registry.

Protobuf definitions are text, not JSON, and are sent verbatim -- but the registry reformats them too (measured: a blank line inserted after the `syntax` line), and the provider compares the reformatted text with yours. Until the provider compares normalized protobuf, a protobuf subject re-plans a replacement on every refreshed Terraform plan (Pulumi's preview does not refresh, so it stays quiet). Manage protobuf subjects with Pulumi, or author the text exactly as the registry renders it (register once, read it back, paste that), and never apply a Terraform plan that proposes replacing a protobuf subject you did not mean to change.

## Compatibility level is not here

The registry's subject compatibility level (BACKWARD, FULL, etc.) has no surface in the provider at the pinned version -- DigitalOcean's API carries it, the provider does not. What this resource registers is the document itself; compatibility policy stays whatever the registry defaults to (or whatever was set out-of-band).

## Imports do not work

The upstream importer is broken at the pinned provider version (it never restores the subject name, so the post-import read addresses nothing). Adopting an existing subject therefore means letting this resource re-register it. That is gentler than it sounds: registering a subject with a definition the registry already holds is accepted and returns the SAME schema id and version (measured 2026-09-17 -- an identical re-register is a no-op on the registry's side), so adopting a subject whose current definition your manifest matches keeps its history. Adopting with a DIFFERENT definition appends a new version under the registry's compatibility rules, exactly as a producer would. Only a later change through this resource -- which destroys first -- drops versions.

## What is deliberately NOT here

Multi-version subject management and compatibility levels (no provider surface -- see above); topics (their own kind, DigitalOceanDatabaseKafkaTopic); and registry credentials (the cluster's users).
