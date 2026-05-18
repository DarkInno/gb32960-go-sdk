package auth

import (
	"context"
	"testing"
)

func TestVINWhitelist(t *testing.T) {
	t.Run("authenticate existing", func(t *testing.T) {
		w := NewVINWhitelist([]string{"VIN001", "VIN002", "VIN003"})
		ok, err := w.Authenticate(context.Background(), "VIN001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("expected VIN to be authenticated")
		}
	})

	t.Run("authenticate missing", func(t *testing.T) {
		w := NewVINWhitelist([]string{"VIN001"})
		ok, err := w.Authenticate(context.Background(), "VIN999")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected VIN to be rejected")
		}
	})

	t.Run("empty whitelist", func(t *testing.T) {
		w := NewVINWhitelist(nil)
		ok, err := w.Authenticate(context.Background(), "VIN001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected empty whitelist to reject")
		}
	})

	t.Run("add and authenticate", func(t *testing.T) {
		w := NewVINWhitelist([]string{"VIN001"})
		w.Add("VIN002")
		ok, _ := w.Authenticate(context.Background(), "VIN002")
		if !ok {
			t.Error("expected newly added VIN to be authenticated")
		}
	})

	t.Run("remove and reject", func(t *testing.T) {
		w := NewVINWhitelist([]string{"VIN001", "VIN002"})
		w.Remove("VIN001")
		ok, _ := w.Authenticate(context.Background(), "VIN001")
		if ok {
			t.Error("expected removed VIN to be rejected")
		}
		ok, _ = w.Authenticate(context.Background(), "VIN002")
		if !ok {
			t.Error("expected non-removed VIN to still be authenticated")
		}
	})

	t.Run("remove non-existent", func(t *testing.T) {
		w := NewVINWhitelist([]string{"VIN001"})
		w.Remove("VIN999")
		ok, _ := w.Authenticate(context.Background(), "VIN001")
		if !ok {
			t.Error("expected original VIN to remain")
		}
	})
}

func TestAllowAll(t *testing.T) {
	a := &AllowAll{}

	t.Run("any VIN", func(t *testing.T) {
		ok, err := a.Authenticate(context.Background(), "ANY_VIN_HERE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("AllowAll should always return true")
		}
	})

	t.Run("empty VIN", func(t *testing.T) {
		ok, err := a.Authenticate(context.Background(), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("AllowAll should return true even for empty VIN")
		}
	})
}
