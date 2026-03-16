package item

import "testing"

func TestItemValidateRequiresItemID(t *testing.T) {
	err := (Item{}).Validate()
	if err != ErrItemIDRequired {
		t.Fatalf("expected ErrItemIDRequired, got %v", err)
	}
}
