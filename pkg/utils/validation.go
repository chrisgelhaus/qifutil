package utils

import (
	"fmt"
	"sync"
)

// ValidationTracker tracks data quality warnings during export
type ValidationTracker struct {
	mu sync.Mutex

	// Transaction issues
	MissingPayees   int
	MissingCategory int
	ZeroAmounts     int

	// Duplicates (same date, payee, amount)
	DuplicateTransactions []DuplicateWarning

	// Mapping issues
	UnusedMappings map[string][]string // mapping type -> list of unused values
	UnmatchedData  map[string]int      // payee/category -> count of times it appeared unmapped

	// Stats
	TotalTransactions int
	TotalProcessed    int
}

// DuplicateWarning represents a potential duplicate transaction
type DuplicateWarning struct {
	Date   string
	Payee  string
	Amount string
	Count  int // How many duplicates found
}

// NewValidationTracker creates a new validation tracker
func NewValidationTracker() *ValidationTracker {
	return &ValidationTracker{
		DuplicateTransactions: make([]DuplicateWarning, 0),
		UnusedMappings:        make(map[string][]string),
		UnmatchedData:         make(map[string]int),
	}
}

// AddMissingPayee records a transaction with missing payee
func (vt *ValidationTracker) AddMissingPayee() {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	vt.MissingPayees++
}

// AddMissingCategory records a transaction with missing category
func (vt *ValidationTracker) AddMissingCategory() {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	vt.MissingCategory++
}

// AddZeroAmount records a transaction with zero amount
func (vt *ValidationTracker) AddZeroAmount() {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	vt.ZeroAmounts++
}

// RecordTransaction increments transaction counter
func (vt *ValidationTracker) RecordTransaction() {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	vt.TotalTransactions++
	vt.TotalProcessed++
}

// AddUnmatchedData records data that wasn't in mapping
func (vt *ValidationTracker) AddUnmatchedData(dataType, value string) {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	key := dataType + ":" + value
	vt.UnmatchedData[key]++
}

// RecordUnusedMapping records mapping values that were never used
func (vt *ValidationTracker) RecordUnusedMapping(mappingType string, values []string) {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	if len(values) > 0 {
		vt.UnusedMappings[mappingType] = values
	}
}

// AddDuplicate records a potential duplicate
func (vt *ValidationTracker) AddDuplicate(date, payee, amount string, count int) {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	if count > 1 {
		vt.DuplicateTransactions = append(vt.DuplicateTransactions, DuplicateWarning{
			Date:   date,
			Payee:  payee,
			Amount: amount,
			Count:  count,
		})
	}
}

// hasWarningsUnlocked checks for warnings without acquiring the lock
// Must only be called when the lock is already held
func (vt *ValidationTracker) hasWarningsUnlocked() bool {
	return vt.MissingPayees > 0 ||
		vt.MissingCategory > 0 ||
		vt.ZeroAmounts > 0 ||
		len(vt.DuplicateTransactions) > 0 ||
		len(vt.UnusedMappings) > 0 ||
		len(vt.UnmatchedData) > 0
}

// HasWarnings returns true if there are any warnings to report
func (vt *ValidationTracker) HasWarnings() bool {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	return vt.hasWarningsUnlocked()
}

// PrintSummary prints a validation summary
func (vt *ValidationTracker) PrintSummary() {
	vt.mu.Lock()
	defer vt.mu.Unlock()

	if !vt.hasWarningsUnlocked() {
		fmt.Println("\n✓ Data validation: No issues found")
		return
	}

	fmt.Println("\n⚠️  Data Validation Summary:")
	fmt.Println("=============================")

	if vt.MissingPayees > 0 {
		fmt.Printf("  • Missing payees: %d transactions\n", vt.MissingPayees)
	}

	if vt.MissingCategory > 0 {
		fmt.Printf("  • Missing categories: %d transactions\n", vt.MissingCategory)
	}

	if vt.ZeroAmounts > 0 {
		fmt.Printf("  • Zero amounts: %d transactions\n", vt.ZeroAmounts)
	}

	if len(vt.DuplicateTransactions) > 0 {
		fmt.Printf("  • Potential duplicates: %d groups detected\n", len(vt.DuplicateTransactions))
		for i, dup := range vt.DuplicateTransactions {
			if i < 5 { // Show first 5
				fmt.Printf("    - %s | %s | %s (%d times)\n", dup.Date, dup.Payee, dup.Amount, dup.Count)
			}
		}
		if len(vt.DuplicateTransactions) > 5 {
			fmt.Printf("    ... and %d more\n", len(vt.DuplicateTransactions)-5)
		}
	}

	if len(vt.UnusedMappings) > 0 {
		for mappingType, values := range vt.UnusedMappings {
			fmt.Printf("  • %s mapping: %d rules never used\n", mappingType, len(values))
			for i, val := range values {
				if i < 3 {
					fmt.Printf("    - \"%s\"\n", val)
				}
			}
			if len(values) > 3 {
				fmt.Printf("    ... and %d more\n", len(values)-3)
			}
		}
	}

	if len(vt.UnmatchedData) > 0 {
		unmatchedCount := len(vt.UnmatchedData)
		fmt.Printf("  • Unmapped values: %d different payees/categories not in mapping files\n", unmatchedCount)
	}

	fmt.Println("=============================")
	fmt.Printf("Note: Review validation.log for full details\n\n")
}
