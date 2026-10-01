package kernel

// contains checks if string contains substring (helper function).
// TestLeverageFallback, which used to live here, was retired with the crypto
// leverage knobs (BTCETHMaxLeverage/AltcoinMaxLeverage); the two helpers below
// are still used by the blackout and no-trade-band tests.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
