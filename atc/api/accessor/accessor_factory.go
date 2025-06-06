package accessor

import (
	"fmt"
	"net/http"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/db"
)

//counterfeiter:generate . TokenVerifier
type TokenVerifier interface {
	Verify(req *http.Request) (map[string]any, error)
}

//counterfeiter:generate . TeamFetcher
type TeamFetcher interface {
	GetTeams() ([]db.Team, error)
}

func NewAccessFactory(
	tokenVerifier TokenVerifier,
	teamFetcher TeamFetcher,
	systemClaimKey string,
	systemClaimValues []string,
	displayUserIdGenerator atc.DisplayUserIdGenerator,
) AccessFactory {
	return &accessFactory{
		tokenVerifier:          tokenVerifier,
		teamFetcher:            teamFetcher,
		systemClaimKey:         systemClaimKey,
		systemClaimValues:      systemClaimValues,
		displayUserIdGenerator: displayUserIdGenerator,
	}
}

type accessFactory struct {
	tokenVerifier          TokenVerifier
	teamFetcher            TeamFetcher
	systemClaimKey         string
	systemClaimValues      []string
	displayUserIdGenerator atc.DisplayUserIdGenerator
}

func (a *accessFactory) Create(req *http.Request, role string) (Access, error) {
	teams, err := a.teamFetcher.GetTeams()
	if err != nil {
		return nil, fmt.Errorf("fetch teams: %w", err)
	}
	return NewAccessor(a.verifyToken(req), role, a.systemClaimKey, a.systemClaimValues, teams, a.displayUserIdGenerator), nil
}

func (a *accessFactory) verifyToken(req *http.Request) Verification {
	claims, err := a.tokenVerifier.Verify(req)
	if err != nil {
		switch err {
		case ErrVerificationNoToken:
			return Verification{HasToken: false, IsTokenValid: false}
		default:
			return Verification{HasToken: true, IsTokenValid: false}
		}
	}

	return Verification{HasToken: true, IsTokenValid: true, RawClaims: claims}
}
