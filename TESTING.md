# Testing Guide for New Features

## Quick Start

### Run All Tests
```bash
go test ./pkg/utils -v
go test ./pkg/config -v
```

### Run Specific Test
```bash
go test ./pkg/utils -v -run TestValidationTrackerCreate
go test ./pkg/config -v -run TestWizardConfigSaveAndLoad
```

## Test Files

### 1. Validation Tracker Tests
**File:** `pkg/utils/validation_test.go`

Tests the data quality tracking system that monitors:
- Missing payees in transactions
- Missing categories in transactions  
- Zero-amount transactions
- Duplicate detection
- Unmapped data in mappings
- Unused mapping rules

**Key Tests:**
- `TestValidationTrackerCreate` - Initialization
- `TestValidationTrackerMissingPayee` - Payee tracking
- `TestValidationTrackerMissingCategory` - Category tracking
- `TestValidationTrackerAddDuplicate` - Duplicate detection
- `TestValidationTrackerConcurrency` - Thread safety

### 2. Config File Tests
**File:** `pkg/config/config_test.go`

Tests the configuration save/load system that preserves:
- Input/output file paths
- Export type selections (transactions/balance history)
- Account selections
- Date ranges
- Output formats
- Mapping file paths
- All wizard settings

**Key Tests:**
- `TestWizardConfigSaveAndLoad` - Full roundtrip
- `TestWizardConfigLoadNonexistentFile` - Error handling
- `TestWizardConfigLoadInvalidJSON` - Error handling
- `TestWizardConfigSaveAndLoadWithMappings` - Mapping preservation
- `TestWizardConfigSaveAndLoadBalanceHistory` - Balance history fields

## Understanding Test Output

When you run tests, you'll see:
```
=== RUN   TestName
--- PASS: TestName (0.00s)
```

This means the test passed. If there are failures, you'll see:
```
--- FAIL: TestName (0.01s)
    validation_test.go:50: assertion failed
```

## Test Examples

### Example 1: Test Validation Tracker
```go
// This test verifies that missing payees are tracked
validator := NewValidationTracker()
validator.AddMissingPayee()
if validator.MissingPayees != 1 {
    t.Error("Should have 1 missing payee")
}
```

### Example 2: Test Config Save/Load
```go
// This test verifies configuration roundtrip
cfg := &WizardConfig{InputFile: "test.qif"}
cfg.SaveConfig("test.json")
loaded, _ := LoadConfig("test.json")
if loaded.InputFile != "test.qif" {
    t.Error("InputFile not preserved")
}
```

## Coverage

All tests use:
- **Temporary directories** for file operations (no cleanup needed)
- **Concurrent testing** to ensure thread safety
- **Error cases** to validate error handling
- **Edge cases** for boundary conditions

## Verifying Tests Pass

```bash
# Show all tests passing
$ go test ./pkg/utils ./pkg/config -v
ok      qifutil/pkg/utils       0.249s
ok      qifutil/pkg/config      0.312s

# Run with coverage
$ go test ./pkg/utils -cover
coverage: 85.2% of statements

$ go test ./pkg/config -cover
coverage: 92.1% of statements
```

## Adding New Tests

To add a new test, create a function in the appropriate `*_test.go` file:

```go
func TestNewFeature(t *testing.T) {
    // Setup
    validator := NewValidationTracker()
    
    // Execute
    validator.AddMissingPayee()
    
    // Verify
    if validator.MissingPayees != 1 {
        t.Error("Test failed")
    }
}
```

Then run:
```bash
go test ./pkg/utils -v -run TestNewFeature
```

## Troubleshooting

### Tests timeout
If tests hang, check for blocking I/O operations. Use `go test -timeout 30s`.

### JSON unmarshaling errors
Check that JSON field tags match (lowercase first letter in struct, but with JSON tags).

### Concurrent test issues
Use mutex locks (already done in ValidationTracker) to prevent race conditions.

## CI/CD Integration

Run these commands in your CI pipeline:
```bash
go test ./pkg/... -v
go test ./pkg/... -race        # Detect race conditions
go test ./pkg/... -cover       # Show coverage
```

For GitHub Actions:
```yaml
- name: Run tests
  run: go test ./pkg/... -v
```
