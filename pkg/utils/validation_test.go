package utils

import (
	"testing"
)

func TestValidationTrackerCreate(t *testing.T) {
	validator := NewValidationTracker()
	if validator == nil {
		t.Fatal("NewValidationTracker returned nil")
	}
	if validator.HasWarnings() {
		t.Error("New validator should have no warnings")
	}
}

func TestValidationTrackerMissingPayee(t *testing.T) {
	validator := NewValidationTracker()

	validator.AddMissingPayee()
	if validator.MissingPayees != 1 {
		t.Errorf("Expected 1 missing payee, got %d", validator.MissingPayees)
	}
	if !validator.HasWarnings() {
		t.Error("Should have warnings after adding missing payee")
	}

	validator.AddMissingPayee()
	if validator.MissingPayees != 2 {
		t.Errorf("Expected 2 missing payees, got %d", validator.MissingPayees)
	}
}

func TestValidationTrackerMissingCategory(t *testing.T) {
	validator := NewValidationTracker()

	validator.AddMissingCategory()
	if validator.MissingCategory != 1 {
		t.Errorf("Expected 1 missing category, got %d", validator.MissingCategory)
	}
	if !validator.HasWarnings() {
		t.Error("Should have warnings after adding missing category")
	}
}

func TestValidationTrackerZeroAmount(t *testing.T) {
	validator := NewValidationTracker()

	validator.AddZeroAmount()
	if validator.ZeroAmounts != 1 {
		t.Errorf("Expected 1 zero amount, got %d", validator.ZeroAmounts)
	}
	if !validator.HasWarnings() {
		t.Error("Should have warnings after adding zero amount")
	}
}

func TestValidationTrackerRecordTransaction(t *testing.T) {
	validator := NewValidationTracker()

	validator.RecordTransaction()
	if validator.TotalTransactions != 1 {
		t.Errorf("Expected 1 transaction, got %d", validator.TotalTransactions)
	}
	if validator.TotalProcessed != 1 {
		t.Errorf("Expected 1 processed, got %d", validator.TotalProcessed)
	}

	validator.RecordTransaction()
	validator.RecordTransaction()
	if validator.TotalTransactions != 3 {
		t.Errorf("Expected 3 transactions, got %d", validator.TotalTransactions)
	}
}

func TestValidationTrackerAddDuplicate(t *testing.T) {
	validator := NewValidationTracker()

	// Should not add if count is 1
	validator.AddDuplicate("2025-01-01", "Starbucks", "5.45", 1)
	if len(validator.DuplicateTransactions) != 0 {
		t.Error("Should not add duplicate with count of 1")
	}

	// Should add if count is > 1
	validator.AddDuplicate("2025-01-01", "Starbucks", "5.45", 2)
	if len(validator.DuplicateTransactions) != 1 {
		t.Errorf("Expected 1 duplicate, got %d", len(validator.DuplicateTransactions))
	}

	if !validator.HasWarnings() {
		t.Error("Should have warnings after adding duplicate")
	}

	dup := validator.DuplicateTransactions[0]
	if dup.Date != "2025-01-01" || dup.Payee != "Starbucks" || dup.Amount != "5.45" || dup.Count != 2 {
		t.Error("Duplicate data not stored correctly")
	}
}

func TestValidationTrackerAddUnmatchedData(t *testing.T) {
	validator := NewValidationTracker()

	validator.AddUnmatchedData("payee", "Whole Foods")
	if count, ok := validator.UnmatchedData["payee:Whole Foods"]; !ok || count != 1 {
		t.Error("Unmatched data not recorded correctly")
	}

	validator.AddUnmatchedData("payee", "Whole Foods")
	if count, ok := validator.UnmatchedData["payee:Whole Foods"]; !ok || count != 2 {
		t.Error("Unmatched data count not incremented")
	}

	if !validator.HasWarnings() {
		t.Error("Should have warnings after adding unmatched data")
	}
}

func TestValidationTrackerRecordUnusedMapping(t *testing.T) {
	validator := NewValidationTracker()

	unused := []string{"OldCategory1", "OldCategory2", "OldCategory3"}
	validator.RecordUnusedMapping("category", unused)

	if mappings, ok := validator.UnusedMappings["category"]; !ok || len(mappings) != 3 {
		t.Error("Unused mappings not recorded correctly")
	}

	if !validator.HasWarnings() {
		t.Error("Should have warnings after recording unused mappings")
	}
}

func TestValidationTrackerMultipleWarnings(t *testing.T) {
	validator := NewValidationTracker()

	validator.AddMissingPayee()
	validator.AddMissingCategory()
	validator.AddZeroAmount()
	validator.AddDuplicate("2025-01-01", "Store", "100.00", 2)

	if validator.MissingPayees != 1 {
		t.Errorf("Expected 1 missing payee, got %d", validator.MissingPayees)
	}
	if validator.MissingCategory != 1 {
		t.Errorf("Expected 1 missing category, got %d", validator.MissingCategory)
	}
	if validator.ZeroAmounts != 1 {
		t.Errorf("Expected 1 zero amount, got %d", validator.ZeroAmounts)
	}
	if len(validator.DuplicateTransactions) != 1 {
		t.Errorf("Expected 1 duplicate, got %d", len(validator.DuplicateTransactions))
	}

	if !validator.HasWarnings() {
		t.Error("Should have warnings with multiple issues")
	}
}

func TestValidationTrackerNoWarnings(t *testing.T) {
	validator := NewValidationTracker()

	// Record a transaction but no issues
	validator.RecordTransaction()

	if validator.HasWarnings() {
		t.Error("Should have no warnings with clean transaction")
	}
}

func TestValidationTrackerConcurrency(t *testing.T) {
	validator := NewValidationTracker()

	// Simulate concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			validator.RecordTransaction()
			validator.AddMissingPayee()
			validator.AddMissingCategory()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if validator.TotalTransactions != 10 {
		t.Errorf("Expected 10 transactions, got %d", validator.TotalTransactions)
	}
	if validator.MissingPayees != 10 {
		t.Errorf("Expected 10 missing payees, got %d", validator.MissingPayees)
	}
	if validator.MissingCategory != 10 {
		t.Errorf("Expected 10 missing categories, got %d", validator.MissingCategory)
	}
}
