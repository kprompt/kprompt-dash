# TASK007: Refresh stale default models

## Goal
Implement a hygiene feature to refresh stale default models automatically.

## Requirements
- Identify default models by metadata.
- Detect staleness by comparing last updated timestamp against a configurable threshold.
- Refresh stale default models by replacing them with canonical copies.
- Update metadata lastUpdated timestamp on refresh.
- Fail closed on errors (do not refresh if error occurs).
- Provide tests covering:
  - Refreshing stale default models.
  - Skipping fresh or non-default models.
  - Handling malformed metadata gracefully.
- Update user-facing docs and changelog.

## Acceptance Criteria
- `refreshStaleDefaultModels` function implemented with above behavior.
- Unit tests cover all cases.
- CHANGELOG updated with user-visible feature note.
