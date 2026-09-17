# QIFUTIL Code Review Status

Current as of v1.10.0. `CODE_REVIEW_FINDINGS.md` is the older, longer write-up;
where the two disagree, this file is right.

## ✅ Fixed

| Issue | Where it was | Fixed in |
|-------|--------------|----------|
| Nil pointer from `FindStringIndex` | categories.go, tags.go | v1.10.0 |
| Silent file creation errors | accounts, categories, payees, tags | v1.9.0 |
| Ignored file read errors | list commands carried on with empty content | v1.10.0 |
| Directory change pattern | transactions.go changed the process working directory | v1.10.0 |
| `os.Exit()` throughout | 39 sites; only `Execute` still exits | v1.10.0 |
| Unchecked date parsing | century read from the separator, bad dates reported | v1.10.0 |
| Regex compiled inside the loop | the pattern is gone entirely | v1.10.0 |
| No bounds check on regex matches | records are read by field code, not by group index | v1.10.0 |
| Hard-coded date assembly | `utils.ParseQIFDate`, unit tested | v1.10.0 |
| Resource leak when splitting files | error paths close the file and skip the account | v1.10.0 |
| "catergory" typo | payees.go, categories.go | v1.9.0 |
| Validation reporting that never ran | duplicates, unmapped values and unused rules are computed now | unreleased |
| `applyMapping` scanned the whole map | a lookup, and one count per mapping instead of a line per value | unreleased |
| Unmapped list ordered at random | ranged over a map; ordered by frequency now | unreleased |
| Summary named a log file that is not written | it pointed at `validation.log` | unreleased |

## 🟠 Open

### Ignored regex compilation errors
`regex, _ := regexp.Compile(...)` still discards the error in three places:
`categories.go:115`, `payees.go:83`, `tags.go:118`. These patterns are constants,
so a failure would be a defect in the program rather than bad input, but the
error should be returned rather than dropped.

### Unknown `--csvColumns` names produce blank columns
`buildCSVRow` falls through to an empty string for a name it does not recognise,
so a typo yields a silently blank column. `--csvColumns` is also ignored without
comment for JSON and XML.

### Money held in `float64`
`balance-history` accumulates balances in `float64`. Over a long register the
running total drifts. Integer cents is the usual fix.

### CSV injection
Values are quoted but not guarded against a leading `=`, `+` or `@`. Low risk for
your own data, and note that a naive fix breaks negative amounts.

### `friendlyError` is never called
`cmd/errors.go` turns common failures into multi-line guidance and has no
callers. Now that commands return errors there is somewhere to route them
through.

### Empty `PostRun` hooks
`accounts.go` and `export.go` define `PostRun` and `PersistentPostRun` bodies
that do nothing.

### `go vet` failure in a test
`pkg/config/config_test.go:270` has unreachable code. Worth fixing and wiring
`go vet` into CI.

## 🔵 Noted, not planned

**Split transactions** are supported through `--expandSplits`. Without the flag a
split still exports as a single row at the transaction total, which is the
historical behaviour and is deliberate.

**Mapping files are not validated** beyond skipping rows whose target is empty. A
malformed row is ignored rather than reported.

**`ValidationTracker` locking** is sound; the tracker is mutex guarded and the
export is single threaded.
