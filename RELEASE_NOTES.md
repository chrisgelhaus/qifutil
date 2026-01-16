# QIFUTIL Release Notes

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
