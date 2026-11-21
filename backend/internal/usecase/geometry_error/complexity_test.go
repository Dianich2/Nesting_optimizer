package geometryerror

import (
	"errors"
	"fmt"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
	"testing"
)

func TestComplexityPreservesCauseWithoutExposingDetails(t *testing.T) {
	e := Complexity(fmt.Errorf("private fixture detail: %w", geometry.ErrComplexity))
	if !errors.Is(e, geometry.ErrComplexity) || e.Code != apperror.AppCodeValidation || e.Message != "geometry complexity budget exceeded" {
		t.Fatal("complexity contract changed")
	}
}

func TestRestoredFailureDiscriminatesComplexity(t *testing.T) {
	e := Complexity(geometry.ErrComplexity)
	e.Err = nil
	if !errors.Is(Restore(e), geometry.ErrComplexity) {
		t.Fatal("persisted complexity lost")
	}
	other := apperror.Validation("geometry complexity budget exceeded", apperror.NewFieldError("quantity", apperror.FieldCodeTooLong, "invalid"))
	if errors.Is(Restore(other), geometry.ErrComplexity) {
		t.Fatal("unrelated validation became complexity")
	}
}
