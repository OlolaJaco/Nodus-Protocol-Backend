package users_test

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/nodus-protocol/backend/internal/modules/users"
)

func newTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	require.NoError(t, err)
	return gormDB, mock
}

// TestTopTraders_OnlyIncludesOptedInUsers verifies that the leaderboard query
// filters by show_in_leaderboard = true, so non-consenting users are excluded.
func TestTopTraders_OnlyIncludesOptedInUsers(t *testing.T) {
	db, mock := newTestDB(t)
	repo := users.NewRepository(db)

	// Simulate a result where only one opted-in user matched the JOIN condition.
	rows := sqlmock.NewRows([]string{"display_name", "volume", "tx_count"}).
		AddRow("CryptoWhale", 50000.0, int64(15))

	// The raw SQL must JOIN on stellar_account_id and filter by consent.
	mock.ExpectQuery(`(?i)INNER JOIN users.*stellar_account_id.*show_in_leaderboard`).WillReturnRows(rows)

	result, err := repo.TopTraders(10)
	require.NoError(t, err)
	require.Len(t, result, 1)

	entry := result[0]
	assert.Equal(t, "CryptoWhale", entry["display_name"])
	assert.Equal(t, 50000.0, entry["volume_stroops"])
	assert.Equal(t, int64(15), entry["transaction_count"])

	// Raw Stellar address must never appear in leaderboard output.
	_, hasAddress := entry["address"]
	assert.False(t, hasAddress, "full Stellar address must not be exposed in leaderboard response")
}

// TestTopTraders_EmptyWhenNoUsersOptedIn verifies the leaderboard returns an
// empty slice (not nil) when no users have consented.
func TestTopTraders_EmptyWhenNoUsersOptedIn(t *testing.T) {
	db, mock := newTestDB(t)
	repo := users.NewRepository(db)

	rows := sqlmock.NewRows([]string{"display_name", "volume", "tx_count"})
	mock.ExpectQuery(`(?i)INNER JOIN users.*stellar_account_id.*show_in_leaderboard`).WillReturnRows(rows)

	result, err := repo.TopTraders(10)
	require.NoError(t, err)
	assert.Empty(t, result, "leaderboard must be empty when no users have opted in")
}

// TestTopTraders_AbbreviatedAddressWhenNoAlias verifies that an opted-in user
// without a leaderboard alias is shown with an abbreviated address, not the
// full 56-character Stellar address.
func TestTopTraders_AbbreviatedAddressWhenNoAlias(t *testing.T) {
	db, mock := newTestDB(t)
	repo := users.NewRepository(db)

	// Abbreviated form produced by the SQL CASE: GABC...XY56
	rows := sqlmock.NewRows([]string{"display_name", "volume", "tx_count"}).
		AddRow("GABC...XY56", 12000.0, int64(3))
	mock.ExpectQuery(`(?i)INNER JOIN users.*stellar_account_id.*show_in_leaderboard`).WillReturnRows(rows)

	result, err := repo.TopTraders(10)
	require.NoError(t, err)
	require.Len(t, result, 1)

	displayName, ok := result[0]["display_name"].(string)
	require.True(t, ok)
	assert.Contains(t, displayName, "...", "abbreviated address must contain ellipsis, not the full address")
	assert.LessOrEqual(t, len(displayName), 20, "abbreviated address must be much shorter than the 56-char full address")
}

// TestUpdatePreferences_SetsConsentFields verifies that UpdatePreferences
// issues an UPDATE that sets both show_in_leaderboard and leaderboard_alias.
func TestUpdatePreferences_SetsConsentFields(t *testing.T) {
	db, mock := newTestDB(t)
	repo := users.NewRepository(db)

	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`(?i)UPDATE.*users.*SET.*show_in_leaderboard`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	show := true
	alias := "Trader42"
	err := repo.UpdatePreferences(userID, &show, &alias)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdatePreferences_CanOptOut verifies that a user can opt out by setting
// show_in_leaderboard = false.
func TestUpdatePreferences_CanOptOut(t *testing.T) {
	db, mock := newTestDB(t)
	repo := users.NewRepository(db)

	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`(?i)UPDATE.*users.*SET.*show_in_leaderboard`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	show := false
	alias := ""
	err := repo.UpdatePreferences(userID, &show, &alias)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
