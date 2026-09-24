# Copywriting

How the words on planton.ai change.

The site is data first: every sentence that states what Planton is or does lives in `src/data/` and names the chapter of the story it mirrors (`company/marketing/positioning/the-planton-story.md` in the company repository, mirrored as `src/data/story.ts`). Pages render records and state nothing. A copy change is a change to a record, or to the story with the record following.

```
content/copywriting/
├── _rules/
│   ├── update-planton-ai-copy-writing.mdc     # feedback -> draft with sources -> handoff
│   ├── implement-planton-ai-copy-writing.mdc  # handoff -> data edit -> build, captures, review, record
│   └── pdf_converter/                         # feedback PDFs to Markdown
├── _workspace/                                # the gitignored drop zone for materials
└── _stage-area/
    └── YYYY-MM-DD-<slug>/
        ├── draft-1.md                         # every sentence beside what backs it; the cold reads
        └── handoff.md                         # routes, data files, argument, review record, verification
```

Start with `_rules/update-planton-ai-copy-writing.mdc`. The site `README.md` names every law the words obey and how a change is proven.
