# QIFUTIL
## Overview

QIFUTIL is a utility for exporting financial data from Quicken in QIF format. This tool helps users to easily extract and manage their financial data for use in other applications or for backup purposes.

## Features

- Export transactions from Quicken to QIF format
- Support for multiple accounts
- Easy-to-use command-line interface
- Customizable export options

## Usage

To use QIFUTIL, run the executable with the desired options. 

### Export Account List
To export the list of accounts, use the following command:

```sh
./qifutil export accounts --inputFile "" --output "accounts.csv"
```

### Export Categories List
To export the list of categories, use the following command:

```sh
./qifutil export categories --inputFile "" --output "categories.csv"
```

### Export Payees List
To export the list of payees, use the following command:

```sh
./qifutil export payees --inputFile "" --output "payees.csv"
```

### Export Tags List
To export the list of tags, use the following command:

```sh
./qifutil export tags --inputFile "" --output "tags.csv"
```

### Export Transactions List
To export the transactions, use the following command:

```sh
./qifutil export transactions --inputFile "AllAccounts.QIF" --outputPath "C:\export\\" --categoryMapFile "categories.csv" --accountMapFile "accounts.csv" --payeeMapFile "payees.csv" --tagMapFile "tags.csv" --addTagForImport true
```

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request with your changes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contact

For any questions or issues, please open an issue on GitHub or contact the maintainer at chrisgelhaus@live.com.
