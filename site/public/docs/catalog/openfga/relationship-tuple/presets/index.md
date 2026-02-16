---
title: "Presets"
description: "Ready-to-deploy configuration presets for Relationship Tuple"
type: "preset-list"
componentSlug: "relationship-tuple"
componentTitle: "Relationship Tuple"
provider: "openfga"
icon: "package"
order: 200
presets:
  - slug: "01-user-document-access"
    rank: "01"
    title: "User-Document Access Tuple"
    excerpt: "This preset grants a specific user direct access to a document. This is the most fundamental relationship tuple in OpenFGA -- a single user gaining a specific permission on a single resource. It..."
  - slug: "02-group-membership"
    rank: "02"
    title: "Group Membership Tuple"
    excerpt: "This preset adds a user to a group, enabling group-based access control. When the authorization model grants `group#member` access to resources, all members of the group inherit those permissions...."
---

# Relationship Tuple Presets

Ready-to-deploy configuration presets for Relationship Tuple. Each preset is a complete manifest you can copy, customize, and deploy.
