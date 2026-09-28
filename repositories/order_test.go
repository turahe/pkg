package repositories

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/turahe/pkg/types"
)

func TestValidateOrder(t *testing.T) {
	valid := []string{
		"name",
		"name ASC",
		"created_at desc",
		"users.created_at DESC",
		"name ASC, id DESC",
		"score DESC NULLS LAST",
	}
	for _, o := range valid {
		assert.NoError(t, ValidateOrder(o), o)
	}

	invalid := []string{
		"",
		"name; DROP TABLE users",
		"name ASC, ",
		"(SELECT password FROM users LIMIT 1)",
		"CASE WHEN 1=1 THEN name END",
		"name ASC -- comment",
		"1",
		"name SIDEWAYS",
		"a.b.c",
	}
	for _, o := range invalid {
		assert.ErrorIs(t, ValidateOrder(o), ErrInvalidOrder, o)
	}
}

// TestRepository_RejectsInvalidOrder verifies no SQL is executed for an unsafe order; sqlmock fails on
// any unexpected query.
func TestRepository_RejectsInvalidOrder(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()
	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	repo := NewBaseRepositoryWithDB(gdb)
	ctx := context.Background()
	bad := "id; DROP TABLE users"
	var out []TestModel

	assert.ErrorIs(t, repo.Find(ctx, &out, types.Conditions{}, bad), ErrInvalidOrder)

	_, err = repo.Scan(ctx, "", &TestModel{}, &out, types.Conditions{}, bad)
	assert.ErrorIs(t, err, ErrInvalidOrder)

	_, err = repo.SimplePagination(ctx, &TestModel{}, &out, 1, 10, types.Conditions{}, []string{bad})
	assert.ErrorIs(t, err, ErrInvalidOrder)

	assert.NoError(t, mock.ExpectationsWereMet())
}
