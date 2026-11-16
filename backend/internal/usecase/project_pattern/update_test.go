package projectpattern

import (
	"context"
	"errors"
	dg "server_nesting_optimizer/internal/domain/geometry"
	domain "server_nesting_optimizer/internal/domain/project_pattern"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"server_nesting_optimizer/pkg/apperror"
	"testing"
)

type resourceStub struct {
	ProjectPatternRepository
	item    domain.ProjectPattern
	active  bool
	updates int
	deletes int
}

func (r *resourceStub) GetByID(context.Context, int64, int64, int64) (domain.ProjectPattern, error) {
	return r.item, nil
}
func (r *resourceStub) HasActivePlacements(context.Context, int64, int64, int64) (bool, error) {
	return r.active, nil
}
func (r *resourceStub) Update(_ context.Context, _, _, _ int64, name *string, geometry *dg.Polygon) (domain.ProjectPattern, error) {
	r.updates++
	if name != nil {
		r.item.Name = *name
	}
	if geometry != nil {
		r.item.Geometry = *geometry
	}
	return r.item, nil
}
func (r *resourceStub) Delete(context.Context, int64, int64, int64) error {
	r.deletes++
	return nil
}
func assertCode(t *testing.T, err error, code apperror.Code) {
	t.Helper()
	var app *apperror.AppError
	if !errors.As(err, &app) || app.Code != code {
		t.Fatalf("want %s, got %v", code, err)
	}
}
func TestGeometryAndPlacementRestrictions(t *testing.T) {
	polygon := dg.Polygon{Exterior: dg.Ring{Points: []dg.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}}}}
	repo := &resourceStub{item: domain.ProjectPattern{ID: 1, ProjectID: 1, Geometry: polygon}}
	uc := NewUpdateProjectPatternUseCase(repo, simplefeatures.NewEngine())
	input := UpdateProjectPatternInput{UserID: 1, ProjectID: 1, ProjectPatternID: 1, Geometry: &polygon}
	if _, err := uc.Execute(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	repo.active = true
	_, err := uc.Execute(context.Background(), input)
	assertCode(t, err, apperror.AppCodeConflict)
	if repo.updates != 1 {
		t.Fatal("conflict persisted geometry")
	}
	name := "renamed"
	input.Geometry = nil
	input.Name = &name
	if _, err = uc.Execute(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if repo.item.Name != name {
		t.Fatal("rename not persisted")
	}
	del := NewDeleteProjectPatternUseCase(repo)
	err = del.Execute(context.Background(), DeleteProjectPatternInput{UserID: 1, ProjectID: 1, ProjectPatternID: 1})
	assertCode(t, err, apperror.AppCodeConflict)
	if repo.deletes != 0 {
		t.Fatal("deleted used resource")
	}
	repo.active = false
	invalid := dg.Polygon{}
	input.Geometry = &invalid
	_, err = uc.Execute(context.Background(), input)
	assertCode(t, err, apperror.AppCodeValidation)
	scale := 2.0
	input.Geometry = &polygon
	input.Scale = &scale
	_, err = uc.Execute(context.Background(), input)
	assertCode(t, err, apperror.AppCodeValidation)
}
