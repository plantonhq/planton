# Hierarchical Document Access Model (DSL)

This preset creates an OpenFGA authorization model with hierarchical folder/document permissions where access is inherited through parent relationships (Google Drive-style). Viewers and editors on a folder automatically gain access to all documents and subfolders within it.

## When to Use

- Document management systems with folder hierarchies
- Applications needing inherited permissions (access to a parent grants access to children)
- Google Drive, Notion, or Confluence-style permission models

## Key Configuration Choices

- **DSL format** (`modelDsl`) -- human-readable, recommended format
- **Hierarchical inheritance** -- `viewer from parent` and `editor from parent` mean permissions cascade from folders to documents
- **Folder nesting** -- folders can have parent folders (`parent: [folder]`), enabling deep hierarchies
- **Document ownership** -- documents have a direct `owner` relation that does not inherit (intentional: ownership is explicit)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<store-id>` | ID of the target OpenFgaStore | `OpenFgaStore` status outputs |

## Related Presets

- **01-rbac-dsl** -- Use instead for flat RBAC without hierarchical folder inheritance
