# QIFUTIL Release Notes

## Version 1.11.0 - The Wizard, and Reporting That Runs

**Release Date:** September 18, 2026

### The wizard was crashing

`qifutil wizard` died before listing a single account:

```
Step 3: Let me check what accounts are in your file...
panic: runtime error: invalid memory address or nil pointer dereference
```

Step 3 lists accounts by copying the `list-accounts` command and calling `Run` on
the copy. When the commands stopped calling `os.Exit` in 1.10.0 and moved to
`RunE`, `Run` became nil. The wizard has been unusable for the whole of 1.10.0.
It is the entry point new users are pointed at, and nothing caught it: its only
test covered a helper rather than the flow.

Two more problems at the prompt people reach first:

- A mistyped filename printed "Could not find file" and **ended the wizard**, so
  the whole run had to be started again. Every other prompt re-asks on bad
  input; this one gave up.
- Pressing Enter was accepted. The empty path resolved to the current directory,
  which exists, so it passed the check and the *next* answer was read as the
  output location. The wizard created a directory named after the QIF file and
  carried on with every later answer shifted by one.

The wizard also printed `Conversion completed!` whatever happened. Since the
commands now report failure rather than ending the process, a run that failed on
a missing mapping file still finished with a cheerful summary.

### The wizard offers the newer options

Neither `--preserveOriginalCategory` nor `--expandSplits` could be reached from
the wizard. Both are about keeping categories intact, which is what most of the
recent work has been about:

```
Record the original category in the Notes column, so a mapped category
can be traced back? (y/n): y

Export each part of a split transaction as its own row, keeping its own
category? (y/n): y
```

The first is asked only when a category mapping was given, since without one
there is no original to record. Both are saved with the rest of your answers, so
a saved configuration reproduces the same export.

### Validation reporting that actually runs

The summary and the validation log had full sections for duplicate transactions,
unmapped values and unused mapping rules. None of the three was ever computed:
the tool reported that it had checked for duplicates, and it had not.

```
⚠️  Data Validation Summary:
  • Potential duplicates: 1 groups detected
    - 2023-01-15 | Grocery Store | -45.23 (2 times)
  • Category mapping: 2 rules never used
    - "Also:Wrong"
    - "Typo:Categry"
  • Unmapped values: 1 different payees/categories not in mapping files
```

Duplicates are counted per account on date, payee and amount, and tallied once
per transaction rather than once per row, so a split is not mistaken for a
repeat. The count does not cross account boundaries, because the same date,
payee and amount in two accounts is what a transfer between them looks like.
Unmapped values are recorded only where a mapping file was supplied; without one
every value is unmapped and the list says nothing.

The unmapped list also came out in a different order on every run, because it
ranged over a map. It is ordered by frequency now, so the values worth adding to
a mapping file come first. Entries read `category:Food:Dining` with the kind
jammed onto the value; they now read `category "Food:Dining"`. And the summary
pointed at `validation.log`, which neither export writes.

### Mapping output

Tracking which rules a mapping used meant `applyMapping` became a type that owns
its rules. It looks a value up rather than scanning every rule for each
transaction, and the line it printed for every value changed - thousands on a
real file, burying the summary - is replaced by one count per mapping:

```
Category mapping: applied to 2 values
```

### Split files

Both commands that split their output closed the full file and opened the next
one *after* writing the record that filled it. When the count divided exactly,
that file was opened for records that never arrived and left holding nothing but
a header:

```
10 transactions at 5 per file -> Checking_1.csv, Checking_2.csv,
                                 Checking_3.csv (header only)
```

Splitting exists for Monarch's row limit, so anyone who hits it squarely got a
file that may fail on import.

### Column names are checked

A misspelled name in `--csvColumns` fell through to an empty string, so the
column was blank in every row and nothing said why:

```
$ qifutil export transactions ... --csvColumns "Date,Merchent,Amount"
Error: unknown column "Merchent" in --csvColumns; valid names are
Date,Merchant,Category,Account,Original Statement,Notes,Amount,Tags
```

Passing the flag with JSON or XML output, where it does nothing, prints a note
rather than failing.

### Housekeeping

- Five places compiled a QIF pattern and discarded the error with `_`. The
  account block pattern was written out identically in seven command files. All
  of it is now compiled once, at package level.
- `friendlyError` was deleted rather than wired in. Its branches matched unix
  phrasing this tool never produces - Windows reports "The system cannot find
  the file specified" - so every error would have fallen to its default branch
  and been replaced by a generic help block.
- `go vet ./...` is clean, and CI runs it.

### Testing

91 tests, up from 67. The wizard had one, covering a helper; it has nine now,
each driving the built binary through a pipe as a person does.

A note recorded in CODE_REVIEW.md: `float64` money in balance-history was listed
as a correctness risk and has been measured rather than assumed. 20,000
transactions of -0.07 against an opening balance of 9,999,999.99 end at exactly
9,998,599.99. Not worth a change. CSV injection is likewise left alone
deliberately: the data is your own file going into your own spreadsheet, and
every mitigation alters the payee name that reaches the import.

---

## Version 1.10.0 - Correct Parsing and Reported Failures

**Release Date:** September 16, 2026

This release is mostly about transactions that were being lost. QIF was read with
a single regular expression that required five fields in a fixed order. Records
that did not match were skipped without a word, and the only symptom was a
smaller number next to "transactions found". The committed sample file was itself
affected: it exported 53 of its 58 transactions.

### New Options

**`--preserveOriginalCategory`**

Category mappings are destructive. Once `Insurance:Auto` has been collapsed into
`Insurance`, nothing in the export says what it used to be. This flag appends the
pre-mapping category to the Notes field when a mapping changes it:

```
"Semi-annual premium [Original Category: Insurance:Auto]"
```

Notes is a field Monarch Money imports, so the reference survives the import. The
note is added only where a mapping actually changed the category.

**`--expandSplits`**

A split transaction records several categorised line items under one entry. The
export wrote a single row at the transaction total under the top level category,
so a $100 shop divided between groceries and household goods arrived as $100 of
groceries. With this flag each line item becomes its own row, carrying its own
category, amount and memo. The parent row is not written as well, so the account
total is unchanged. Line items that do not sum to the total are still exported,
with the difference reported.

Both flags are off by default.

### Transactions No Longer Silently Dropped

QIF is now read field by field rather than matched against one pattern. Records
are split on the `^` terminator and each line read by its code, so field order
does not matter and optional fields may be absent. This recovers:

- Records with no `U` line. `U` is a Quicken extension many QIF writers omit; the
  amount now falls back to the `T` field.
- Records with no `C` (cleared) or `L` (category) line.
- Records whose fields appear in an unexpected order. A payee written before the
  number was dropped; a payee written after the category produced a row with a
  blank merchant, original statement and memo.

All five commands that read transactions were affected and all are fixed:
transactions, categories, payees, tags, account-stats and balance-history.
balance-history was the worst placed to lose records, because every transaction
it failed to read was left out of the arithmetic and the balance it reported was
simply wrong.

### Dates Before 2000

Quicken marks the century with the separator before the year: a slash for the
1900s and an apostrophe for the 2000s. Only the apostrophe form was matched, and
the year was pasted onto a fixed `20` prefix, so a register spanning the 1990s
lost every transaction written as `3/15/99`. The separator is now read, along
with four digit years and the space padding Quicken uses for narrow fields. A
date outside the possible range is reported and skipped rather than being rolled
silently into the following month by the date library.

### File Names

Account names went into file names untouched, and Quicken allows characters a
file name cannot hold:

| Account | What happened |
| --- | --- |
| `Amex / Joint` | the path could not be found, and the error ended the whole run |
| `Savings*2024?` | the same, from an invalid character |
| `Fidelity: Roth` | no error at all - NTFS read the colon as an alternate data stream, leaving an empty file named `Fidelity` with the CSV hidden inside it |

Forbidden characters are now replaced with an underscore and the substitution is
reported, so the file can still be matched to its account. Two accounts whose
names sanitise alike get distinct files rather than one overwriting the other. A
write failure skips that account and the export continues, with a count of what
was skipped in the summary; one unusable account name no longer costs every other
account its export.

### Other Fixes

- **XML output was not XML.** Records were accumulated only for JSON, so the XML
  marshalling never had anything to write and every transaction fell through to
  the CSV writer. The output was CSV rows under an XML declaration.
- **The wizard discarded its mapping files.** It prompted for all four, accepted
  the paths, and exported unmapped, because the paths were collected into local
  variables while the export read package level ones that nothing set.
- **`Original Statement` duplicated `Merchant`.** It was filled in after payee
  mapping had rewritten the value. It now holds the payee as it appeared in the
  file. *This changes existing output for anyone using `--payeeMapFile`.*
- **Relative `--inputFile` always failed.** The export changed into the output
  directory before checking the input file existed, so a relative path was looked
  for in the wrong place.
- **`-i` and `-o` did not exist on `export transactions`,** though every usage
  example in the help text and README showed them.
- **A read failure was printed and then ignored** by the categories, payees, tags
  and accounts commands, which carried on with empty content, wrote an empty list
  and reported success.
- **`account-stats` reported the first account's figures for every account,**
  because it searched the whole file for its account block inside the loop.

### Failures Are Reported

Thirty-nine failure paths ended in `os.Exit`. Commands now return errors and the
program decides the exit status in one place. The exit status is unchanged, 1 on
failure and 0 on success, but errors are printed by cobra as `Error: <message>`,
so messages lost their own `Error:` prefix and initial capital. Usage text is no
longer printed when an export fails; an unknown flag still prints it.

None of those paths could be tested before, because calling a command in-process
ended the test binary.

### Visibility

- The `QIFIMPORT` tag is on by default and the export now says so, naming the
  flag that turns it off. The default is unchanged.
- Records with no date are counted and reported rather than passing unremarked.

### Testing

Coverage went from 20 tests to 66 top level tests, 126 counting table cases
separately. The QIF record parser, the date parser and the file name sanitiser
are unit tested in isolation. The categories, payees, tags, account-stats and
balance-history commands had no tests at all and now have them.

---

## Version 1.9.0 - Enhanced QIF Compatibility & Validation

**Release Date:** January 16, 2026

### Major Features

#### 1. Improved QIF Parsing - Real Quicken Support ✨
- **Fixed optional Payee fields** - Quicken exports often omit the P (Payee) field for certain transactions. The regex now correctly handles this, capturing 100% of transactions instead of skipping those without payees.
- **Named group extraction** - Replaced fragile hard-coded indices with named capture groups for more robust parsing
- **Better real-world compatibility** - Now properly handles actual Quicken exports with optional transaction fields

#### 2. Detailed Validation Logging 📊
- **Transaction-specific details** - Validation logs now include exact transaction information (date, payee, amount, category) for any issues found
- **Separate log files** - Transactions and balance history now generate separate validation logs to prevent data loss
  - `transactions_validation.log` - Issues from transaction export
  - `balance_history_validation.log` - Issues from balance history export
- **Better debugging** - Easily identify and fix specific problematic transactions

#### 3. Mapping File Robustness ✅
- **Fixed Excel-created mappings** - When mapping files were created/edited in Excel and saved as CSV, empty cells would create mappings to empty strings, blanking out data
- **Smart empty entry handling** - Now skips any mapping entries where the target value is empty, preserving original data
- **No more accidental data loss** - Your payees and categories are safe even with incomplete mapping files

#### 4. Zero-Amount Transaction Filtering 🚀
- **New `--skipZeroAmounts` flag** - Exclude zero-amount transactions (0.00 or 0) during export
- **Detailed reporting** - Shows exactly which transactions have zero amounts in the validation log
- **Useful for data cleaning** - Easily remove pending/placeholder transactions
- **Available in wizard** - Interactive option to skip zeros during the guided process

#### 5. Improved Configuration Flow 💾
- **Smarter config reuse** - When loading a saved configuration and choosing to use it unchanged, the tool no longer asks to save again (reduces redundant prompts)
- **Better UX** - "Use these settings?" flow prevents re-asking questions when reusing configs

### Bug Fixes

- ✅ Fixed payee/category missing from exports when optional QIF fields weren't present
- ✅ Fixed mapping files creating empty payee/category values
- ✅ Fixed validation log overwriting when running both transactions and balance history
- ✅ Fixed redundant save config prompt when reusing loaded configurations

### Improvements

- Better error handling for edge cases in QIF parsing
- More informative validation output with transaction details
- Cleaner code using named capture groups instead of hard-coded indices
- More reliable mapping file handling

### Technical Details

**QIF Parsing Changes:**
- Regex now uses named capture groups: `(?<payee>...)`, `(?<category>...)`, etc.
- Payee field (P) is now optional in the regex pattern
- `getGroup()` helper function extracts values by name, making parsing more robust

**Mapping File Validation:**
```go
// Only add mapping if the target value is not empty
if value != "" {
    mapping[key] = value
}
```

**Validation Logging:**
- New `WriteValidationLogWithName()` method allows separate logs per command
- Transaction issues stored as `TransactionIssue` structs with full details

### Migration Notes

**For Users:**
- No breaking changes - all existing functionality still works
- Saved configs from v1.8.4 are compatible
- If you had mapping files with empty entries, they will now be ignored (which is the correct behavior)

**For Developers:**
- The regex pattern now accepts optional Payee fields
- Named groups are used for field extraction - if you extend this, use `getGroup(match, "fieldname")` instead of hard-coded indices

### Known Limitations

- QIF files must still be valid Quicken exports
- Date format validation requires YYYY-MM-DD format
- Maximum 10,000 records per file before splitting (configurable with `--recordsPerFile`)

### Testing

Tested with:
- Real Quicken exports (5000+ transactions)
- Multiple account types (Checking, Savings, Credit Card)
- Real mapping files with incomplete entries
- Various QIF variations and edge cases

### Previous Changes (v1.8.4)

For changes from previous versions, see the full commit history on GitHub.

---

## Support

- **Documentation:** See README.md for detailed usage instructions
- **Issues:** Report bugs on GitHub
- **Questions:** Check the help output with `qifutil --help`
