# QIFUTIL
## Overview

QIFUTIL is a utility for exporting financial data from Quicken in QIF format. This tool helps users to easily extract and manage their financial data for use in other applications or for backup purposes.

## Features

- Export transactions from Quicken to QIF format
- Support for multiple accounts with filtering
- Date range filtering for transactions
- Account statistics and analysis
- List and explore available accounts
- Easy-to-use command-line interface
- Customizable export options
 - Supports CSV, JSON, and XML output formats

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

### Export Transactions List
To export the transactions, use the following command:

```sh
qifutil export transactions --inputFile "AllAccounts.QIF" --outputPath "C:\export\\"
```

#### Advanced Export Options
You can customize your transaction export with the following options:

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
