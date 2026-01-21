package repositories

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/database"
	"github.com/turahe/pkg/types"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dbMutex sync.Mutex

// TestModel is a test model for testing repository operations
type TestModel struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	Name      string    `gorm:"column:name"`
	Email     string    `gorm:"column:email"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (TestModel) TableName() string {
	return "test_models"
}

// setupTestDB creates a test database connection
// Tries MySQL first (for CI), falls back to SQLite (for local development)
func setupTestDB(t *testing.T) *gorm.DB {
	// Try MySQL first (for CI environments)
	mysqlCfg := &config.DatabaseConfiguration{
		Driver:   "mysql",
		Dbname:   getTestEnvOrDefault("DATABASE_DBNAME", "testing"),
		Username: getTestEnvOrDefault("DATABASE_USERNAME", "root"),
		Password: getTestEnvOrDefault("DATABASE_PASSWORD", "secret"),
		Host:     getTestEnvOrDefault("DATABASE_HOST", "127.0.0.1"),
		Port:     getTestEnvOrDefault("DATABASE_PORT", "3306"),
		Logmode:  false,
	}

	db, err := database.CreateDatabaseConnection(mysqlCfg)
	if err == nil {
		// MySQL connection successful
		// Auto-migrate test model
		err = db.AutoMigrate(&TestModel{})
		require.NoError(t, err, "Failed to migrate test model")
		return db
	}

	// Fallback to SQLite for local development
	sqliteCfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  ":memory:",
		Logmode: false,
	}

	db, err = database.CreateDatabaseConnection(sqliteCfg)
	if err != nil {
		// Skip test if neither MySQL nor SQLite is available
		t.Skipf("Database not available (MySQL: %v, SQLite: CGO may be disabled)", err)
		return nil
	}

	// Auto-migrate test model
	err = db.AutoMigrate(&TestModel{})
	require.NoError(t, err, "Failed to migrate test model")

	return db
}

// getTestEnvOrDefault gets environment variable or returns default value for testing
func getTestEnvOrDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// setupTestDBForRepo sets up the database and assigns it to the package-level DB variable
func setupTestDBForRepo(t *testing.T) func() {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	
	// Setup test DB
	db := setupTestDB(t)
	
	// Save original setup
	originalDB := database.DB
	
	// Set our test DB
	database.DB = db

	// Clean up any existing test data to ensure test isolation
	_ = db.Exec("DELETE FROM test_models").Error

	// Return cleanup function
	return func() {
		dbMutex.Lock()
		defer dbMutex.Unlock()
		
		// Clean up test data
		_ = db.Exec("DELETE FROM test_models").Error
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		database.DB = originalDB
	}
}

func TestNewBaseRepository(t *testing.T) {
	repo := NewBaseRepository()
	assert.NotNil(t, repo)
	assert.Implements(t, (*IBaseRepository)(nil), repo)

	baseRepo, ok := repo.(*BaseRepository)
	require.True(t, ok)
	assert.False(t, baseRepo.useSiteDB)
}

func TestNewSiteBaseRepository(t *testing.T) {
	repo := NewSiteBaseRepository()
	assert.NotNil(t, repo)
	assert.Implements(t, (*IBaseRepository)(nil), repo)

	baseRepo, ok := repo.(*BaseRepository)
	require.True(t, ok)
	assert.True(t, baseRepo.useSiteDB)
}

func TestBaseRepository_Create(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()
	model := &TestModel{
		Name:  "Test User",
		Email: "test@example.com",
	}

	err := repo.Create(model)
	assert.NoError(t, err)
	assert.NotZero(t, model.ID)

	// Verify the record was created
	var found TestModel
	err = database.GetDB().First(&found, model.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, model.Name, found.Name)
	assert.Equal(t, model.Email, found.Email)
}

func TestBaseRepository_Save(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()

	// Create a model first
	model := &TestModel{
		Name:  "Original Name",
		Email: "original@example.com",
	}
	err := database.GetDB().Create(model).Error
	require.NoError(t, err)

	// Update and save
	model.Name = "Updated Name"
	err = repo.Save(model)
	assert.NoError(t, err)

	// Verify the update
	var found TestModel
	err = database.GetDB().First(&found, model.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
}

func TestBaseRepository_Updates(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()

	// Create a model first
	model := &TestModel{
		Name:  "Original Name",
		Email: "original@example.com",
	}
	err := database.GetDB().Create(model).Error
	require.NoError(t, err)

	// Update using Updates method
	updates := map[string]interface{}{
		"name": "Updated Name",
	}
	err = repo.Updates(model, updates)
	assert.NoError(t, err)

	// Verify the update
	var found TestModel
	err = database.GetDB().First(&found, model.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
}

func TestBaseRepository_Delete(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()

	// Create test models
	model1 := &TestModel{Name: "User 1", Email: "user1@example.com"}
	model2 := &TestModel{Name: "User 2", Email: "user2@example.com"}
	err := database.GetDB().Create([]*TestModel{model1, model2}).Error
	require.NoError(t, err)

	// Delete with conditions
	conditions := types.Conditions{
		"id = ?": model1.ID,
	}
	count, err := repo.Delete("", &TestModel{}, conditions)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Verify deletion
	var found TestModel
	err = database.GetDB().First(&found, model1.ID).Error
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestBaseRepository_Delete_WithTableName(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()

	// Create test model
	model := &TestModel{Name: "User 1", Email: "user1@example.com"}
	err := database.GetDB().Create(model).Error
	require.NoError(t, err)

	// Delete using table name
	conditions := types.Conditions{
		"id = ?": model.ID,
	}
	count, err := repo.Delete("test_models", nil, conditions)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestBaseRepository_First(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()

	// Create test model
	model := &TestModel{Name: "Test User", Email: "test@example.com"}
	err := database.GetDB().Create(model).Error
	require.NoError(t, err)

	// Find existing record
	var found TestModel
	conditions := types.Conditions{
		"id = ?": model.ID,
	}
	notFound, err := repo.First(&found, conditions)
	assert.NoError(t, err)
	assert.False(t, notFound)
	assert.Equal(t, model.Name, found.Name)
	assert.Equal(t, model.Email, found.Email)

	// Find non-existing record
	var notFoundModel TestModel
	conditions = types.Conditions{
		"id = ?": 99999,
	}
	notFound, err = repo.First(&notFoundModel, conditions)
	assert.NoError(t, err)
	assert.True(t, notFound)
}

func TestBaseRepository_Find(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()

	// Create test models
	models := []*TestModel{
		{Name: "User 1", Email: "user1@example.com"},
		{Name: "User 2", Email: "user2@example.com"},
		{Name: "User 3", Email: "user3@example.com"},
	}
	for _, m := range models {
		err := database.GetDB().Create(m).Error
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different created_at timestamps
	}

	// Find all records
	var found []TestModel
	conditions := types.Conditions{}
	err := repo.Find(&found, conditions)
	assert.NoError(t, err)
	assert.Len(t, found, 3)

	// Find with conditions
	var filtered []TestModel
	conditions = types.Conditions{
		"name = ?": "User 1",
	}
	err = repo.Find(&filtered, conditions)
	assert.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "User 1", filtered[0].Name)

	// Find with order
	var ordered []TestModel
	err = repo.Find(&ordered, conditions, "name ASC")
	assert.NoError(t, err)
	if len(ordered) > 0 {
		assert.Equal(t, "User 1", ordered[0].Name)
	}
}

func TestBaseRepository_Scan(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()

	// Create test model
	model := &TestModel{Name: "Test User", Email: "test@example.com"}
	err := database.GetDB().Create(model).Error
	require.NoError(t, err)

	// Scan with model
	var found TestModel
	conditions := types.Conditions{
		"id = ?": model.ID,
	}
	notFound, err := repo.Scan("", &TestModel{}, &found, conditions)
	assert.NoError(t, err)
	assert.False(t, notFound)
	assert.Equal(t, model.Name, found.Name)

	// Scan with table name
	var foundWithTable TestModel
	notFound, err = repo.Scan("test_models", nil, &foundWithTable, conditions)
	assert.NoError(t, err)
	assert.False(t, notFound)
	assert.Equal(t, model.Name, foundWithTable.Name)

	// Scan non-existing record
	// Note: GORM's Scan doesn't return ErrRecordNotFound, it just returns empty result
	var notFoundModel TestModel
	conditions = types.Conditions{
		"id = ?": 99999,
	}
	notFound, err = repo.Scan("", &TestModel{}, &notFoundModel, conditions)
	assert.NoError(t, err)
	// Scan returns notFound=false when no records found (unlike First)
	// The result will just be empty (zero value)
	assert.Zero(t, notFoundModel.ID)
	assert.Empty(t, notFoundModel.Name)
}

func TestBaseRepository_RawSQL(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()

	// Create test model
	model := &TestModel{Name: "Test User", Email: "test@example.com"}
	err := database.GetDB().Create(model).Error
	require.NoError(t, err)

	// Execute raw SQL
	result := repo.RawSQL(nil, "SELECT * FROM test_models WHERE id = ?", model.ID)
	assert.NotNil(t, result)

	var found TestModel
	err = result.Scan(&found).Error
	assert.NoError(t, err)
	assert.Equal(t, model.Name, found.Name)

	// Test with specified DB
	specifiedDB := database.GetDB()
	result = repo.RawSQL(specifiedDB, "SELECT COUNT(*) as count FROM test_models")
	assert.NotNil(t, result)
}

func TestBaseRepository_ExecSQL(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()

	// Create test model
	model := &TestModel{Name: "Test User", Email: "test@example.com"}
	err := database.GetDB().Create(model).Error
	require.NoError(t, err)

	// Execute SQL update
	err = repo.ExecSQL(nil, "UPDATE test_models SET name = ? WHERE id = ?", "Updated Name", model.ID)
	assert.NoError(t, err)

	// Verify update
	var found TestModel
	err = database.GetDB().First(&found, model.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)

	// Test with specified DB
	err = repo.ExecSQL(database.GetDB(), "UPDATE test_models SET name = ? WHERE id = ?", "Updated Again", model.ID)
	assert.NoError(t, err)
}

func TestBaseRepository_IsEmpty(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()

	// Should be empty initially
	isEmpty := repo.IsEmpty(&TestModel{})
	assert.True(t, isEmpty)

	// Create a model
	model := &TestModel{Name: "Test User", Email: "test@example.com"}
	err := database.GetDB().Create(model).Error
	require.NoError(t, err)

	// Should not be empty now
	isEmpty = repo.IsEmpty(&TestModel{})
	assert.False(t, isEmpty)
}

func TestBaseRepository_SimplePagination(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := NewBaseRepository()

	// Create 15 test models
	for i := 1; i <= 15; i++ {
		model := &TestModel{
			Name:  "User",
			Email: "user@example.com",
		}
		err := database.GetDB().Create(model).Error
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different created_at timestamps
	}

	// Test first page
	var page1 []TestModel
	total, err := repo.SimplePagination(&TestModel{}, &page1, 1, 10, types.Conditions{}, []string{})
	assert.NoError(t, err)
	assert.Len(t, page1, 10)
	assert.Greater(t, total, int64(0)) // Has more pages

	// Test second page
	var page2 []TestModel
	total, err = repo.SimplePagination(&TestModel{}, &page2, 2, 10, types.Conditions{}, []string{})
	assert.NoError(t, err)
	assert.Len(t, page2, 5) // Remaining 5 items
	assert.Equal(t, int64(0), total) // No more pages

	// Test with invalid page number (should default to 1)
	var pageDefault []TestModel
	total, err = repo.SimplePagination(&TestModel{}, &pageDefault, 0, 10, types.Conditions{}, []string{})
	assert.NoError(t, err)
	assert.Len(t, pageDefault, 10)

	// Test with page size exceeding max (should cap at 100)
	var pageLarge []TestModel
	total, err = repo.SimplePagination(&TestModel{}, &pageLarge, 1, 200, types.Conditions{}, []string{})
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(pageLarge), 15) // Should return all available

	// Test with conditions
	var filtered []TestModel
	conditions := types.Conditions{
		"name = ?": "User",
	}
	total, err = repo.SimplePagination(&TestModel{}, &filtered, 1, 10, conditions, []string{})
	assert.NoError(t, err)
	assert.Greater(t, len(filtered), 0)
}

func TestBaseRepository_applyWhereCondition(t *testing.T) {
	cleanup := setupTestDBForRepo(t)
	defer cleanup()

	repo := &BaseRepository{}
	db := database.GetDB()

	// Test single placeholder
	result := repo.applyWhereCondition(db, "id = ?", 1)
	assert.NotNil(t, result)

	// Test multiple placeholders with single value
	result = repo.applyWhereCondition(db, "id IN (?, ?)", 1)
	assert.NotNil(t, result)

	// Test multiple placeholders with slice value
	result = repo.applyWhereCondition(db, "id IN (?, ?)", []interface{}{1, 2})
	assert.NotNil(t, result)

	// Test no placeholder
	result = repo.applyWhereCondition(db, "1 = 1", nil)
	assert.NotNil(t, result)
}

func TestBaseRepository_getColumnNames(t *testing.T) {
	repo := &BaseRepository{}

	// Test with struct
	columns := repo.getColumnNames(&TestModel{})
	assert.Contains(t, columns, "id")
	assert.Contains(t, columns, "name")
	assert.Contains(t, columns, "email")
	assert.Contains(t, columns, "created_at")

	// Test with pointer
	columns = repo.getColumnNames((*TestModel)(nil))
	assert.Contains(t, columns, "id")
	assert.Contains(t, columns, "name")

	// Test with slice
	var models []TestModel
	columns = repo.getColumnNames(&models)
	assert.Contains(t, columns, "id")
	assert.Contains(t, columns, "name")

	// Test with non-struct (should return empty)
	columns = repo.getColumnNames("not a struct")
	assert.Empty(t, columns)
}

func TestBaseRepository_extractColumnName(t *testing.T) {
	repo := &BaseRepository{}

	// Test standard gorm tag
	column := repo.extractColumnName("column:user_name;type:varchar(255)")
	assert.Equal(t, "user_name", column)

	// Test with multiple attributes
	column = repo.extractColumnName("column:email;type:varchar(255);not null")
	assert.Equal(t, "email", column)

	// Test with no column tag
	column = repo.extractColumnName("type:varchar(255);not null")
	assert.Empty(t, column)

	// Test empty tag
	column = repo.extractColumnName("")
	assert.Empty(t, column)
}
