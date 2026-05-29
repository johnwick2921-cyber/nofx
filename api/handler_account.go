package api

import (
	"fmt"
	"net/http"

	"nofx/logger"
	ntwire "nofx/provider/ninjatrader"
	"nofx/trader"
	ntTrader "nofx/trader/ninjatrader"

	"github.com/gin-gonic/gin"
)

// AccountInfo mirrors the wire protocol struct for JSON response.
type AccountInfo struct {
	Name      string `json:"name"`        // e.g. "Sim101", "LiveAcct"
	IsSim     bool   `json:"is_sim"`      // true if a SIM account
	IsCurrent bool   `json:"is_current"`  // true if this is the currently selected account
}

// GetAccountsResponse is the response structure for GET /api/accounts.
type GetAccountsResponse struct {
	Current  string        `json:"current"`  // currently selected account, or ""
	Accounts []AccountInfo `json:"accounts"` // list of all available accounts
}

// SelectAccountRequest is the request structure for POST /api/account/select.
type SelectAccountRequest struct {
	Account string `json:"account" binding:"required"` // e.g. "Sim101"
}

// SelectAccountResponse is the response structure for POST /api/account/select.
type SelectAccountResponse struct {
	Current string `json:"current"` // the newly selected account
	Message string `json:"message"` // status message
}

// handleGetAccounts returns the list of available NT accounts discovered by the
// C# AddOn and the currently selected account. Protected endpoint (?trader_id=).
func (s *Server) handleGetAccounts(c *gin.Context) {
	traderID := c.Query("trader_id")
	if traderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trader_id query parameter required"})
		return
	}

	// Get the trader to access its NT server instance
	autoTrader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("trader not found: %s", traderID)})
		return
	}

	// Verify trader uses NinjaTrader via TCP (not CSV or other broker)
	tcpServer := extractTCPServer(autoTrader)
	if tcpServer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trader does not use NinjaTrader TCP bridge"})
		return
	}

	// Retrieve account list and current account from the TCP server
	accounts, current := tcpServer.GetAccountsList()

	// Convert internal wire structs to API response structs
	apiAccounts := make([]AccountInfo, len(accounts))
	for i, a := range accounts {
		apiAccounts[i] = AccountInfo{
			Name:      a.Name,
			IsSim:     a.IsSim,
			IsCurrent: a.Name == current,
		}
	}

	c.JSON(http.StatusOK, GetAccountsResponse{
		Current:  current,
		Accounts: apiAccounts,
	})
}

// handleSelectAccount switches the selected account to the one specified in the request.
// Server-side SIM guard: only SIM accounts (is_sim == true) are selectable.
// Protected endpoint, requires ?trader_id=.
func (s *Server) handleSelectAccount(c *gin.Context) {
	traderID := c.Query("trader_id")
	if traderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trader_id query parameter required"})
		return
	}

	var req SelectAccountRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %v", err)})
		return
	}

	if req.Account == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account name cannot be empty"})
		return
	}

	// Get the trader to access its NT server instance
	autoTrader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("trader not found: %s", traderID)})
		return
	}

	// Verify trader uses NinjaTrader via TCP
	tcpServer := extractTCPServer(autoTrader)
	if tcpServer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trader does not use NinjaTrader TCP bridge"})
		return
	}

	// Retrieve the full account list to validate and check SIM status
	accounts, _ := tcpServer.GetAccountsList()
	var targetAccount *ntwire.AccountInfo
	for i := range accounts {
		if accounts[i].Name == req.Account {
			targetAccount = &accounts[i]
			break
		}
	}

	if targetAccount == nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": fmt.Sprintf("account '%s' not found in available accounts", req.Account)})
		return
	}

	// SERVER-SIDE SIM GUARD: reject non-SIM accounts
	if !targetAccount.IsSim {
		logger.Warnf("api/account/select: attempted to select non-SIM account %s", req.Account)
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "Non-SIM accounts not selectable for auto-trading"})
		return
	}

	// Send the account select command to the C# AddOn
	payload := ntwire.AccountSelectPayload{Account: req.Account}
	if err := tcpServer.SendAccountSelect(payload); err != nil {
		logger.Warnf("api/account/select: failed to send account select: %v", err)
		c.JSON(http.StatusServiceUnavailable,
			gin.H{"error": fmt.Sprintf("NT client not connected: %v", err)})
		return
	}

	logger.Infof("api/account/select: switched to account %s (SIM=%v)", req.Account, targetAccount.IsSim)
	c.JSON(http.StatusOK, gin.H{
		"current_account": req.Account,
		"message":         fmt.Sprintf("switched to account %s", req.Account),
	})
}

// extractTCPServer attempts to extract a TCPServer from an AutoTrader.
// Returns the server if the trader uses NinjaTrader TCP bridge, otherwise nil.
func extractTCPServer(autoTrader *trader.AutoTrader) *ntwire.TCPServer {
	if autoTrader == nil {
		return nil
	}

	// Get the underlying broker trader (Trader interface)
	underlyingTrader := autoTrader.GetUnderlyingTrader()
	if underlyingTrader == nil {
		return nil
	}

	// Try to assert as TCPTrader (NinjaTrader TCP bridge)
	if tcpTrader, ok := underlyingTrader.(*ntTrader.TCPTrader); ok {
		return tcpTrader.GetServer()
	}

	return nil
}
