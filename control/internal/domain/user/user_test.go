package user

import "testing"

func TestUserValidateRequiresUserID(t *testing.T) {
	err := (User{}).Validate()
	if err != ErrUserIDRequired {
		t.Fatalf("expected ErrUserIDRequired, got %v", err)
	}
}
