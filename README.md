# QIFUTIL

**Latest Update (v1.8.0):** Major wizard improvements and new balance history export feature! 
- ✨ **New: Balance History Export** - Generate daily balance history files for Monarch Money imports with backward/forward calculation modes
- 🛠️ **Enhanced Wizard** - Improved input validation with reprompting for invalid entries, better date range validation (enforces start ≤ end date), cleaner prompts
- 🐛 **Wizard Bug Fixes** - Fixed yes/no response validation (now accepts y/Y for yes, n/N or Enter for no), removed misleading "drag and drop" references, proper error messages for invalid input

Previous features: Export commands correctly respect the `--outputPath` flag, mapping files fully integrated into wizard, improved category/tag splitting, fixed GitHub Actions build pipeline.

## Quick Start Guide

1. Download the latest `qifutil.exe` from the [releases page](https://github.com/chrisgelhaus/qifutil/releases)
2. In Quicken, export your data:
   - File -> Export -> QIF File
   - Select the accounts you want to export
   - Save the file (e.g., as "MyData.QIF")

3. Run the interactive wizard (recommended for new users):
```sh
qifutil wizard
```
The wizard will guide you through:
- Selecting your input file (enter the file path)
- Choosing where to save the exported files
- Selecting specific accounts by number (e.g., "1,3,5")
- Setting date ranges with validation (YYYY-MM-DD format)
- Optionally generating balance history for Monarch Money
- Choosing output format (CSV/JSON/XML)
- Applying optional mapping files to standardize your data

4. Or use command-line options for more control:
```sh
# Basic conversion - creates CSV files for each account
qifutil transactions --inputFile "C:\Users\YourName\Downloads\MyData.QIF" --outputPath "C:\Users\YourName\Documents\Exported\"

# With account mapping file
qifutil transactions --inputFile "MyData.QIF" --outputPath "export/" --accountMapFile "account_mappings.csv"

# Just see what accounts are in your file
qifutil list accounts --inputFile "C:\Users\YourName\Downloads\MyData.QIF"
```

Need help? Type `qifutil --help` or see the detailed instructions below.

## Overview

QIFUTIL is a utility for exporting financial data from Quicken in QIF format. This tool helps users to easily extract and manage their financial data for use in other applications or for backup purposes.

## Features

- Interactive wizard for easy file conversion
  - Simple file path entry with validation
  - Easy account selection by number (no typing long account names)
  - Guided process for all options with input validation
  - Optional balance history generation for Monarch Money
  - Optional mapping files to standardize data
- Smart file handling
  - Automatic file splitting for Monarch compatibility (5000 records per file)
  - Preserves headers in split files
  - Creates organized output with clear file naming
  - All output respects `--outputPath` for organized exports
- Export options
  - **NEW:** Balance history export (daily account balance tracking)
  - CSV format (optimized for Monarch import)
  - JSON format (for technical users)
  - XML format (for system integration)
- Data transformation with mapping files
  - Map categories to standardized names
  - Rename payees for consistency
  - Standardize account names
  - Transform tag values
- Flexible filtering
  - Select specific accounts to export
  - Filter by date range with validation
  - Apply mappings to transform extracted data
- Analysis tools
  - Account statistics and analysis
  - List and explore available accounts
  - Transaction summaries
- Improved validation
  - Yes/no prompts with clear input handling
  - Date range validation (start date ≤ end date)
  - Balance history with forward/backward calculation modes

## Usage

To use QIFUTIL, run the executable with the desired options. 

### Export Account List
To export the list of accounts, use the following command:

```sh
qifutil export accounts --inputFile "AllAccounts.QIF" --output "accounts.csv"
```
Use the `--outputFormat` flag to specify `CSV`, `JSON`, or `XML` (default `CSV`).

### Export Categories List
To export the list of categories, use the following command:

```sh
qifutil export categories --inputFile "AllAccounts.QIF" --output "categories.csv"
```
Use the `--outputFormat` flag to specify `CSV`, `JSON`, or `XML` (default `CSV`).

### Export Payees List
To export the list of payees, use the following command:

```sh
qifutil export payees --inputFile "AllAccounts.QIF" --output "payees.csv"
```
Use the `--outputFormat` flag to specify `CSV`, `JSON`, or `XML` (default `CSV`).

### Export Tags List
To export the list of tags, use the following command:

```sh
qifutil export tags --inputFile "AllAccounts.QIF" --output "tags.csv"
```
Use the `--outputFormat` flag to specify `CSV`, `JSON`, or `XML` (default `CSV`).

### List Available Accounts
To see all accounts in your QIF file:

```sh
qifutil list accounts --inputFile "AllAccounts.QIF"
```

### View Account Statistics
To get statistics about transactions in your accounts:

```sh
qifutil account-stats --inputFile "AllAccounts.QIF"
```

### Export Transactions
For the easiest experience, use the interactive wizard:
```sh
qifutil wizard
```

Or use command-line options for more control:
```sh
qifutil transactions --inputFile "AllAccounts.QIF" --outputPath "C:\export\\"
```

#### Advanced Export Options
You can customize your transaction export with the following options:

- `--recordsPerFile`: Maximum records per file (default: 5000 for Monarch compatibility)
- `--accounts`: Comma-separated list of accounts to export (e.g., "Checking,Savings")
- `--startDate`: Filter transactions from this date (YYYY-MM-DD)
- `--endDate`: Filter transactions until this date (YYYY-MM-DD)
- `--outputFormat`: Choose CSV (default), JSON, or XML
- `--categoryMapFile`: Map categories using a CSV file
- `--accountMapFile`: Map account names using a CSV file
- `--payeeMapFile`: Map payee names using a CSV file
- `--tagMapFile`: Map tags using a CSV file

The tool automatically:
- Splits large files into chunks of 5000 records for Monarch compatibility
- Adds headers to each split file
- Names files consistently with account and part number
- Shows progress and record ranges for split files

- Filter by accounts:
```sh
qifutil export transactions --inputFile "AllAccounts.QIF" --outputPath "C:\export\\" --accounts "Checking,Savings"
```

- Filter by date range:
```sh
qifutil export transactions --inputFile "AllAccounts.QIF" --outputPath "C:\export\\" --startDate "2025-01-01" --endDate "2025-12-31"
```

- Combine filters and mapping files:
```sh
qifutil export transactions --inputFile "AllAccounts.QIF" --outputPath "C:\export\\" \
    --accounts "Checking" \
    --startDate "2025-01-01" \
    --categoryMapFile "categories.csv" \
    --accountMapFile "accounts.csv" \
    --payeeMapFile "payees.csv" \
    --tagMapFile "tags.csv" \
    --addTagForImport true
```
Use the `--outputFormat` flag to specify `CSV`, `JSON`, `XML`, or `MONARCH` (default `CSV`).

### Export Balance History (NEW in v1.8.0)
Generate daily balance history files for Monarch Money imports. This shows account balance changes over time, useful for migrating historical data.

For the easiest experience, the wizard will ask if you want balance history when exporting transactions. Or use the command:

```sh
# Using current balance (backward calculation from known ending balance)
qifutil export balance-history --inputFile "AllAccounts.QIF" --outputPath "C:\export\\" \
    --accounts "Checking Account" \
    --currentBalance 5000.00

# Using opening balance (forward calculation from known starting balance)
qifutil export balance-history --inputFile "AllAccounts.QIF" --outputPath "C:\export\\" \
    --accounts "Savings Account" \
    --openingBalance 10000.00 \
    --startDate "2025-01-01" \
    --endDate "2025-12-31
```

**Balance History Features:**
- Generates CSV file with `Date,Balance` columns
- One record per day (only for days with transactions)
- Backward calculation: works from a known current balance back through transactions
- Forward calculation: works from a known opening balance forward through transactions
- Respects date filtering (--startDate, --endDate)
- Respects file splitting settings (--recordsPerFile)
- File naming: `{AccountName}_balance_history_1.csv`
- Perfect for visualizing account balance trends in Monarch Money

**Balance Calculation Examples:**
If your current balance is $2500 and you have two transactions (-$45.23, -$35.50):
- Jan 15: $2535.50 (2500 - 45.23)
- Jan 16: $2500.00 (2535.50 - 35.50) ✓

If your opening balance is $2500 with the same transactions:
- Jan 15: $2454.77 (2500 - 45.23)
- Jan 16: $2419.27 (2454.77 - 35.50) ✓

## Output Formats

QIFUTIL supports multiple output formats to suit different use cases:

### Monarch Money Format (Optimized for Import)
For the easiest import into Monarch Money, use the MONARCH format:

```sh
qifutil export transactions --inputFile "data.qif" --outputPath "export/" --outputFormat MONARCH
```

This format uses the recommended column order and structure for Monarch compatibility:
- Date, Merchant, Category, Account, Original Statement, Notes, Amount, Tags

The MONARCH format automatically:
- Splits large files into 5000-record chunks (Monarch's recommended limit)
- Maintains proper headers in each split file
- Preserves all transaction details
- Handles special characters properly

### Generic CSV Format with Custom Columns
The CSV format allows you to select exactly which columns to include in your export:

```sh
# Only export Date, Merchant, and Amount
qifutil export transactions --inputFile "data.qif" --outputPath "export/" \
    --outputFormat CSV --csvColumns "Date,Merchant,Amount"

# Custom column order
qifutil export transactions --inputFile "data.qif" --outputPath "export/" \
    --outputFormat CSV --csvColumns "Merchant,Category,Amount,Date"
```

**Available Columns (in any order):**
- `Date` - Transaction date (YYYY-MM-DD)
- `Merchant` - Payee/merchant name
- `Category` - Transaction category
- `Account` - Account name
- `Original Statement` - Original payee from QIF
- `Notes` - Transaction memo/notes
- `Amount` - Transaction amount
- `Tags` - Tags extracted from category

If `--csvColumns` is not specified, CSV format uses the Monarch Money defaults.

### JSON Format
For technical users and system integration:

```sh
qifutil export transactions --inputFile "data.qif" --outputPath "export/" --outputFormat JSON
```

Produces structured JSON with all transaction details, suitable for:
- REST API imports
- Data processing scripts
- System integration
- Database imports

### XML Format
For enterprise system integration:

```sh
qifutil export transactions --inputFile "data.qif" --outputPath "export/" --outputFormat XML
```

Produces well-formed XML with transaction elements, suitable for:
- Enterprise data systems
- XML-based workflows
- Legacy system imports

## Mapping Files

Mapping files allow you to transform and standardize your financial data during export. Each mapping file is a simple CSV with two columns: the source value and the target (replacement) value.

### Creating Mapping Files

Mapping files use CSV format with two columns (with quotes):
```csv
"source","target"
"Whole Foods","Groceries"
"Trader Joe's","Groceries"
"Gas Station XYZ","Transportation:Fuel"
```

### Supported Mapping Types

1. **Category Mapping** (`--categoryMapFile`)
   - Maps category names extracted from QIF
   - Useful for consolidating varied category names
   - Example: Map "Groceries" and "Food" to a standard "Groceries" category

2. **Payee Mapping** (`--payeeMapFile`)
   - Maps merchant/payee names
   - Useful for standardizing retailer names across transactions
   - Example: Map "WHOLE FOODS MKT #1234" to "Whole Foods Market"

3. **Account Mapping** (`--accountMapFile`)
   - Maps account names to standardized names
   - Useful for renaming accounts for compatibility with other systems
   - Example: Map "USAA CHECKING XX5681" to "Primary Checking"

4. **Tag Mapping** (`--tagMapFile`)
   - Maps tag values extracted from categories
   - QIF categories can include tags in `category/tag` format
   - Example: Map "Work" to "Business:Work"

### Using Mapping Files in the Wizard

When you run `qifutil wizard`, you'll be prompted to optionally provide mapping files:
```
Would you like to apply mapping files to transform data? (y/n): y

Category mapping file (drag and drop, or press Enter to skip): C:\mappings\categories.csv
Payee mapping file (drag and drop, or press Enter to skip): C:\mappings\payees.csv
Account mapping file (drag and drop, or press Enter to skip): C:\mappings\accounts.csv
Tag mapping file (drag and drop, or press Enter to skip): C:\mappings\tags.csv
```

### Using Mapping Files from Command Line

```sh
# Apply category and payee mappings
qifutil export transactions --inputFile "data.qif" --outputPath "export/" \
    --categoryMapFile "mappings/categories.csv" \
    --payeeMapFile "mappings/payees.csv"

# Apply all mapping types
qifutil export transactions --inputFile "data.qif" --outputPath "export/" \
    --categoryMapFile "categories.csv" \
    --payeeMapFile "payees.csv" \
    --accountMapFile "accounts.csv" \
    --tagMapFile "tags.csv"
```

### Mapping File Best Practices

- Keep mapping files in the same directory or a dedicated `mappings/` folder
- Use UTF-8 encoding for the CSV files
- Include quotes around values, especially if they contain commas or special characters
- Test mappings on a small subset first
- Source values are case-sensitive (exact match required)
- Unmapped values pass through unchanged
- Comment your mappings with descriptive source names

## Testing

QIFUTIL includes a comprehensive test suite to ensure reliability and correctness. The test framework consists of:

### Test Structure

```
test/
├── helper.go           # Test utilities and helpers
└── testdata/          # Sample data files for testing
    ├── sample.qif     # Sample QIF file
    └── categories.csv # Sample mapping file
```

### Running Tests

Run all tests:
```sh
go test ./...
```

Run specific package tests with verbose output:
```sh
# Test command implementations
go test -v ./cmd

# Test utility functions
go test -v ./pkg/utils
```

Run a specific test:
```sh
go test -v ./cmd -run TestTransactionsCmd/account_filtering
```

### Writing Tests

The test framework provides helper functions for common testing tasks:

```go
func TestYourFeature(t *testing.T) {
    helper := test.NewHelper(t)
    
    // Create temporary test directory
    tempDir := helper.CreateTempDir()
    
    // Copy test data
    helper.CopyTestData("sample.qif", filepath.Join(tempDir, "input.qif"))
    
    // Capture command output
    output := helper.CaptureOutput(func() {
        // Run your command
    })
    
    // Verify results
    helper.AssertFileExists("expected.csv")
    helper.AssertFileContains("output.csv", "expected content")
    helper.AssertOutputContains(output, "success message")
}
```

### Test Coverage

The test suite covers:
- Basic command functionality
- Account filtering
- Date range filtering
- Multiple output formats (CSV, JSON, XML)
- Error cases and edge conditions
- Utility functions

## Recent Improvements (v1.5.0+)

### ✅ Fixed Export Command Output Paths
**Issue:** Export commands (`accounts`, `categories`, `payees`, `tags`) were ignoring the `--outputPath` flag and creating files in the current working directory.

**Solution:** All export commands now respect the `--outputPath` flag and create output files in the specified directory.

**Impact:** Users can now organize all exports in a single directory using a common `--outputPath` parameter.

```powershell
# Output file is now created at: C:\export\accounts.csv
qifutil export accounts --inputFile "data.qif" --outputPath "C:\export"
```

### ✅ Mapping Files Integration with Wizard
**Enhancement:** The interactive wizard now supports mapping files for data transformation.

**Features:**
- Prompts for optional mapping files (categories, payees, accounts, tags)
- Full drag-and-drop support
- Automatic path handling and validation
- Clear summary showing which mappings will be applied

**How to Use:**
```powershell
qifutil wizard

# When prompted:
# Would you like to apply mapping files to transform data? (y/n): y
# Category mapping file (drag and drop, or press Enter to skip): C:\mappings\categories.csv
# ... etc ...
```

### ✅ Fixed Category/Tag Splitting for Multi-Slash Tags
**Issue:** The `SplitCategoryAndTag()` function was failing when tags contained multiple slashes (e.g., `Food:Groceries/Monthly/Extra`).

**Solution:** Updated the splitting logic to only split on the first `/`, preserving all subsequent slashes in the tag portion.

**Example:**
```
Input: "Food:Groceries/Monthly/Extra"
Output: category="Food:Groceries", tag="Monthly/Extra"
```

**Impact:** Tags with nested values (like `/`) are now correctly preserved during data transformation.

### ✅ Fixed GitHub Actions Build Pipeline
**Issue:** Tests were failing on GitHub Actions (Ubuntu) due to:
1. Syntax error in test helper (missing newline)
2. Test data path resolution issues
3. Failing unit test for category splitting

**Solution:** 
- Fixed syntax errors in `test/helper.go`
- Consolidated testdata to single location: `test/testdata/`
- Fixed category splitting logic to handle all test cases
- Improved test robustness

**Impact:** Automated builds on GitHub Actions now pass all tests and build successfully on Linux.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request with your changes. Make sure to:

1. Add tests for any new functionality
2. Update documentation as needed
3. Follow the existing code style
4. Verify all tests pass: `go test -v ./...`
5. Build successfully: `go build -v ./...`
4. Verify all tests pass before submitting

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contact

For any questions or issues, please open an issue on GitHub or contact the maintainer at chrisgelhaus@live.com.
