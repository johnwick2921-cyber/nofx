package store

import (
	"fmt"
	"nofx/crypto"
	"nofx/logger"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExchangeStore exchange storage
type ExchangeStore struct {
	db *gorm.DB
}

// Exchange exchange configuration
type Exchange struct {
	ID                      string                 `gorm:"primaryKey" json:"id"`
	ExchangeType            string                 `gorm:"column:exchange_type;not null;default:''" json:"exchange_type"`
	AccountName             string                 `gorm:"column:account_name;not null;default:''" json:"account_name"`
	UserID                  string                 `gorm:"column:user_id;not null;default:default;index" json:"user_id"`
	Name                    string                 `gorm:"not null" json:"name"`
	Type                    string                 `gorm:"not null" json:"type"` // "cex" or "dex"
	Enabled                 bool                   `gorm:"default:false" json:"enabled"`
	APIKey                  crypto.EncryptedString `gorm:"column:api_key;default:''" json:"apiKey"`
	SecretKey               crypto.EncryptedString `gorm:"column:secret_key;default:''" json:"secretKey"`
	Passphrase              crypto.EncryptedString `gorm:"column:passphrase;default:''" json:"passphrase"`
	Testnet                 bool                   `gorm:"default:false" json:"testnet"`
	HyperliquidWalletAddr   string                 `gorm:"column:hyperliquid_wallet_addr;default:''" json:"hyperliquidWalletAddr"`
	HyperliquidUnifiedAcct  bool                   `gorm:"column:hyperliquid_unified_account;default:true" json:"hyperliquidUnifiedAccount"` // Unified Account mode (Spot as collateral)
	AsterUser               string                 `gorm:"column:aster_user;default:''" json:"asterUser"`
	AsterSigner             string                 `gorm:"column:aster_signer;default:''" json:"asterSigner"`
	AsterPrivateKey         crypto.EncryptedString `gorm:"column:aster_private_key;default:''" json:"asterPrivateKey"`
	LighterWalletAddr       string                 `gorm:"column:lighter_wallet_addr;default:''" json:"lighterWalletAddr"`
	LighterPrivateKey       crypto.EncryptedString `gorm:"column:lighter_private_key;default:''" json:"lighterPrivateKey"`
	LighterAPIKeyPrivateKey crypto.EncryptedString `gorm:"column:lighter_api_key_private_key;default:''" json:"lighterAPIKeyPrivateKey"`
	LighterAPIKeyIndex      int                    `gorm:"column:lighter_api_key_index;default:0" json:"lighterAPIKeyIndex"`
	// NinjaTrader CSV bridge configuration (no API key required)
	NTDataDir            string    `gorm:"column:nt_data_dir;default:''" json:"ntDataDir"`
	NTInstrumentName     string    `gorm:"column:nt_instrument_name;default:''" json:"ntInstrumentName"`
	NTDefaultContractQty int       `gorm:"column:nt_default_contract_qty;default:0" json:"ntDefaultContractQty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (Exchange) TableName() string { return "exchanges" }

// NewExchangeStore creates a new ExchangeStore
func NewExchangeStore(db *gorm.DB) *ExchangeStore {
	return &ExchangeStore{db: db}
}

func (s *ExchangeStore) initTables() error {
	// For PostgreSQL with existing table, skip AutoMigrate
	if s.db.Dialector.Name() == "postgres" {
		var tableExists int64
		s.db.Raw(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'exchanges'`).Scan(&tableExists)
		if tableExists > 0 {
			// Still run data migrations
			s.migrateToMultiAccount()
			s.backfillDefaultAccountName()
			if err := s.cleanupIncompleteExchangeConfigs(); err != nil {
				logger.Warnf("Exchange cleanup migration warning: %v", err)
			}
			return nil
		}
	}

	if err := s.db.AutoMigrate(&Exchange{}); err != nil {
		return err
	}

	// Run migration to multi-account if needed
	if err := s.migrateToMultiAccount(); err != nil {
		logger.Warnf("Multi-account migration warning: %v", err)
	}

	// Fix empty account_name for existing records
	s.backfillDefaultAccountName()
	if err := s.cleanupIncompleteExchangeConfigs(); err != nil {
		logger.Warnf("Exchange cleanup migration warning: %v", err)
	}

	return nil
}

// backfillDefaultAccountName names an unnamed account "Default" — for rows of a
// SUPPORTED exchange type only. A row whose type this build does not construct
// is never written at boot (see cleanupIncompleteExchangeConfigs).
func (s *ExchangeStore) backfillDefaultAccountName() {
	s.db.Model(&Exchange{}).
		Where("(account_name = '' OR account_name IS NULL) AND exchange_type IN ?", supportedExchangeTypes).
		Update("account_name", "Default")
}

// cleanupIncompleteExchangeConfigs runs on every boot (initTables). For a row
// of a SUPPORTED type it deletes an incomplete config and enables a complete
// one, as before. A row whose exchange_type is NOT in the registry — a broker
// removed from the build, or no type at all — is LEFT UNTOUCHED: never
// deleted, never enabled, no data.db write. Its traders are refused by name at
// load (trader.ExchangeRefusal). Deleting it here would have been a silent
// boot-time write that turned "refused" into "gone".
func (s *ExchangeStore) cleanupIncompleteExchangeConfigs() error {
	var exchanges []Exchange
	if err := s.db.Find(&exchanges).Error; err != nil {
		return err
	}
	for _, exchange := range exchanges {
		if err := CheckSupportedExchangeType(exchange.ExchangeType); err != nil {
			logger.Warnf("🧹⛔ Exchange row left untouched at boot (not deleted, not enabled): id=%s user=%s — %v; its traders are refused at load", exchange.ID, exchange.UserID, err)
			continue
		}
		missing := MissingRequiredExchangeCredentialFields(
			exchange.ExchangeType,
			string(exchange.APIKey),
			string(exchange.SecretKey),
			string(exchange.Passphrase),
			exchange.HyperliquidWalletAddr,
			exchange.AsterUser,
			exchange.AsterSigner,
			string(exchange.AsterPrivateKey),
			exchange.LighterWalletAddr,
			string(exchange.LighterAPIKeyPrivateKey),
			exchange.NTDataDir,
		)
		if len(missing) > 0 {
			if err := s.db.Delete(&Exchange{}, "id = ? AND user_id = ?", exchange.ID, exchange.UserID).Error; err != nil {
				return err
			}
			logger.Infof("🧹 Removed incomplete exchange config during migration: id=%s user=%s missing=%s", exchange.ID, exchange.UserID, strings.Join(missing, ","))
			continue
		}
		if !exchange.Enabled {
			if err := s.db.Model(&Exchange{}).Where("id = ? AND user_id = ?", exchange.ID, exchange.UserID).Update("enabled", true).Error; err != nil {
				return err
			}
			logger.Infof("🧹 Enabled complete exchange config during migration: id=%s user=%s", exchange.ID, exchange.UserID)
		}
	}
	return nil
}

// legacyExchangeTypeIDs are the old-schema ids (id = exchange type, empty
// exchange_type) that migrateToMultiAccount rewrites to a UUID row. Only SUPPORTED types
// are listed: an old-schema row for a broker removed from the build is left
// untouched (no boot write) and its traders are refused at load.
var legacyExchangeTypeIDs = []string{"bybit", "okx", "bitget", "hyperliquid", "aster", "lighter"}

// migrateToMultiAccount migrates old schema (id=exchange_type) to new schema (id=UUID)
func (s *ExchangeStore) migrateToMultiAccount() error {
	// Check if migration is needed by looking for old-style IDs (non-UUID)
	var count int64
	err := s.db.Model(&Exchange{}).
		Where("exchange_type = '' AND id IN ?", legacyExchangeTypeIDs).
		Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	logger.Infof("🔄 Migrating %d exchange records to multi-account schema...", count)

	// Get all old records
	var records []Exchange
	err = s.db.Where("exchange_type = '' AND id IN ?", legacyExchangeTypeIDs).
		Find(&records).Error
	if err != nil {
		return err
	}

	// Begin transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, r := range records {
			newID := uuid.New().String()
			oldID := r.ID // This is the exchange type (e.g., "bybit")

			// Update traders table to use new UUID
			if err := tx.Exec("UPDATE traders SET exchange_id = ? WHERE exchange_id = ? AND user_id = ?",
				newID, oldID, r.UserID).Error; err != nil {
				logger.Errorf("Failed to update traders for exchange %s: %v", oldID, err)
				return err
			}

			// Update the exchange record
			if err := tx.Model(&Exchange{}).
				Where("id = ? AND user_id = ?", oldID, r.UserID).
				Updates(map[string]interface{}{
					"id":            newID,
					"exchange_type": oldID,
					"account_name":  "Default",
				}).Error; err != nil {
				logger.Errorf("Failed to migrate exchange %s: %v", oldID, err)
				return err
			}

			logger.Infof("✅ Migrated exchange %s -> UUID %s for user %s", oldID, newID, r.UserID)
		}
		return nil
	})
}

func (s *ExchangeStore) initDefaultData() error {
	// No longer pre-populate exchanges - create on demand when user configures
	return nil
}

// List gets user's exchange list
func (s *ExchangeStore) List(userID string) ([]*Exchange, error) {
	var exchanges []*Exchange
	err := s.db.Where("user_id = ?", userID).Order("exchange_type, account_name").Find(&exchanges).Error
	if err != nil {
		return nil, err
	}
	return exchanges, nil
}

// GetByID gets a specific exchange by UUID
func (s *ExchangeStore) GetByID(userID, id string) (*Exchange, error) {
	var exchange Exchange
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&exchange).Error
	if err != nil {
		return nil, err
	}
	return &exchange, nil
}

// getExchangeNameAndType returns the display name and type for an exchange type
func getExchangeNameAndType(exchangeType string) (name string, typ string) {
	switch exchangeType {
	case "bybit":
		return "Bybit Futures", "cex"
	case "okx":
		return "OKX Futures", "cex"
	case "bitget":
		return "Bitget Futures", "cex"
	case "hyperliquid":
		return "Hyperliquid", "dex"
	case "aster":
		return "Aster DEX", "dex"
	case "lighter":
		return "LIGHTER DEX", "dex"
	case "indodax":
		return "Indodax", "cex"
	case "ninjatrader":
		return "NinjaTrader", "futures"
	default:
		return exchangeType + " Exchange", "cex"
	}
}

// Create creates a new exchange account with UUID.
// NinjaTrader fields (ntDataDir/ntInstrumentName/ntDefaultContractQty) are
// only meaningful when exchangeType=="ninjatrader"; pass "" / 0 otherwise.
func (s *ExchangeStore) Create(userID, exchangeType, accountName string, enabled bool,
	apiKey, secretKey, passphrase string, testnet bool,
	hyperliquidWalletAddr string, hyperliquidUnifiedAcct bool,
	asterUser, asterSigner, asterPrivateKey,
	lighterWalletAddr, lighterPrivateKey, lighterApiKeyPrivateKey string, lighterApiKeyIndex int,
	ntDataDir, ntInstrumentName string, ntDefaultContractQty int) (string, error) {

	// The registry is consulted FIRST, so a type this build cannot construct is
	// refused by name — not reported as "missing exchange_type". (Case-folded
	// like the credential check below, so existing callers keep their casing
	// tolerance.)
	if !IsSupportedExchangeType(strings.ToLower(strings.TrimSpace(exchangeType))) {
		return "", CheckSupportedExchangeType(exchangeType)
	}

	if missing := MissingRequiredExchangeCredentialFields(
		exchangeType, apiKey, secretKey, passphrase,
		hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey,
		lighterWalletAddr, lighterApiKeyPrivateKey,
		ntDataDir,
	); len(missing) > 0 {
		return "", fmt.Errorf("missing required exchange fields: %s", strings.Join(missing, ", "))
	}

	id := uuid.New().String()
	name, typ := getExchangeNameAndType(exchangeType)

	if accountName == "" {
		accountName = "Default"
	}

	logger.Debugf("🔧 ExchangeStore.Create: userID=%s, exchangeType=%s, accountName=%s, id=%s",
		userID, exchangeType, accountName, id)

	exchange := &Exchange{
		ID:                      id,
		ExchangeType:            exchangeType,
		AccountName:             accountName,
		UserID:                  userID,
		Name:                    name,
		Type:                    typ,
		Enabled:                 true,
		APIKey:                  crypto.EncryptedString(apiKey),
		SecretKey:               crypto.EncryptedString(secretKey),
		Passphrase:              crypto.EncryptedString(passphrase),
		Testnet:                 testnet,
		HyperliquidWalletAddr:   hyperliquidWalletAddr,
		HyperliquidUnifiedAcct:  hyperliquidUnifiedAcct,
		AsterUser:               asterUser,
		AsterSigner:             asterSigner,
		AsterPrivateKey:         crypto.EncryptedString(asterPrivateKey),
		LighterWalletAddr:       lighterWalletAddr,
		LighterPrivateKey:       crypto.EncryptedString(lighterPrivateKey),
		LighterAPIKeyPrivateKey: crypto.EncryptedString(lighterApiKeyPrivateKey),
		LighterAPIKeyIndex:      lighterApiKeyIndex,
		NTDataDir:               ntDataDir,
		NTInstrumentName:        ntInstrumentName,
		NTDefaultContractQty:    ntDefaultContractQty,
	}

	if err := s.db.Create(exchange).Error; err != nil {
		return "", err
	}
	return id, nil
}

// Update updates exchange configuration by UUID.
// NinjaTrader fields (ntDataDir/ntInstrumentName/ntDefaultContractQty) are
// only meaningful when the row is type "ninjatrader"; pass "" / 0 otherwise.
func (s *ExchangeStore) Update(userID, id string, enabled bool, apiKey, secretKey, passphrase string, testnet bool,
	hyperliquidWalletAddr string, hyperliquidUnifiedAcct bool,
	asterUser, asterSigner, asterPrivateKey, lighterWalletAddr, lighterPrivateKey, lighterApiKeyPrivateKey string, lighterApiKeyIndex int,
	ntDataDir, ntInstrumentName string, ntDefaultContractQty int) error {

	logger.Debugf("🔧 ExchangeStore.Update: userID=%s, id=%s", userID, id)

	updates := map[string]interface{}{
		"enabled":                     true,
		"testnet":                     testnet,
		"hyperliquid_wallet_addr":     hyperliquidWalletAddr,
		"hyperliquid_unified_account": hyperliquidUnifiedAcct,
		"aster_user":                  asterUser,
		"aster_signer":                asterSigner,
		"lighter_wallet_addr":         lighterWalletAddr,
		"lighter_api_key_index":       lighterApiKeyIndex,
		"updated_at":                  time.Now().UTC(),
	}
	if ntDataDir != "" {
		updates["nt_data_dir"] = ntDataDir
	}
	if ntInstrumentName != "" {
		updates["nt_instrument_name"] = ntInstrumentName
	}
	if ntDefaultContractQty != 0 {
		updates["nt_default_contract_qty"] = ntDefaultContractQty
	}

	// Only update encrypted fields if not empty
	if apiKey != "" {
		updates["api_key"] = crypto.EncryptedString(apiKey)
	}
	if secretKey != "" {
		updates["secret_key"] = crypto.EncryptedString(secretKey)
	}
	if passphrase != "" {
		updates["passphrase"] = crypto.EncryptedString(passphrase)
	}
	if asterPrivateKey != "" {
		updates["aster_private_key"] = crypto.EncryptedString(asterPrivateKey)
	}
	if lighterPrivateKey != "" {
		updates["lighter_private_key"] = crypto.EncryptedString(lighterPrivateKey)
	}
	if lighterApiKeyPrivateKey != "" {
		updates["lighter_api_key_private_key"] = crypto.EncryptedString(lighterApiKeyPrivateKey)
	}

	result := s.db.Model(&Exchange{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("exchange not found: id=%s, userID=%s", id, userID)
	}
	return nil
}

// UpdateAccountName updates the account name for an exchange
func (s *ExchangeStore) UpdateAccountName(userID, id, accountName string) error {
	result := s.db.Model(&Exchange{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"account_name": accountName,
			"updated_at":   time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("exchange not found: id=%s, userID=%s", id, userID)
	}
	return nil
}

// Delete deletes an exchange account
func (s *ExchangeStore) Delete(userID, id string) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Exchange{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("exchange not found: id=%s, userID=%s", id, userID)
	}
	logger.Infof("🗑️ Deleted exchange: id=%s, userID=%s", id, userID)
	return nil
}
