# QIFUTIL

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
- Selecting your input file (drag & drop supported)
- Choosing where to save the exported files
- Selecting specific accounts by number (e.g., "1,3,5")
- Setting date ranges (optional)
- Choosing output format

4. Or use command-line options for more control:
```sh
# Basic conversion - creates CSV files for each account
qifutil export transactions --inputFile "C:\Users\YourName\Downloads\MyData.QIF" --outputPath "C:\Users\YourName\Documents\Exported\"

# Just see what accounts are in your file
qifutil list accounts --inputFile "C:\Users\YourName\Downloads\MyData.QIF"
```

Need help? Type `qifutil --help` or see the detailed instructions below.

## Overview

QIFUTIL is a utility for exporting financial data from Quicken in QIF format. This tool helps users to easily extract and manage their financial data for use in other applications or for backup purposes.

## Features

- Interactive wizard for easy file conversion
  - Simple drag & drop file selection
  - Easy account selection by number (no typing long account names)
  - Guided process for all options
- Smart file handling
  - Automatic file splitting for Monarch compatibility (5000 records per file)
  - Preserves headers in split files
  - Creates organized output with clear file naming
- Export options
  - CSV format (optimized for Monarch import)
  - JSON format (for technical users)
  - XML format (for system integration)
- Flexible filtering
  - Select specific accounts to export
  - Filter by date range
  - Apply category, account, and payee mappings
- Analysis tools
  - Account statistics and analysis
  - List and explore available accounts
  - Transaction summaries

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
qifutil export transactions --inputFile "AllAccounts.QIF" --outputPath "C:\export\\"
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
Use the `--outputFormat` flag to specify `CSV`, `JSON`, or `XML` (default `CSV`).

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

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request with your changes. Make sure to:

1. Add tests for any new functionality
2. Update documentation as needed
3. Follow the existing code style
4. Verify all tests pass before submitting

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contact

For any questions or issues, please open an issue on GitHub or contact the maintainer at chrisgelhaus@live.com.
