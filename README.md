# QIFUTIL
## Overview

QIFUTIL is a utility for exporting financial data from Quicken in QIF format. This tool helps users to easily extract and manage their financial data for use in other applications or for backup purposes.

## Features

- Export transactions from Quicken to QIF format
- Support for multiple accounts
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

### Export Transactions List
To export the transactions, use the following command:

```sh
qifutil export transactions --inputFile "AllAccounts.QIF" --outputPath "C:\export\\" --categoryMapFile "categories.csv" --accountMapFile "accounts.csv" --payeeMapFile "payees.csv" --tagMapFile "tags.csv" --addTagForImport true
```
Use the `--outputFormat` flag to specify `CSV`, `JSON`, or `XML` (default `CSV`).

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request with your changes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contact

For any questions or issues, please open an issue on GitHub or contact the maintainer at chrisgelhaus@live.com.
