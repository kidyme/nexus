package feedback

import "testing"

func TestFeedbackValidateRequiresKeyFields(t *testing.T) {
	err := (Feedback{}).Validate()
	if err != ErrFeedbackTypeRequired {
		t.Fatalf("expected ErrFeedbackTypeRequired, got %v", err)
	}
}

func TestKeyValidateRequiresKeyFields(t *testing.T) {
	err := (Key{}).Validate()
	if err != ErrFeedbackTypeRequired {
		t.Fatalf("expected ErrFeedbackTypeRequired, got %v", err)
	}
}
