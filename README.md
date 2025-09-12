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

## Usage

To use QIFUTIL, run the executable with the desired options. 

### Export Account List
To export the list of accounts, use the following command:

```sh
qifutil export accounts --inputFile "AllAccounts.QIF" --output "accounts.csv"
```

### Export Categories List
To export the list of categories, use the following command:

```sh
qifutil export categories --inputFile "AllAccounts.QIF" --output "categories.csv"
```

### Export Payees List
To export the list of payees, use the following command:

```sh
qifutil export payees --inputFile "AllAccounts.QIF" --output "payees.csv"
```

### Export Tags List
To export the list of tags, use the following command:

```sh
qifutil export tags --inputFile "AllAccounts.QIF" --output "tags.csv"
```

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

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request with your changes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contact

For any questions or issues, please open an issue on GitHub or contact the maintainer at chrisgelhaus@live.com.
