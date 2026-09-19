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
	db        *gorm.DB // optional: when set, used instead of database.GetDB()
	siteDB    *gorm.DB // optional: when set, used instead of database.GetDBSite()
	useSiteDB bool     // when db/siteDB are nil, use GetDBSite() when true else GetDB()
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

func (r *BaseRepository) Delete(ctx context.Context, tableName string, model interface{}, conditions types.Conditions) (int64, error) {
	db := r.getDB().WithContext(ctx)

	for key, value := range conditions {
		db = r.applyWhereCondition(db, key, value)
	}

	if model == nil && tableName != "" {
		db = db.Table(tableName).Delete(nil)
	} else {
		db = db.Delete(model)
	}

	err := db.Error
	if err != nil {
		return 0, err
	}

	return db.RowsAffected, nil
}

func (r *BaseRepository) First(ctx context.Context, out interface{}, conditions types.Conditions) (bool, error) {
	db := r.getDB().WithContext(ctx)

	for key, value := range conditions {
		db = r.applyWhereCondition(db, key, value)
	}

	// Get column names from struct
	columns := r.getColumnNames(out)

	// Use Model() to ensure proper table mapping with explicit column selection
	// Order by created_at DESC only (remove any default ordering by primary key)
	err := db.Model(out).Select(columns).Order("created_at DESC").First(out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}
	return false, err
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

// ScanOptions groups parameters for Scan / ScanWith.
type ScanOptions struct {
	TableName  string
	Model      interface{}
	Out        interface{}
	Conditions types.Conditions
	Orders     []string
}

// ScanWith runs a Scan using ScanOptions (preferred for new call sites).
func (r *BaseRepository) ScanWith(ctx context.Context, opts ScanOptions) (bool, error) {
	db := r.getDB().WithContext(ctx)

	for key, value := range opts.Conditions {
		db = r.applyWhereCondition(db, key, value)
	}

	if len(opts.Orders) > 0 {
		for _, order := range opts.Orders {
			db = db.Order(order)
		}
	} else {
		db = db.Order("created_at DESC")
	}

	columns := []string{}
	if opts.Model != nil {
		columns = r.getColumnNames(opts.Model)
	} else if opts.Out != nil {
		columns = r.getColumnNames(opts.Out)
	}

	useTable := opts.Model == nil && opts.TableName != ""
	if useTable {
		db = db.Table(opts.TableName)
	} else {
		db = db.Model(opts.Model)
	}
	if len(columns) > 0 {
		db = db.Select(columns)
	}
	err := db.Scan(opts.Out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}
	return false, err
}

// Scan runs a Scan. Prefer ScanWith with ScanOptions for new call sites.
func (r *BaseRepository) Scan(ctx context.Context, tableName string, model, out interface{}, conditions types.Conditions, orders ...string) (bool, error) {
	return r.ScanWith(ctx, ScanOptions{
		TableName:  tableName,
		Model:      model,
		Out:        out,
		Conditions: conditions,
		Orders:     orders,
	})
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

// PaginationOptions groups parameters for SimplePagination / Paginate.
type PaginationOptions struct {
	Model      interface{}
	Out        interface{}
	PageNumber int
	PageSize   int
	Conditions types.Conditions
	Orders     []string
	Preloads   []string
}

// Paginate performs offset-based pagination using PaginationOptions (preferred for new call sites).
// Returns 1 when more pages exist, 0 otherwise.
func (r *BaseRepository) Paginate(ctx context.Context, opts PaginationOptions) (int64, error) {
	pageNumber := opts.PageNumber
	pageSize := opts.PageSize
	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100
	}

	offset := (pageNumber - 1) * pageSize
	dataDB := r.getDB().WithContext(ctx).Model(opts.Model)

	for key, value := range opts.Conditions {
		dataDB = r.applyWhereCondition(dataDB, key, value)
	}
	for _, preload := range opts.Preloads {
		dataDB = dataDB.Preload(preload)
	}
	if len(opts.Orders) > 0 {
		for _, order := range opts.Orders {
			dataDB = dataDB.Order(order)
		}
	} else {
		dataDB = dataDB.Order("created_at DESC")
	}

	columns := r.getColumnNames(opts.Model)
	fetchLimit := pageSize + 1
	if err := dataDB.Select(columns).Limit(fetchLimit).Offset(offset).Find(opts.Out).Error; err != nil {
		return 0, err
	}

	val := reflect.ValueOf(opts.Out)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	var hasMore bool
	if val.Kind() == reflect.Slice && val.Len() > pageSize {
		hasMore = true
		truncatedSlice := reflect.MakeSlice(val.Type(), pageSize, pageSize)
		reflect.Copy(truncatedSlice, val.Slice(0, pageSize))
		val.Set(truncatedSlice)
	}
	if hasMore {
		return 1, nil
	}
	return 0, nil
}

// SimplePagination performs offset-based pagination. Prefer Paginate with PaginationOptions.
func (r *BaseRepository) SimplePagination(
	ctx context.Context,
	model, out interface{},
	pageNumber, pageSize int,
	conditions types.Conditions,
	orders []string,
	preloads ...string,
) (int64, error) {
	return r.Paginate(ctx, PaginationOptions{
		Model:      model,
		Out:        out,
		PageNumber: pageNumber,
		PageSize:   pageSize,
		Conditions: conditions,
		Orders:     orders,
		Preloads:   preloads,
	})
}

// getColumnNames extracts column names from struct tags using reflection.
// Results are cached by type to reduce allocations and CPU in hot path.
func (r *BaseRepository) getColumnNames(model interface{}) []string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Pointer {
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
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return []string{}
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
