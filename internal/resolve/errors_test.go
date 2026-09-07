package resolve

import "testing"

func TestCycleDetectionErrorError(t *testing.T) {
	t.Run("joins path with arrows", func(t *testing.T) {
		err := &CycleDetectionError{path: []string{"var:a", "var:b", "var:a"}}
		want := "cycle detected: var:a -> var:b -> var:a"
		if got := err.Error(); got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("single element path", func(t *testing.T) {
		err := &CycleDetectionError{path: []string{"cmd:x"}}
		want := "cycle detected: cmd:x"
		if got := err.Error(); got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("empty path", func(t *testing.T) {
		err := &CycleDetectionError{}
		want := "cycle detected: "
		if got := err.Error(); got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}
