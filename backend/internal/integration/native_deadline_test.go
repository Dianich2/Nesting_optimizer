package integration

import (
	"context"
	"github.com/google/uuid"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	optimizer "server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/pkg/apperror"
	"sync/atomic"
	"testing"
	"time"
)

type countedNativeBaseline struct{ calls atomic.Int64 }

func (n *countedNativeBaseline) Optimize(ctx context.Context, p optimizer.Problem) (optimizer.Result, error) {
	n.calls.Add(1)
	return optimizer.NewBaselineOptimizer(sf.NewEngine()).Optimize(ctx, p)
}
func TestNativeDeadlinePreservesSavedLayout(t *testing.T) {
	f := newNestingFixture(t)
	if _, err := f.db.Exec(`UPDATE project_surfaces SET geometry=ST_GeomFromText('POLYGON((0 0,10000 0,10000 10000,0 10000,0 0))') WHERE id=1;
 INSERT INTO placements(id,project_surface_id,project_pattern_id,x,y,rotation) VALUES(99,1,1,20,20,0)`); err != nil {
		t.Fatal(err)
	}
	native := &countedNativeBaseline{}
	uc := newFixtureUseCase(t, f.scoped, native)
	input := f.input
	input.Patterns[0].Quantity = 500
	input.AllowedRotations = make([]float64, 36)
	for i := range input.AllowedRotations {
		input.AllowedRotations[i] = float64(i) * 10
	}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	key := uuid.NewString()
	_, err := uc.ExecuteWithKey(ctx, input, key)
	expectCode(t, err, apperror.AppCodeTimeout)
	if native.calls.Load() != 1 {
		t.Fatal("deadline did not exercise real native optimizer")
	}
	f.counts(t, 1, 0)
	var id int
	if err = f.db.Get(&id, "SELECT id FROM placements"); err != nil || id != 99 {
		t.Fatal("native timeout changed old layout")
	}
	_, err = uc.ExecuteWithKey(t.Context(), input, key)
	expectCode(t, err, apperror.AppCodeTimeout)
	if native.calls.Load() != 1 {
		t.Fatal("terminal timeout replay computed again")
	}
}
