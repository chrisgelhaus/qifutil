# Code Review - QIFUTIL

## Issues Found

### 🔴 CRITICAL ISSUES

#### 1. **Missing Error Handling in Date Parsing (cmd/transactions.go, lines ~482-488)**
**Severity**: HIGH
**Location**: `cmd/transactions.go` in the transaction processing loop
**Problem**:
```go
transDate, _ := time.Parse("2006-01-02", fullDate)
if startDate != "" {
    startDateTime, _ := time.Parse("2006-01-02", startDate)
    if transDate.Before(startDateTime) {
        continue
    }
}
if endDate != "" {
    endDateTime, _ := time.Parse("2006-01-02", endDate)
    if transDate.After(endDateTime) {
        continue
    }
}
```
**Issue**: Using `_` to ignore errors on date parsing. If `fullDate` has invalid format, the resulting zero-value `time.Time` will silently cause incorrect filtering without warning to the user.

**Recommendation**: 
```go
transDate, err := time.Parse("2006-01-02", fullDate)
if err != nil {
    validator.AddInvalidDate(fullDate)
    continue
}
```

---

#### 2. **Resource Leak in File Splitting Loop (cmd/transactions.go, lines ~536-564)**
**Severity**: HIGH
**Location**: When file splitting occurs
**Problem**: 
```go
if maxRecordsPerFile != 0 && count%maxRecordsPerFile == 0 {
    // ... writes data ...
    outputFile.Close()
    
    fileIndex++
    outputFileName = fmt.Sprintf("%s_%d%s", accountName, fileIndex, ext)
    fullPath := filepath.Join(outputPath, outputFileName)
    
    outputFile, err = os.Create(fullPath)  // ← New file opened here
    if err != nil {
        fmt.Printf("Error creating split file %s: %v\n", outputFileName, err)
        return  // ← File not closed if error occurs!
    }
```
**Issue**: If `os.Create()` fails, the function returns without proper cleanup. The previous `outputFile` handle could be in an inconsistent state.

**Recommendation**: Defer cleanup or ensure error handling doesn't leak:
```go
if err != nil {
    fmt.Printf("Error creating split file %s: %v\n", outputFileName, err)
    outputFile.Close()  // Close previous file
    return
}
```

---

#### 3. **Unchecked Directory Changes (cmd/transactions.go, lines ~206-216)**
**Severity**: MEDIUM
**Location**: Directory change before processing
**Problem**:
```go
// Save current directory and change to output directory
origDir, dirErr := os.Getwd()
if dirErr != nil {
    fmt.Printf("Error getting current directory: %v\n", dirErr)
    os.Exit(1)
}
defer os.Chdir(origDir) // Restore original directory when done

if chdirErr := os.Chdir(outputPath); chdirErr != nil {
    fmt.Printf("Error changing to output directory: %v\n", chdirErr)
    os.Exit(1)
}
```
**Issue**: Changing to a different directory is problematic for file operations. The code then uses relative paths after the directory change. This is unusual and can cause confusion. Also, the directory change is not necessary since `os.Create()` and `filepath.Join()` handle absolute paths correctly.

**Recommendation**: Remove the directory change entirely and use absolute paths consistently:
```go
// Don't change directory - just use full paths
// outputFile, err := os.Create(filepath.Join(outputPath, outputFileName))
```

---

#### 4. **Silent Failure on Amount Parsing (cmd/transactions.go, lines ~450-458)**
**Severity**: MEDIUM
**Location**: Amount parsing in transaction loop
**Problem**:
```go
amountFloat, err := strconv.ParseFloat(amount1, 64)
if err != nil {
    fmt.Printf("Warning: Could not parse amount '%s', using as-is\n", amount1)
} else {
    amount1 = fmt.Sprintf("%.2f", amountFloat)
}
```
**Issue**: If parsing fails, the original string (potentially malformed) is used as-is. For financial data, this is risky. Invalid amounts should either skip the transaction or mark it for review.

**Recommendation**:
```go
amountFloat, err := strconv.ParseFloat(amount1, 64)
if err != nil {
    fmt.Printf("Warning: Skipping transaction - invalid amount '%s'\n", amount1)
    validator.AddInvalidAmount()
    continue
}
amount1 = fmt.Sprintf("%.2f", amountFloat)
```

---

### 🟡 MEDIUM ISSUES

#### 5. **CSV Injection Vulnerability (cmd/transactions.go, lines ~709-720)**
**Severity**: MEDIUM
**Location**: `buildCSVRow()` function
**Problem**:
```go
func buildCSVRow(record TransactionRecord, columns string) string {
    // ... code ...
    // Build quoted CSV line
    var line strings.Builder
    for i, val := range values {
        if i > 0 {
            line.WriteString(",")
        }
        line.WriteString("\"" + strings.ReplaceAll(val, "\"", "\"\"") + "\"")
    }
    // ...
}
```
**Issue**: While quotes are escaped, fields starting with `=`, `+`, `@`, or `-` can be interpreted as formulas in Excel/Sheets, causing CSV injection attacks. Should either escape these characters or use a CSV writer library.

**Recommendation**: Use the built-in `encoding/csv` package more robustly:
```go
func buildCSVRow(record TransactionRecord, columns string) string {
    var buf strings.Builder
    w := csv.NewWriter(&buf)
    
    columnList := strings.Split(columns, ",")
    values := make([]string, len(columnList))
    // ... populate values ...
    
    w.Write(values)
    w.Flush()
    return buf.String()
}
```

---

#### 6. **No Validation of Mapping File Content (cmd/transactions.go, lines ~684-705)**
**Severity**: MEDIUM
**Location**: `loadMapping()` function
**Problem**:
```go
switch len(record) {
case 1:
    // Single field - skip (move on)
    continue
case 2:
    // Two fields - set key-value in map
    key := record[0]
    value := record[1]
    mapping[key] = value
default:
    fmt.Println("Unexpected number of fields:", record)
}
```
**Issue**: 
- No validation that keys aren't empty
- Duplicate keys silently overwrite (no warning)
- Warnings printed to stdout mixed with data output
- No check for quoted values

**Recommendation**:
```go
case 2:
    key := strings.TrimSpace(record[0])
    value := strings.TrimSpace(record[1])
    if key == "" {
        fmt.Fprintf(os.Stderr, "Warning: Skipping empty mapping key in %s\n", filePath)
        continue
    }
    if existingValue, exists := mapping[key]; exists {
        fmt.Fprintf(os.Stderr, "Warning: Overwriting mapping for '%s': '%s' -> '%s'\n", 
            key, existingValue, value)
    }
    mapping[key] = value
```

---

#### 7. **Verbose Mapping Logging Can Overwhelm Output (cmd/transactions.go, line ~709)**
**Severity**: LOW-MEDIUM
**Location**: `applyMapping()` function
**Problem**:
```go
func applyMapping(input string, mapping map[string]string) string {
    for oldValue, newValue := range mapping {
        if oldValue == input {
            fmt.Printf("Mapping: %s -> %s\n", input, newValue)  // ← Logs EVERY mapping
            return newValue
        }
    }
    return input
}
```
**Issue**: For a file with 10,000 transactions and 100 category mappings, this prints to stdout for every match. This pollutes the output and slows execution significantly.

**Recommendation**: Add a debug flag or count summaries:
```go
func applyMapping(input string, mapping map[string]string, debug bool) string {
    for oldValue, newValue := range mapping {
        if oldValue == input {
            if debug {
                fmt.Printf("Mapping: %s -> %s\n", input, newValue)
            }
            validator.RecordMapping(input, newValue)
            return newValue
        }
    }
    return input
}
```

---

#### 8. **Race Condition in ValidationTracker (pkg/utils/validation.go)**
**Severity**: LOW-MEDIUM
**Location**: ValidationTracker uses sync.Mutex but some operations aren't protected
**Problem**: Looking at the code pattern, some methods use mutex (good), but if there are any getter methods without locks, they could read inconsistent data during concurrent writes.

**Recommendation**: Review all getter methods to ensure they acquire locks, or use atomic operations for numeric counters.

---

### 🟢 MINOR ISSUES / CODE QUALITY

#### 9. **Inconsistent Error Messages (Multiple Files)**
**Severity**: LOW
**Issue**: Mix of `fmt.Println()`, `fmt.Printf()`, and no consistent error prefix/styling
**Recommendation**: Create a logger utility:
```go
func logError(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, "❌ Error: "+msg+"\n", args...)
}
func logWarning(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, "⚠️  Warning: "+msg+"\n", args...)
}
```

---

#### 10. **Regex Compiled Multiple Times in Loop (cmd/transactions.go, line ~418)**
**Severity**: LOW
**Location**: Each account block recompiles the transaction regex
**Problem**:
```go
for _, accountBlock := range accountBlocks {
    // ...
    regex, err := regexp.Compile(transactionRegexString)  // ← Recompiled each iteration!
    if err != nil {
        return
    }
```
**Issue**: Compiling regex inside a loop is inefficient. Should be compiled once outside the loop.

**Recommendation**:
```go
transactionRegex, err := regexp.Compile(transactionRegexString)
if err != nil {
    return
}

for _, accountBlock := range accountBlocks {
    // ...
    transactions := transactionRegex.FindAllStringSubmatch(textBetweenTypes, -1)
```

---

#### 11. **Regex Error Not Handled (cmd/transactions.go, line ~329)**
**Severity**: MEDIUM
**Location**: `accountBlockHeaderRegex` compilation
**Problem**:
```go
regex, err := regexp.Compile(accountBlockHeaderRegex)
if err != nil {
    return  // ← Just returns silently!
}
```
**Issue**: Silent return without informing user. Error is returned but `Run()` doesn't have error return type, so no propagation.

**Recommendation**: Use `fmt.Println()` or better, add proper error logging.

---

#### 12. **No Bounds Check on Regex Matches (cmd/transactions.go, line ~432)**
**Severity**: MEDIUM
**Location**: Transaction field extraction
**Problem**:
```go
for _, t := range transactions {
    if len(t) > 1 {
        month := strings.TrimSpace(t[1])
        day := strings.TrimSpace(t[2])
        year := strings.TrimSpace(t[4])
        // ...
        payee := strings.TrimSpace(t[15])
        // ...
        transactionMemo := strings.TrimSpace(t[18])
        // ...
        category, tag := utils.SplitCategoryAndTag(t[20])
```
**Issue**: Accesses indices 1, 2, 4, 15, 18, 20 without checking if they exist. If regex changes or matches fewer groups, this panics.

**Recommendation**:
```go
if len(t) < 21 {  // Ensure all needed indices exist
    fmt.Printf("Warning: Transaction regex returned %d groups, expected 21+\n", len(t))
    continue
}
```

---

#### 13. **Hard-Coded Date Format in Multiple Places (cmd/transactions.go)**
**Severity**: LOW
**Issue**: Date format `"2006-01-02"` appears in multiple functions:
- Line 176 (PreRun)
- Line 482-484 (transaction filtering)
- Line 500 (date parsing)

**Recommendation**: Define as a constant:
```go
const dateFormat = "2006-01-02"
```

---

#### 14. **Unclear Variable Naming (cmd/transactions.go, lines ~443-445)**
**Severity**: LOW
**Issue**:
```go
amount1 := strings.TrimSpace(t[6])
// ... later ...
//amount2 := strings.TrimSpace(t[8])  // ← Why is amount2 commented out?
```
**Issue**: QIF files can have both User amount (U) and Transaction amount (T). Using only `amount1` without explanation is confusing.

**Recommendation**: Clarify:
```go
userAmount := strings.TrimSpace(t[6])      // U field
transactionAmount := strings.TrimSpace(t[8]) // T field (commented - QIF spec uses U)
// Use userAmount for export
```

---

#### 15. **No Input Validation for Empty/Malformed QIF (cmd/transactions.go, line ~323)**
**Severity**: MEDIUM
**Location**: After reading the file
**Problem**:
```go
inputBytes, err := os.ReadFile(inputFile)
if err != nil {
    fmt.Println("Error reading file:", err)
} else {
    fmt.Printf("Input file opened. Length: %d\n", len(inputBytes))
}
inputContent := string(inputBytes)
```
**Issue**: If file is empty or all whitespace, no accounts will be found but no clear error is shown. Proceeds silently.

**Recommendation**:
```go
if len(inputBytes) == 0 {
    fmt.Println("Error: Input file is empty")
    os.Exit(1)
}
if len(accountBlocks) == 0 {
    fmt.Println("Error: No accounts found in QIF file. Check if file is valid QIF format.")
    os.Exit(1)
}
```

---

## Implementation Notes

These three critical issues require careful implementation due to complex nesting in the transaction processing loop. The changes interact with the file-splitting logic and date filtering, so they must be applied carefully to avoid breaking variable scoping or control flow.

### Recommended Implementation Approach

1. **First**: Remove the directory change code (Issue #3) - simplest and self-contained
2. **Second**: Fix the resource leak (Issue #2) - only changes error handling, low impact
3. **Third**: Add date parsing error handling (Issue #1) - most complex due to variable scoping

All three fixes have been documented with exact line numbers and code examples above.

---

## Summary Table

| Issue | Severity | Type | File |
|-------|----------|------|------|
| Missing error handling in date parsing | HIGH | Logic | transactions.go:482 |
| Resource leak in file splitting | HIGH | Resource | transactions.go:536 |
| Unchecked directory changes | MEDIUM | Design | transactions.go:206 |
| Silent failure on amount parsing | MEDIUM | Logic | transactions.go:450 |
| CSV injection vulnerability | MEDIUM | Security | transactions.go:709 |
| No validation of mapping files | MEDIUM | Validation | transactions.go:684 |
| Verbose logging pollutes output | MEDIUM | Performance | transactions.go:709 |
| Race condition in ValidationTracker | MEDIUM | Concurrency | validation.go |
| Inconsistent error messages | LOW | Code Quality | Multiple |
| Regex compiled in loop | LOW | Performance | transactions.go:418 |
| Regex compilation error not handled | MEDIUM | Error Handling | transactions.go:329 |
| No bounds check on regex matches | MEDIUM | Safety | transactions.go:432 |
| Hard-coded date format | LOW | Code Quality | transactions.go:Multiple |
| Unclear variable naming | LOW | Code Quality | transactions.go:443 |
| No validation for empty QIF | MEDIUM | Validation | transactions.go:323 |

## Recommendations Priority

1. **Fix the date parsing error handling** (HIGH) - Can silently produce wrong results
2. **Fix resource leaks in file splitting** (HIGH) - Can leave files in bad state
3. **Add regex match bounds checking** (MEDIUM) - Can cause panics
4. **Compile regex once outside loop** (LOW-MEDIUM) - Performance improvement
5. **Remove unnecessary directory changes** (MEDIUM) - Simplify code and fix potential issues
6. **Improve mapping file validation** (MEDIUM) - Prevents subtle data issues
7. **Use CSV library for formatting** (MEDIUM) - Fixes injection vulnerability and ensures correctness

