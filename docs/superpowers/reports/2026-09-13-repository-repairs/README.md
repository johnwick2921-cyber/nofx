# Repository control-boundary repairs — in progress

Owner requested detailed traced review and repairs until complete. Source base:
`63968be62e44db2fb07a92883e02127b9064b0be`; branch
`fix/repo-audit-control-boundaries-20260913`; session
`reporepair-8d38e6ac/Codex[unlisted]`. Isolated worktree
`/tmp/nofx-audit-repairs-20260913`. No production database, settings, account,
order, deployment or runtime changes. Source review evidence is on
`docs/repo-understanding-20260913` under its separately claimed report scope.

Spec freshness: `dfda15e1425ba38671deb7ffa7a4f211a27846e2 test: isolate session clock fixtures from weekly backfill workers`
was the last AUDIT-CHECKLIST change at acceptance. Current tracked CLAUDE-canon
is read; main stays deploy-only. Branch names identify provenance, not author.

## Confirmed and repaired so far

- Authenticated object ownership: production router regression reproduces a
  foreign trader DELETE returning 200 and removing its equity child row. The
  middleware now validates all query/body trader IDs and trader path IDs. The
  delete handler additionally checks ownership; the store authorizes before
  touching children and performs both deletes transactionally. Injected child
  deletion failure verifies rollback. Own-trader reads remain available.
- Q&A: no plan plus available bars reproduces a nil row panic. Creation time
  is absent (zero) when no row exists. Historical fallback uses the existing
  trader-scoped query, tested with a foreign plan and an empty owned trader.
- Strategy updates: oversized prompt reproduces HTTP rejection after persisted
  changes. Token validation now runs before write/log/reload. Regression compares
  stored name and configuration bytes after rejection.

## Verification to date

Go 1.25.13. Focused production-route/transaction/context/rejected-save tests fail
on the preceding implementation and pass on repaired code. Existing plan/risk
ownership and strategy create/edit preservation tests included. Tests use
throwaway SQLite files and synthetic bars, never the owner's database or wire.
A first sandbox run failed because the external Go cache was read-only; it was
rerun with approved cache access. That setup failure is not a product test result.

Full suite, race coverage where relevant, independent repair review, guide rev
finalization, publication and final artifact verification are **pending**.
No runtime incident, exploitation, real fill, or deployment success is asserted.
