# Starter Notebook

This preset gives a data scientist a ready Jupyter workstation on the
cheapest current-generation instance — billed hourly while running,
with the verified figure in the kind's generated estimate at
`catalog/_pricing/estimates/awssagemakernotebookinstance.yaml` — and
the everyday Python stack installed once at creation.

## When to Use

- The first notebook for exploration and prototyping
- Lightweight data work that doesn't need a GPU

## What You Get

- An `ml.t3.medium` instance with a 20 GB ML storage volume
- A one-time `onCreate` bootstrap installing pandas, scikit-learn, and
  matplotlib — run as root, well inside AWS's 5-minute script limit

## Customize

- Move installs that must survive platform patches into `onStart` (it
  runs on every start, including the first) — keep it fast or push
  long work to the background
- Grow `volumeSizeGb` any time (it updates in place); shrinking
  replaces the instance, so size generously up front
- Most other changes stop, update, and restart the instance — budget
  several minutes and batch them
