# Test Suite

Current as of v1.10.0. Run everything with `go test ./...`.

## Layout

| File | Tests | Covers |
|------|-------|--------|
| `cmd/transactions_test.go` | 27 | the transactions export: output formats, column selection, mappings, splits, dates, file names, failure paths |
| `cmd/lists_test.go` | 9 | categories, payees, tags, account-stats and balance-history |
| `cmd/wizard_test.go` | 1 | the wizard forwards the mapping files it collected |
| `pkg/utils/record_test.go` | 6 | the QIF record parser: field codes, optional fields, field order, split lines |
| `pkg/utils/qifdate_test.go` | 2 | the date parser: century from the separator, four digit years, padding, invalid input |
| `pkg/utils/validation_test.go` | 11 | the validation tracker |
| `pkg/utils/filename_test.go` | 1 | replacing characters a file name cannot hold |
| `pkg/utils/category_test.go` | 1 | splitting a category from its tag |
| `pkg/config/config_test.go` | 9 | wizard config load and save |

67 top level tests; 126 counting table cases separately.

## What the tests are for

Most of them exist because something was silently wrong. The valuable ones assert
on behaviour that used to fail quietly:

- **Records that the old pattern could not read.** Missing `U` or `C` lines,
  fields in an unexpected order, dates written in the 1900s form. These were
  dropped without a word, so the tests check the payee or amount actually reaches
  the output rather than checking a count.
- **File names.** An account called `Fidelity: Roth` produced no error and no
  visible file, because NTFS read the colon as an alternate data stream. The test
  asserts on the sanitised name.
- **Failure paths.** These could not be tested at all until commands returned
  errors instead of calling `os.Exit`, which ended the test binary. Fourteen of
  them are covered now.
- **Defaults staying put.** `--expandSplits` and `--preserveOriginalCategory` are
  off by default, and there are tests asserting the output is unchanged without
  them.

## Notes on writing tests here

**Commands share package level variables.** The flags are package globals, so a
test that sets one leaks into the next. `setupTransactionExport` saves and
restores them; use it, or save and restore by hand as the `lists_test.go` tests
do.

**Argument checks live in `PreRunE`, not `RunE`.** A test that calls `RunE`
directly skips them and gets a different error. Run both, in that order, as
cobra does.

**`test.CaptureOutput` restores stdout in a deferred call,** so a `t.Fatalf`
inside the captured function is safe. It was not always: `t.Fatalf` unwinds
through `runtime.Goexit`, which skipped a plain restore and left stdout pointing
at the pipe for the rest of the run.

**`TestMain` builds the binary once** for the few tests that need a subprocess.
Almost nothing needs one now that commands return errors.

## Known gap

`go vet` reports unreachable code at `pkg/config/config_test.go:270`. It is the
only vet failure in the tree.
