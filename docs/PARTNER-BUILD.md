# Building nofx on a machine without a C compiler (partner mirror)

**Wave:** W-CGOFREE-SQLITE-UPSTREAM (2026-09-17). Not a knob — no guide entry.

The default build links `gorm.io/driver/sqlite`, which sits on `mattn/go-sqlite3`
and needs cgo (a working `gcc`). `modernc.org/sqlite`, used for the raw
`database/sql` path, is already pure Go. Machines without gcc (the partner
mirror "Binnie") build with the `cgofree` tag, which swaps BOTH sqlite backends
for pure-Go equivalents:

```
go build -tags cgofree -o nofx-bin .
go test  -tags cgofree ./store/... ./researchsnapshot/...
```

| tag | database/sql `"sqlite"` driver | GORM dialector |
|-----|-------------------------------|----------------|
| (default) | `modernc.org/sqlite` (pure Go) | `gorm.io/driver/sqlite` (mattn, cgo) |
| `cgofree` | `github.com/glebarez/go-sqlite` (pure Go) | `github.com/glebarez/sqlite` (pure Go) |

`glebarez/go-sqlite` is a fork of the `modernc.org/sqlite` driver layer over the
same `modernc.org/sqlite/lib` C translation, so both tag sets run the same
SQLite engine; the owner's machine stays on the default because it is the
binary that has been live since day one and nothing in this wave changes it.

**Rule:** the ONLY package that imports a sqlite driver is
`store/sqlitedriver`. Everything else calls `sqlitedriver.Open(dsn)` or
`sqlitedriver.GormDialector(dsn)`. A second blank import
(`_ "modernc.org/sqlite"`, `_ "github.com/glebarez/go-sqlite"`) anywhere else
panics at init with `sql: Register called twice for driver sqlite` — that was
Binnie's first-boot panic after `researchsnapshot/archive.go` grew its own
import. `store/sqlitedriver.TestSingleRegistration` pins the invariant.

Binnie update script line (replaces the private rebase branch):

```
git pull --ff-only origin dev && go build -tags cgofree -o nofx-bin . && (cd web && npm run build)
```

Web side, same wave: `EquityChart` no longer throws on a missing
`total_equity` (renders the empty state / `0.00`), and a root
`<ErrorBoundary>` under `App` shows a one-line error with a Reload button
instead of a white screen.
