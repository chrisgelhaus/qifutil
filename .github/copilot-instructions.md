# QIFUTIL Copilot Instructions

## Project Overview

**QIFUTIL** is a Go CLI tool for extracting and converting financial data from Quicken Interchange Format (QIF) files. It uses regex-based parsing to read QIF file structure and exports data to CSV, JSON, or XML formats.

### Core Architecture

- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) for command hierarchies
- **Main entry**: `main.go` → `cmd.Execute()` → command-specific Run handlers
- **Root command state**: Global variables in `cmd/root.go` (`inputFile`, `outputPath`, `selectedAccounts`, `startDate`, `endDate`, `outputFormat`)
- **Data processing**: Each command handles its own QIF parsing via compiled regex patterns

### Key Design Patterns

1. **Cobra Command Structure**: Commands are organized hierarchically
   - `root` → `export` (container for export subcommands) or `list` (container for list subcommands)
   - `export accounts|categories|payees|tags|transactions` (action commands)
   - Each command has `PreRun` validation and `Run` execution

2. **QIF Parsing via Regex**: 
   - Account blocks: `^!Account\nN{name}\nT{type}\n^\n!Type:(Bank|CCard)`
   - Transaction records: Named capture groups for date/amount/payee/category/memo
   - Line ending normalization: Convert `\r\n` to `\n` for consistent regex matching
   - See `cmd/transactions.go` for the master transaction regex pattern

3. **Global State Management**:
   - Persistent flags in `root.go` apply to all commands (`--inputFile`, `--outputPath`, `--accounts`, `--startDate`, `--endDate`)
   - Local flags for individual commands (e.g., `--output` for export commands)
   - The wizard (`cmd/wizard.go`) also sets these globals before delegating to transactions command

## Critical Developer Workflows

### Building
```pwsh
go build -o qifutil.exe
```

### Running Tests
```pwsh
go test ./cmd/...
go test ./pkg/...
```
Test files are in `cmd/transactions_test.go` and `pkg/utils/category_test.go`. Tests use a helper (`test/helper.go`) that provides `CreateTempDir()`, `CopyTestData()`, `AssertFileExists()`, etc.

### Development Iteration
1. Add test cases first in corresponding `*_test.go` files
2. Use `test/testdata/sample.qif` as test fixture (or create new fixtures in `test/testdata/`)
3. Most commands involve file I/O, so test with actual QIF files rather than mocks

### Debugging QIF Parsing
- QIF files are line-based text with structure markers (`!Type:`, `^` as record separator)
- Use `fmt.Printf()` debug statements in regex matching loops (common pattern)
- Test regex patterns independently before integrating into command logic

## Project-Specific Conventions

### Path Handling (Windows-First)
- Always clean paths with `filepath.Clean()` before using
- Convert relative paths to absolute with `filepath.Abs()` in command PreRun
- Sanitize PowerShell artifacts: trim `& ` prefix and outer quotes (`'` or `"`)
- See `pkg/utils/path.go` for `sanitizePath()` function
- Use `filepath` package, never hardcode path separators

### File Output Splitting
- Transactions automatically split files when exceeding `maxRecordsPerFile` (default: 5000)
- Output filename pattern: `{AccountName}_{fileNumber}.csv` (e.g., `Checking Account_1.csv`, `Checking Account_2.csv`)
- This is for Monarch Money compatibility

### Mapping Files
- Supported for categories, payees, accounts, and tags
- CSV format: two columns `"source","target"` (with quotes)
- Loaded via `loadMapping(filePath string) (map[string]string, error)` in `transactions.go`
- Applied during transaction processing to transform extracted data

### Category/Tag Splitting
- QIF categories can include tags as `category/tag` format
- Split via `utils.SplitCategoryAndTag()` which splits on `/` and trims spaces
- See `pkg/utils/category.go`

### Output Format Support
- CSV (default): Simple tabular format, optimal for Excel/Monarch
- JSON: Marshaled with `json.MarshalIndent()`
- XML: Uses Go struct tags (`xml:"fieldname"`) on custom type definitions (e.g., `transactionList`, `accountList`)

### User-Friendly Error Handling
- `cmd/errors.go` provides `friendlyError()` that translates common OS errors to guidance
- Used for file not found, permission denied, invalid date format cases

## Data Flow Example: Transaction Export

```
wizard.go (user input) 
  ↓ sets global vars ↓
root.go (inputFile, outputPath, outputFormat, etc.)
  ↓ calls ↓
transactions.go Run() handler
  ↓ loads & normalizes input ↓
Read QIF file → Standardize line endings (\r\n → \n)
  ↓ regex matching ↓
Find account blocks (accountBlockHeaderRegex)
  ↓ for each account ↓
Extract transactions (transactionRegexString with named groups)
  ↓ apply filters & mappings ↓
Date filtering, account filtering, category/payee/tag mapping
  ↓ split if needed ↓
Write CSV/JSON/XML (respecting maxRecordsPerFile)
  ↓ stdout ↓
"Export completed successfully"
```

## Integration Points & Dependencies

- **External**: Only `github.com/spf13/cobra` and `github.com/spf13/pflag` (Cobra's flag package)
- **Test fixtures**: `test/testdata/sample.qif` used by test suite
- **Cross-package imports**: 
  - `cmd/` imports `qifutil/pkg/utils` for `SplitCategoryAndTag()`
  - Most logic in `cmd/` (monolithic for transaction parsing)

## Common Pitfalls

1. **Regex Compilation Errors**: Always check regex syntax; use raw strings with backticks
2. **Line Ending Inconsistency**: QIF files may have mixed line endings; always normalize before regex
3. **Unvalidated Global State**: Some commands depend on globals set by previous operations (e.g., wizard sets before calling transactions command)
4. **Path Separators**: Tests and wizard may run on Windows; use `filepath` package consistently
5. **Empty Account Names**: Account filtering validation should reject empty account names after trim

## Key Files Reference

- **Main command entry**: `cmd/root.go` - Global flags, root command initialization
- **Transaction parsing engine**: `cmd/transactions.go` - Core regex, mapping logic, file splitting (654 lines)
- **Interactive wizard**: `cmd/wizard.go` - User-guided CLI, output capture, account listing (250+ lines)
- **Utilities**: `pkg/utils/category.go`, `pkg/utils/path.go` - Helper functions
- **Tests**: `cmd/transactions_test.go` - Test patterns and test helper usage
- **Testdata**: `test/testdata/sample.qif` - QIF fixture with sample Checking/Credit Card accounts
