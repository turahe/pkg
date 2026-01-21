package repositories

import (
	"errors"
	"reflect"
	"strings"

	"gorm.io/gorm"

	"github.com/turahe/pkg/database"
	"github.com/turahe/pkg/types"
)

// IBaseRepository defines the base repository interface
type IBaseRepository interface {
	Create(value interface{}) error
	Save(value interface{}) error
	Updates(where interface{}, value interface{}) error
	Delete(tableName string, model interface{}, conditions types.Conditions) (count int64, err error)
	First(out interface{}, conditions types.Conditions) (notFound bool, err error)
	Find(out interface{}, conditions types.Conditions, orders ...string) error
	Scan(tableName string, model, out interface{}, conditions types.Conditions, orders ...string) (notFound bool, err error)
	RawSQL(specifyDb *gorm.DB, query string, args ...interface{}) *gorm.DB
	ExecSQL(specifyDb *gorm.DB, query string, args ...interface{}) error
	IsEmpty(model interface{}) bool
	SimplePagination(model, out interface{}, pageNumber, pageSize int, conditions types.Conditions, orders []string, preloads ...string) (total int64, err error)
}

// BaseRepository implements IBaseRepository using GORM
type BaseRepository struct {
	useSiteDB bool // Flag to determine which database to use
}

// NewBaseRepository creates a new base repository instance using main database
func NewBaseRepository() IBaseRepository {
	return &BaseRepository{
		useSiteDB: false,
	}
}

// NewSiteBaseRepository creates a new base repository instance using site database
func NewSiteBaseRepository() IBaseRepository {
	return &BaseRepository{
		useSiteDB: true,
	}
}

// getDB returns the appropriate database connection
func (r *BaseRepository) getDB() *gorm.DB {
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

	// If there are multiple placeholders and value is not already a slice, duplicate the value
	if placeholderCount > 1 {
		// Check if value is already a slice/array
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			// Value is already a slice, use it as-is
			var args []interface{}
			for i := 0; i < val.Len(); i++ {
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

func (r *BaseRepository) Create(value interface{}) error {
	return r.getDB().Create(value).Error
}

func (r *BaseRepository) Save(value interface{}) error {
	return r.getDB().Save(value).Error
}

func (r *BaseRepository) Updates(where interface{}, value interface{}) error {
	return r.getDB().Model(where).Updates(value).Error
}

func (r *BaseRepository) Delete(tableName string, model interface{}, conditions types.Conditions) (count int64, err error) {
	db := r.getDB()

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

func (r *BaseRepository) First(out interface{}, conditions types.Conditions) (notFound bool, err error) {
	db := r.getDB()

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

func (r *BaseRepository) Find(out interface{}, conditions types.Conditions, orders ...string) (err error) {
	db := r.getDB()

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

func (r *BaseRepository) Scan(tableName string, model, out interface{}, conditions types.Conditions, orders ...string) (notFound bool, err error) {
	db := r.getDB()

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

func (r *BaseRepository) RawSQL(specifyDb *gorm.DB, query string, args ...interface{}) *gorm.DB {
	var db *gorm.DB

	if specifyDb != nil {
		db = specifyDb
	} else {
		db = r.getDB()
	}

	return db.Raw(query, args...)
}

func (r *BaseRepository) ExecSQL(specifyDb *gorm.DB, query string, args ...interface{}) error {
	var db *gorm.DB

	if specifyDb != nil {
		db = specifyDb
	} else {
		db = r.getDB()
	}

	err := db.Exec(query, args...).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *BaseRepository) IsEmpty(model interface{}) bool {
	if err := r.getDB().Model(model).First(nil).Error; err != nil {
		return true
	}

	return false
}

// SimplePagination performs simple offset-based pagination
// pageNumber: current page number (default: 1)
// pageSize: number of items per page (default: 10, max: 100)
// conditions: filter conditions
// orders: optional order clauses (defaults to "created_at DESC")
// preloads: optional preload relationships
func (r *BaseRepository) SimplePagination(model, out interface{}, pageNumber, pageSize int, conditions types.Conditions, orders []string, preloads ...string) (total int64, err error) {
	maxPageSize := 100

	// Validate and set defaults
	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	offset := (pageNumber - 1) * pageSize

	// Build data query (removed COUNT query for performance - fetch limit+1 to determine if there's more data)
	dataDB := r.getDB().Model(model)

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

// getColumnNames extracts column names from struct tags using reflection
func (r *BaseRepository) getColumnNames(model interface{}) []string {
	var columns []string
	t := reflect.TypeOf(model)

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Handle slice types
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	// Must be a struct
	if t.Kind() != reflect.Struct {
		return columns
	}

	// Iterate through all fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get gorm tag
		gormTag := field.Tag.Get("gorm")
		if gormTag == "" {
			// If no gorm tag, check if it's an embedded struct
			if field.Anonymous {
				// Recursively get columns from embedded struct
				embeddedColumns := r.getColumnNamesFromType(field.Type)
				columns = append(columns, embeddedColumns...)
			}
			continue
		}

		// Extract column name from gorm tag
		// Format: gorm:"column:column_name;..."
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

// getColumnNamesFromType extracts column names from a reflect.Type
func (r *BaseRepository) getColumnNamesFromType(t reflect.Type) []string {
	var columns []string

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Must be a struct
	if t.Kind() != reflect.Struct {
		return columns
	}

	// Iterate through all fields
	for i := 0; i < t.NumField(); i++ {
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
