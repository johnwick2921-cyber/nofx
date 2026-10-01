package api

import (
	"github.com/gin-gonic/gin"
)

type beginnerOnboardingResponse struct {
	Address           string `json:"address"`
	PrivateKey        string `json:"private_key"`
	Chain             string `json:"chain"`
	Asset             string `json:"asset"`
	Provider          string `json:"provider"`
	DefaultModel      string `json:"default_model"`
	ConfiguredModelID string `json:"configured_model_id"`
	BalanceUSDC       string `json:"balance_usdc"`
	EnvSaved          bool   `json:"env_saved"`
	EnvPath           string `json:"env_path,omitempty"`
	ReusedExisting    bool   `json:"reused_existing"`
	EnvWarning        string `json:"env_warning,omitempty"`
}

type currentBeginnerWalletResponse struct {
	Found         bool   `json:"found"`
	Address       string `json:"address,omitempty"`
	BalanceUSDC   string `json:"balance_usdc,omitempty"`
	Source        string `json:"source,omitempty"`
	Claw402Status string `json:"claw402_status"`
}

// requireOwner (N-1, verify/0926-system) — true only for a NON-machine JWT
// whose user is the first-created account, the single-user system's owner
// (registration closes after the first user). Fail closed: nil claims, a
// machine token, an empty users table, or a store error all refuse.
func (s *Server) requireOwner(c *gin.Context) bool {
	cl := authClaimsFrom(c)
	if cl == nil || cl.UserID == "" || cl.IsMachine() {
		return false
	}
	users, err := s.store.User().GetAll()
	if err != nil || len(users) == 0 {
		return false
	}
	return users[0].ID == cl.UserID
}

// apiTradingMode is a seam kept for tests: the build is futures-only (C2), so
// it always reports "futures".
var apiTradingMode = func() string {
	return "futures"
}

// tradingModeIsFutures reports whether this build is running the futures
