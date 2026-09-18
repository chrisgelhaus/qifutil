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
| Empty trailing file on an exact split boundary | both exports opened a file for records that never arrived | unreleased |
| Unknown `--csvColumns` names | a typo produced a blank column; refused now, with the valid names listed | unreleased |
| Regex errors discarded with `_` | patterns are compiled once at package level | unreleased |
| The account block pattern written out seven times | one shared variable | unreleased |
| A broken `contains` helper in config_test.go | six assertions passed whatever they were given | unreleased |
| Empty `PostRun` hooks | removed from accounts and export | unreleased |
| `go vet` failure | the tree is clean, and CI runs `go vet ./...` | unreleased |
| `friendlyError` never called | deleted; see below | unreleased |
| Summary named a log file that is not written | it pointed at `validation.log` | unreleased |

## 🟠 Open

Nothing outstanding from this review.

## 🔵 Noted, not planned

**CSV injection.** Values beginning with `=`, `+` or `@` are exported as they
appear. The data is the user's own Quicken file going into their own
spreadsheet, so the untrusted-input threat this warning is written for does not
apply here, and every mitigation alters the payee name that reaches the import:
`+1-555-CALL` would become `'+1-555-CALL`. Deliberately left alone.

**`friendlyError`** was deleted rather than wired in. Its three branches looked
for unix phrasing (`no such file`, `permission denied`) that this tool never
produces: Windows reports "The system cannot find the file specified", and the
commands name the file and the operation themselves. Routing errors through it
would have replaced specific messages with a generic help block.

**`float64` money** was listed here as a correctness risk. Measured rather
than assumed: 20,000 transactions of -0.07 against an opening balance of
9,999,999.99 ends at exactly 9,998,599.99. float64 carries 15 to 16
significant digits, so reaching half a cent of error needs orders of magnitude
more arithmetic than this tool performs. Not worth a change.

**Split transactions** are supported through `--expandSplits`. Without the flag a
split still exports as a single row at the transaction total, which is the
historical behaviour and is deliberate.

**Mapping files are not validated** beyond skipping rows whose target is empty. A
malformed row is ignored rather than reported.

**`ValidationTracker` locking** is sound; the tracker is mutex guarded and the
export is single threaded.
