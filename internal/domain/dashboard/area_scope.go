package dashboard

import (
	"context"
	"errors"
)

// ErrForbidden is returned when the volunteer lacks the leadership flag required
// for the requested area dashboard.
var ErrForbidden = errors.New("forbidden")

// ScopeResolver is the single reusable scope-resolution component. It resolves
// the logged-in volunteer, enforces the leadership permission, and returns ALL
// volunteer ids under the leader's area. Dashboard services receive only the
// []int and never learn how the scope was determined.
type ScopeResolver struct {
	Repo AreaRepository
}

func NewScopeResolver(repo AreaRepository) *ScopeResolver {
	return &ScopeResolver{Repo: repo}
}

// ResolveHuAiVolunteerIDs authorizes the Hu Ai dashboard (leader or deputy) and
// returns the volunteer ids in the leader's area subtree.
func (s *ScopeResolver) ResolveHuAiVolunteerIDs(ctx context.Context, userID int) ([]int, error) {
	v, err := s.Repo.GetScopeVolunteer(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !v.IsHuAiLeader && !v.IsHuAiDeputy {
		return nil, ErrForbidden
	}
	return s.Repo.AreaVolunteerIDs(ctx, v.MasterAreaID)
}

// ResolveXieLiVolunteerIDs authorizes the Xie Li dashboard (leader or deputy)
// and returns the volunteer ids in the leader's area subtree.
func (s *ScopeResolver) ResolveXieLiVolunteerIDs(ctx context.Context, userID int) ([]int, error) {
	v, err := s.Repo.GetScopeVolunteer(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !v.IsXieLiLeader && !v.IsXieLiDeputy {
		return nil, ErrForbidden
	}
	return s.Repo.AreaVolunteerIDs(ctx, v.MasterAreaID)
}
