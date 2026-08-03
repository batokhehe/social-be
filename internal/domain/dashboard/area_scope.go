package dashboard

import (
	"context"
	"errors"
)

// ErrForbidden is returned when the volunteer lacks the leadership flag required
// for the requested area dashboard.
var ErrForbidden = errors.New("forbidden")

// Scope is the resolved reach of an area leader.
//
// VolunteerIDs spans the leader's area subtree (root + descendants) and drives
// the people-based aggregates: activities, hours, active volunteers, donors and
// donations.
//
// MasterAreaID is the leader's OWN area only -- no descendants. Expense
// aggregation is deliberately per-area (business rule): expense dashboards show
// only expenses belonging to the logged-in user's master area.
type Scope struct {
	VolunteerIDs []int
	MasterAreaID int
}

// ScopeResolver is the single reusable scope-resolution component. It resolves
// the logged-in volunteer, enforces the leadership permission, and returns the
// leader's area subtree. Dashboard services receive only the Scope and never
// learn how it was determined.
type ScopeResolver struct {
	Repo AreaRepository
}

func NewScopeResolver(repo AreaRepository) *ScopeResolver {
	return &ScopeResolver{Repo: repo}
}

// ResolveHuAiScope authorizes the Hu Ai dashboard (leader or deputy) and returns
// the leader's area subtree scope.
func (s *ScopeResolver) ResolveHuAiScope(ctx context.Context, userID int) (Scope, error) {
	v, err := s.Repo.GetScopeVolunteer(ctx, userID)
	if err != nil {
		return Scope{}, err
	}
	if !v.IsHuAiLeader && !v.IsHuAiDeputy {
		return Scope{}, ErrForbidden
	}
	return s.scopeFor(ctx, v.MasterAreaID)
}

// ResolveXieLiScope authorizes the Xie Li dashboard (leader or deputy) and
// returns the leader's area subtree scope.
func (s *ScopeResolver) ResolveXieLiScope(ctx context.Context, userID int) (Scope, error) {
	v, err := s.Repo.GetScopeVolunteer(ctx, userID)
	if err != nil {
		return Scope{}, err
	}
	if !v.IsXieLiLeader && !v.IsXieLiDeputy {
		return Scope{}, ErrForbidden
	}
	return s.scopeFor(ctx, v.MasterAreaID)
}

func (s *ScopeResolver) scopeFor(ctx context.Context, masterAreaID int) (Scope, error) {
	// People-based aggregates keep spanning the area subtree.
	volunteerIDs, err := s.Repo.AreaVolunteerIDs(ctx, masterAreaID)
	if err != nil {
		return Scope{}, err
	}
	// Expenses are scoped to the leader's own area only (no subtree).
	return Scope{VolunteerIDs: volunteerIDs, MasterAreaID: masterAreaID}, nil
}
