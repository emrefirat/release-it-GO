# Progress Tracking and Bug Management

## PROGRESS.md Requirement

- **Update `PROGRESS.md` at the end of every development session.**
- Mark completed items with `[x]`.
- Update progress percentages.
- Write important decisions and blockers in the Notes section.
- Update the last-updated date.

## When Adding a New Phase

1. Create a `docs/phase_N.md` PRD document (follow the format of existing phase files).
2. Add a new row to the phase table in `PROGRESS.md`.
3. Add a phase detail section (to-dos, notes).
4. Add a new row to the change history.

## Bug Tracking

- Every detected bug is added to the Bugs section in `PROGRESS.md`.
- Open bug: `- [ ] BUG: Short description (date)`
- Closed bug: `- [x] BUG: Short description (date) → solution summary`
- After the bug fix commit, mark the line `[x]`.
- Edge cases and first-release scenarios must be tested.

## End-of-Session Process

1. Write what changed to PROGRESS.md.
2. If there's a bug, add it to the Bugs section.
3. Add a date + summary to the change history.
4. Commit (in conventional format).
