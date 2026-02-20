package repositories

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"

	"gorm.io/gorm"

	"github.com/turahe/pkg/database"
	"github.com/turahe/pkg/types"
)

// columnCache caches column names by reflect.Type to avoid repeated reflection in hot path.
var columnCache sync.Map

// IBaseRepository defines the base repository interface. All methods accept context.Context for cancellation and timeouts.
//
// For Clean Architecture: define use-case-specific repository ports in domain/port (e.g. port.GetByID)
// and implement them by adapting BaseRepository or by wrapping it. Keep IBaseRepository in this package
// for backward compatibility and for code that needs full CRUD; use domain ports in use cases so they
// depend only on domain.
type IBaseRepository interface {
	Create(ctx context.Context, value interface{}) error
	Save(ctx context.Context, value interface{}) error
	Updates(ctx context.Context, where interface{}, value interface{}) error
	Delete(ctx context.Context, tableName string, model interface{}, conditions types.Conditions) (count int64, err error)
	First(ctx context.Context, out interface{}, conditions types.Conditions) (notFound bool, err error)
	Find(ctx context.Context, out interface{}, conditions types.Conditions, orders ...string) error
	Scan(ctx context.Context, tableName string, model, out interface{}, conditions types.Conditions, orders ...string) (notFound bool, err error)
	RawSQL(ctx context.Context, specifyDb *gorm.DB, query string, args ...interface{}) *gorm.DB
	ExecSQL(ctx context.Context, specifyDb *gorm.DB, query string, args ...interface{}) error
	IsEmpty(ctx context.Context, model interface{}) bool
	SimplePagination(ctx context.Context, model, out interface{}, pageNumber, pageSize int, conditions types.Conditions, orders []string, preloads ...string) (total int64, err error)
}

// BaseRepository implements IBaseRepository using GORM.
// Use NewBaseRepository or NewSiteBaseRepository for global DB, or NewBaseRepositoryWithDB for injection.
type BaseRepository struct {
	db       *gorm.DB   // optional: when set, used instead of database.GetDB()
	siteDB   *gorm.DB   // optional: when set, used instead of database.GetDBSite()
	useSiteDB bool      // when db/siteDB are nil, use GetDBSite() when true else GetDB()
}

// NewBaseRepository creates a new base repository using the global main database (database.GetDB()).
func NewBaseRepository() IBaseRepository {
	return &BaseRepository{useSiteDB: false}
}

// NewSiteBaseRepository creates a new base repository using the global site database (database.GetDBSite()).
func NewSiteBaseRepository() IBaseRepository {
	return &BaseRepository{useSiteDB: true}
}

// NewBaseRepositoryWithDB creates a base repository with an injected *gorm.DB (dependency injection).
func NewBaseRepositoryWithDB(db *gorm.DB) IBaseRepository {
	return &BaseRepository{db: db}
}

// NewSiteBaseRepositoryWithDB creates a base repository with an injected site *gorm.DB.
func NewSiteBaseRepositoryWithDB(siteDB *gorm.DB) IBaseRepository {
	return &BaseRepository{siteDB: siteDB}
}

// getDB returns the appropriate database connection (injected or global).
func (r *BaseRepository) getDB() *gorm.DB {
	if r.siteDB != nil {
		return r.siteDB
	}
	if r.db != nil {
		return r.db
	}
	if r.useSiteDB {
		return database.GetDBSite()
	}
	return database.GetDB()
}

// applyWhereCondition applies a WHERE condition to a GORM query builder
// Handles cases where the condition key contains multiple placeholders (?)
// but only a single value is provided by duplicating the value
func (r *BaseRepository) applyWhereCondition(db *gorm.DB, key string, value interface{}) *gorm.DB {
	// Count the number of placeholders in the key
	placeholderCount := strings.Count(key, "?")

	if placeholderCount > 1 {
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			l := val.Len()
			args := make([]interface{}, 0, l)
			for i := 0; i < l; i++ {
				args = append(args, val.Index(i).Interface())
			}
			return db.Where(key, args...)
		}

		// Value is not a slice, duplicate it for each placeholder
		args := make([]interface{}, placeholderCount)
		for i := 0; i < placeholderCount; i++ {
			args[i] = value
		}
		return db.Where(key, args...)
	}

	// Single placeholder or no placeholder, use the value directly
	return db.Where(key, value)
}

func (r *BaseRepository) Create(ctx context.Context, value interface{}) error {
	return r.getDB().WithContext(ctx).Create(value).Error
}

func (r *BaseRepository) Save(ctx context.Context, value interface{}) error {
	return r.getDB().WithContext(ctx).Save(value).Error
}

func (r *BaseRepository) Updates(ctx context.Context, where interface{}, value interface{}) error {
	return r.getDB().WithContext(ctx).Model(where).Updates(value).Error
}

func (r *BaseRepository) Delete(ctx context.Context, tableName string, model interface{}, conditions types.Conditions) (count int64, err error) {
	db := r.getDB().WithContext(ctx)

	for key, value := range conditions {
		db = r.applyWhereCondition(db, key, value)
	}

	if model == nil && tableName != "" {
		db = db.Table(tableName).Delete(nil)
	} else {
		db = db.Delete(model)
	}

	err = db.Error
	if err != nil {
		return 0, err
	}

	count = db.RowsAffected
	return
}

func (r *BaseRepository) First(ctx context.Context, out interface{}, conditions types.Conditions) (notFound bool, err error) {
	db := r.getDB().WithContext(ctx)

	for key, value := range conditions {
		db = r.applyWhereCondition(db, key, value)
	}

	// Get column names from struct
	columns := r.getColumnNames(out)

	// Use Model() to ensure proper table mapping with explicit column selection
	// Order by created_at DESC only (remove any default ordering by primary key)
	err = db.Model(out).Select(columns).Order("created_at DESC").First(out).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound = true
			err = nil // Clear error for "record not found" as it's an expected case
		}
	}

	return
}

func (r *BaseRepository) Find(ctx context.Context, out interface{}, conditions types.Conditions, orders ...string) (err error) {
	db := r.getDB().WithContext(ctx)

	for key, value := range conditions {
		db = r.applyWhereCondition(db, key, value)
	}

	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	} else {
		// Default order by created_at DESC if no orders specified
		db = db.Order("created_at DESC")
	}

	// Get column names from struct
	columns := r.getColumnNames(out)

	return db.Model(out).Select(columns).Find(out).Error
}

func (r *BaseRepository) Scan(ctx context.Context, tableName string, model, out interface{}, conditions types.Conditions, orders ...string) (notFound bool, err error) {
	db := r.getDB().WithContext(ctx)

	for key, value := range conditions {
		db = r.applyWhereCondition(db, key, value)
	}

	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	} else {
		// Default order by created_at DESC if no orders specified
		db = db.Order("created_at DESC")
	}

	// Get column names from model or out
	var columns []string
	if model != nil {
		columns = r.getColumnNames(model)
	} else if out != nil {
		columns = r.getColumnNames(out)
	}

	if model == nil && tableName != "" {
		if len(columns) > 0 {
			db = db.Table(tableName).Select(columns).Scan(out)
		} else {
			db = db.Table(tableName).Scan(out)
		}
	} else {
		if len(columns) > 0 {
			db = db.Model(model).Select(columns).Scan(out)
		} else {
			db = db.Model(model).Scan(out)
		}
	}

	err = db.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound = true
			err = nil // Clear error for "record not found" as it's an expected case
		}
	}

	return
}

func (r *BaseRepository) RawSQL(ctx context.Context, specifyDb *gorm.DB, query string, args ...interface{}) *gorm.DB {
	var db *gorm.DB
	if specifyDb != nil {
		db = specifyDb
	} else {
		db = r.getDB()
	}
	return db.WithContext(ctx).Raw(query, args...)
}

func (r *BaseRepository) ExecSQL(ctx context.Context, specifyDb *gorm.DB, query string, args ...interface{}) error {
	var db *gorm.DB
	if specifyDb != nil {
		db = specifyDb
	} else {
		db = r.getDB()
	}
	return db.WithContext(ctx).Exec(query, args...).Error
}

func (r *BaseRepository) IsEmpty(ctx context.Context, model interface{}) bool {
	if err := r.getDB().WithContext(ctx).Model(model).First(nil).Error; err != nil {
		return true
	}
	return false
}

// SimplePagination performs simple offset-based pagination.
// pageNumber: current page number (default: 1)
// pageSize: number of items per page (default: 10, max: 100)
// conditions: filter conditions
// orders: optional order clauses (defaults to "created_at DESC")
// preloads: pass association names (e.g. "User", "Items") to avoid N+1 queries when loading relations.
func (r *BaseRepository) SimplePagination(ctx context.Context, model, out interface{}, pageNumber, pageSize int, conditions types.Conditions, orders []string, preloads ...string) (total int64, err error) {
	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100
	}

	offset := (pageNumber - 1) * pageSize
	dataDB := r.getDB().WithContext(ctx).Model(model)

	// Apply conditions
	for key, value := range conditions {
		dataDB = r.applyWhereCondition(dataDB, key, value)
	}

	// Apply preloads
	for _, preload := range preloads {
		dataDB = dataDB.Preload(preload)
	}

	// Apply orders
	if len(orders) > 0 {
		for _, order := range orders {
			dataDB = dataDB.Order(order)
		}
	} else {
		// Default order by created_at DESC
		dataDB = dataDB.Order("created_at DESC")
	}

	// Get column names from struct
	columns := r.getColumnNames(model)

	// Fetch limit+1 items to check if there's more data (without expensive COUNT query)
	fetchLimit := pageSize + 1
	if err := dataDB.Select(columns).Limit(fetchLimit).Offset(offset).Find(out).Error; err != nil {
		return 0, err
	}

	// Check if we have more items than requested
	// Use reflection to check slice length
	val := reflect.ValueOf(out)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	hasMore := false
	if val.Kind() == reflect.Slice && val.Len() > pageSize {
		hasMore = true
		// Truncate to requested page size
		truncatedSlice := reflect.MakeSlice(val.Type(), pageSize, pageSize)
		reflect.Copy(truncatedSlice, val.Slice(0, pageSize))
		val.Set(truncatedSlice)
	}

	// Return hasMore as int64 (1 = has more, 0 = no more)
	// This allows us to pass the information back to the controller
	if hasMore {
		return 1, nil // 1 indicates has more data
	}
	return 0, nil // 0 indicates no more data
}

// getColumnNames extracts column names from struct tags using reflection.
// Results are cached by type to reduce allocations and CPU in hot path.
func (r *BaseRepository) getColumnNames(model interface{}) []string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	if cached, ok := columnCache.Load(t); ok {
		return cached.([]string)
	}
	columns := r.getColumnNamesFromType(t)
	columnCache.Store(t, columns)
	return columns
}

// getColumnNamesFromType extracts column names from a reflect.Type.
// Pre-allocates slice to reduce growth allocations.
func (r *BaseRepository) getColumnNamesFromType(t reflect.Type) []string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	n := t.NumField()
	columns := make([]string, 0, n)

	for i := 0; i < n; i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get gorm tag
		gormTag := field.Tag.Get("gorm")
		if gormTag == "" {
			// If no gorm tag and it's embedded, recurse
			if field.Anonymous {
				embeddedColumns := r.getColumnNamesFromType(field.Type)
				columns = append(columns, embeddedColumns...)
			}
			continue
		}

		// Extract column name from gorm tag
		columnName := r.extractColumnName(gormTag)
		if columnName != "" {
			columns = append(columns, columnName)
		} else if field.Anonymous {
			// Handle embedded structs
			embeddedColumns := r.getColumnNamesFromType(field.Type)
			columns = append(columns, embeddedColumns...)
		}
	}

	return columns
}

// extractColumnName extracts the column name from a gorm tag
func (r *BaseRepository) extractColumnName(gormTag string) string {
	// Parse gorm tag: "column:column_name;type:varchar(255);..."
	parts := strings.Split(gormTag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			columnName := strings.TrimPrefix(part, "column:")
			// Remove any additional attributes after column name
			if idx := strings.Index(columnName, ","); idx != -1 {
				columnName = columnName[:idx]
			}
			return strings.TrimSpace(columnName)
		}
	}
	return ""
}
