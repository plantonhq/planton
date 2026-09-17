# Avro Event Schema

This preset registers an Avro record schema for a topic's message values, following the registry's `<topic>-value` subject naming convention so serializer libraries resolve it automatically.

## When to Use

- Declaring the founding schema for a new event topic before producers ship
- Any Kafka pipeline using Avro with registry-aware serializers

## Key Configuration Choices

- **`subjectName: orders-value`** -- the `<topic>-value` convention is what standard serializers look up; keys get a sibling `<topic>-key` subject if keyed with structured data.
- **Single-line schema string** -- a readability choice only: both provisioners render Avro into the registry's canonical form (keys sorted, no whitespace) before sending, so key order and formatting never count as a change.
- **Avro** -- the registry also accepts `json` and `protobuf`.

## What You Get

A registered subject producers and consumers resolve through the cluster's registry endpoint. Remember the kind's loudest rule: changing this definition REPLACES the subject and drops prior versions -- evolve schemas through producer-side registration if consumers depend on version history.
