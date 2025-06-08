package testutil

import (
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// convertToDriverValues converts interface{} slice to driver.Value slice
func convertToDriverValues(args []interface{}) []driver.Value {
	driverValues := make([]driver.Value, len(args))
	for i, arg := range args {
		driverValues[i] = driver.Value(arg)
	}
	return driverValues
}

// MockDBWithStruct provides a mock database for testing
type MockDBWithStruct struct {
	DB   *sqlx.DB
	Mock sqlmock.Sqlmock
}

// NewMockDBWithStruct creates a new mock database
func NewMockDBWithStruct(t *testing.T) *MockDBWithStruct {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	
	return &MockDBWithStruct{
		DB:   sqlxDB,
		Mock: mock,
	}
}

// Close closes the mock database
func (m *MockDBWithStruct) Close() error {
	return m.DB.Close()
}

// ExpectBegin expects a transaction begin
func (m *MockDBWithStruct) ExpectBegin() *MockDBWithStruct {
	m.Mock.ExpectBegin()
	return m
}

// ExpectCommit expects a transaction commit
func (m *MockDBWithStruct) ExpectCommit() *MockDBWithStruct {
	m.Mock.ExpectCommit()
	return m
}

// ExpectRollback expects a transaction rollback
func (m *MockDBWithStruct) ExpectRollback() *MockDBWithStruct {
	m.Mock.ExpectRollback()
	return m
}

// AssertExpectations asserts all expectations were met
func (m *MockDBWithStruct) AssertExpectations(t *testing.T) {
	err := m.Mock.ExpectationsWereMet()
	require.NoError(t, err)
}

// QueryBuilder helps build expected queries with common patterns
type QueryBuilder struct {
	mock sqlmock.Sqlmock
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(mock sqlmock.Sqlmock) *QueryBuilder {
	return &QueryBuilder{mock: mock}
}

// ExpectUserByID expects a query for user by ID
func (q *QueryBuilder) ExpectUserByID(userID int64, user *models.User) {
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = $1`
	
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "created_at", "updated_at"}).
		AddRow(user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	
	q.mock.ExpectQuery(query).
		WithArgs(userID).
		WillReturnRows(rows)
}

// ExpectUserByUsername expects a query for user by username
func (q *QueryBuilder) ExpectUserByUsername(username string, user *models.User) {
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = $1`
	
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "created_at", "updated_at"}).
		AddRow(user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	
	q.mock.ExpectQuery(query).
		WithArgs(username).
		WillReturnRows(rows)
}

// ExpectCharacterByID expects a query for character by ID
func (q *QueryBuilder) ExpectCharacterByID(charID int64, char *models.Character) {
	query := `SELECT * FROM characters WHERE id = $1`
	
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "name", "race", "class", "level", 
		"experience_points", "hit_points", "max_hit_points",
		"armor_class", "initiative", "speed", "abilities",
		"skills", "proficiencies", "equipment", "spells",
		"created_at", "updated_at",
	}).AddRow(
		char.ID, char.UserID, char.Name, char.Race, char.Class, char.Level,
		char.ExperiencePoints, char.HitPoints, char.MaxHitPoints,
		char.ArmorClass, char.Initiative, char.Speed, char.Attributes,
		char.Skills, char.Proficiencies, char.Equipment, char.Spells,
		char.CreatedAt, char.UpdatedAt,
	)
	
	q.mock.ExpectQuery(query).
		WithArgs(charID).
		WillReturnRows(rows)
}

// ExpectInsertUser expects an insert user query
func (q *QueryBuilder) ExpectInsertUser(user *models.User) {
	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(user.ID, user.CreatedAt, user.UpdatedAt)
	
	q.mock.ExpectQuery(query).
		WithArgs(user.Username, user.Email, user.PasswordHash).
		WillReturnRows(rows)
}

// ExpectUpdateCharacterHP expects an update character HP query
func (q *QueryBuilder) ExpectUpdateCharacterHP(charID int64, hp int) {
	query := `UPDATE characters SET hit_points = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	
	q.mock.ExpectExec(query).
		WithArgs(charID, hp).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

// ExpectNotFound expects a query that returns no rows
func (q *QueryBuilder) ExpectNotFound(query string, args ...interface{}) {
	q.mock.ExpectQuery(query).
		WithArgs(convertToDriverValues(args)...).
		WillReturnError(sql.ErrNoRows)
}

// DBTestCase represents a database test case
type DBTestCase struct {
	Name    string
	Setup   func(mock sqlmock.Sqlmock)
	Run     func(db *sqlx.DB) error
	Assert  func(t *testing.T, err error)
}

// RunDBTestCases runs multiple database test cases
func RunDBTestCases(t *testing.T, cases []DBTestCase) {
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			db, mock := NewMockDB(t)
			defer db.Close()

			if tc.Setup != nil {
				tc.Setup(mock)
			}

			err := tc.Run(db)

			if tc.Assert != nil {
				tc.Assert(t, err)
			}

			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}

// TransactionTestHelper helps test transactional operations
type TransactionTestHelper struct {
	mock sqlmock.Sqlmock
}

// NewTransactionTestHelper creates a new transaction test helper
func NewTransactionTestHelper(mock sqlmock.Sqlmock) *TransactionTestHelper {
	return &TransactionTestHelper{mock: mock}
}

// ExpectSuccessfulTransaction sets up expectations for a successful transaction
func (h *TransactionTestHelper) ExpectSuccessfulTransaction() {
	h.mock.ExpectBegin()
	// Transaction operations will be added here
	h.mock.ExpectCommit()
}

// ExpectFailedTransaction sets up expectations for a failed transaction
func (h *TransactionTestHelper) ExpectFailedTransaction() {
	h.mock.ExpectBegin()
	// Transaction operations will be added here
	h.mock.ExpectRollback()
}

// DBFixtures provides test data fixtures
type DBFixtures struct {
	Users      []*User
	Characters []*Character
	Sessions   []*GameSession
	Combats    []*Combat
}

// LoadTestFixtures loads standard test fixtures
func LoadTestFixtures() *DBFixtures {
	now := time.Now()
	
	return &DBFixtures{
		Users: []*User{
			{
				ID:           1,
				Username:     "testdm",
				Email:        "dm@test.com",
				PasswordHash: "$2a$10$hash1",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			{
				ID:           2,
				Username:     "player1",
				Email:        "player1@test.com",
				PasswordHash: "$2a$10$hash2",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		},
		Characters: []*Character{
			{
				ID:               1,
				UserID:           2,
				Name:             "Thorin",
				Race:             "Dwarf",
				Class:            "Fighter",
				Level:            5,
				ExperiencePoints: 6500,
				HitPoints:        44,
				MaxHitPoints:     44,
				ArmorClass:       18,
				Initiative:       1,
				Speed:            25,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
		Sessions: []*GameSession{
			{
				ID:         1,
				Name:       "Test Campaign",
				DmID:       1,
				Status:     "active",
				MaxPlayers: 6,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
	}
}

// SeedTestData helps seed test data into mock expectations
func SeedTestData(mock sqlmock.Sqlmock, fixtures *DBFixtures) {
	// This can be extended to automatically set up common expectations
	// based on the fixtures provided
}

// AssertTimestampsEqual asserts two timestamps are equal within a tolerance
func AssertTimestampsEqual(t *testing.T, expected, actual time.Time, message string) {
	tolerance := time.Second
	diff := expected.Sub(actual)
	if diff < 0 {
		diff = -diff
	}
	require.True(t, diff < tolerance, 
		"%s: expected %v, got %v (diff: %v)", 
		message, expected, actual, diff)
}

// TestRepository provides a base for repository tests
type TestRepository struct {
	t      *testing.T
	db     *sqlx.DB
	mock   sqlmock.Sqlmock
}

// NewTestRepository creates a new test repository
func NewTestRepository(t *testing.T) *TestRepository {
	db, mock := NewMockDB(t)
	return &TestRepository{
		t:    t,
		db:   db,
		mock: mock,
	}
}

// Cleanup cleans up test resources
func (r *TestRepository) Cleanup() {
	r.db.Close()
	err := r.mock.ExpectationsWereMet()
	require.NoError(r.t, err)
}

// ExpectQuery adds a query expectation
func (r *TestRepository) ExpectQuery(query string) *sqlmock.ExpectedQuery {
	return r.mock.ExpectQuery(query)
}

// ExpectExec adds an exec expectation
func (r *TestRepository) ExpectExec(query string) *sqlmock.ExpectedExec {
	return r.mock.ExpectExec(query)
}

// Simple struct definitions for the mock helpers
// In real implementation, these would be your actual model structs
type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Character struct {
	ID               int64
	UserID           int64
	Name             string
	Race             string
	Class            string
	Level            int
	ExperiencePoints int
	HitPoints        int
	MaxHitPoints     int
	ArmorClass       int
	Initiative       int
	Speed            int
	Abilities        interface{}
	Skills           interface{}
	Proficiencies    interface{}
	Equipment        interface{}
	SpellSlots       interface{}
	KnownSpells      interface{}
	PreparedSpells   interface{}
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type GameSession struct {
	ID         int64
	Name       string
	DmID       int64
	Status     string
	MaxPlayers int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Combat struct {
	ID            int64
	GameSessionID int64
	Round         int
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}