# Building a Comprehensive Database Diff Tool in Go

I'll help you build a robust database diff tool that can compare two PostgreSQL database versions and generate migration scripts. This will be a significant addition to your `dbman` CLI.

---

## Architecture Overview

```
┌─────────────────┐      ┌─────────────────┐
│   Source DB     │      │   Target DB     │
│   (Version A)   │      │   (Version B)   │
└────────┬────────┘      └────────┬────────┘
         │                        │
         └────────┐      ┌────────┘
                  │      │
            ┌─────▼──────▼─────┐
            │   Schema Reader  │
            │  (Extract DDL)   │
            └─────────┬─────────┘
                      │
            ┌─────────▼─────────┐
            │   Diff Engine     │
            │  (Compare Schema) │
            └─────────┬─────────┘
                      │
            ┌─────────▼─────────┐
            │  Script Generator │
            │  (UP/DOWN SQL)    │
            └───────────────────┘
```

---

## Implementation

### 1. Project Structure

```
dbman/
├── cmd/
│   └── dbman/
│       └── main.go
├── pkg/
│   ├── diff/
│   │   ├── differ.go          # Main diff orchestrator
│   │   ├── schema.go          # Schema structures
│   │   ├── reader.go          # Database schema reader
│   │   ├── comparator.go      # Comparison logic
│   │   └── generator.go       # Script generator
│   └── db/
│       └── connection.go
└── go.mod
```

### 2. Schema Structures (`pkg/diff/schema.go`)

```go
package diff

import "time"

// DatabaseSchema represents the complete schema of a database
type DatabaseSchema struct {
	Tables      map[string]*Table
	Sequences   map[string]*Sequence
	Enums       map[string]*Enum
	Functions   map[string]*Function
	Indexes     map[string]*Index
	Constraints map[string]*Constraint
	Views       map[string]*View
	Triggers    map[string]*Trigger
}

// Table represents a database table
type Table struct {
	Name    string
	Schema  string
	Columns map[string]*Column
	Comment string
}

// Column represents a table column
type Column struct {
	Name            string
	DataType        string
	IsNullable      bool
	DefaultValue    *string
	CharMaxLength   *int
	NumericPrecision *int
	NumericScale    *int
	OrdinalPosition int
	Comment         *string
}

// Index represents a database index
type Index struct {
	Name       string
	TableName  string
	Schema     string
	Columns    []string
	IsUnique   bool
	IsPrimary  bool
	Definition string
}

// Constraint represents a database constraint
type Constraint struct {
	Name           string
	TableName      string
	Schema         string
	Type           string // CHECK, FOREIGN KEY, UNIQUE, PRIMARY KEY
	Definition     string
	ForeignTable   *string
	ForeignColumns []string
	OnDelete       *string
	OnUpdate       *string
}

// Sequence represents a database sequence
type Sequence struct {
	Name      string
	Schema    string
	StartValue int64
	Increment int64
}

// Enum represents a custom enum type
type Enum struct {
	Name   string
	Schema string
	Values []string
}

// Function represents a database function
type Function struct {
	Name       string
	Schema     string
	Definition string
	ReturnType string
	Language   string
}

// View represents a database view
type View struct {
	Name       string
	Schema     string
	Definition string
}

// Trigger represents a database trigger
type Trigger struct {
	Name       string
	TableName  string
	Schema     string
	Timing     string // BEFORE, AFTER, INSTEAD OF
	Event      string // INSERT, UPDATE, DELETE
	Definition string
}

// DiffResult holds the differences between two schemas
type DiffResult struct {
	CreatedTables   []*Table
	DroppedTables   []*Table
	AlteredTables   []*TableDiff
	
	CreatedIndexes  []*Index
	DroppedIndexes  []*Index
	
	CreatedConstraints []*Constraint
	DroppedConstraints []*Constraint
	
	CreatedSequences []*Sequence
	DroppedSequences []*Sequence
	
	CreatedEnums    []*Enum
	DroppedEnums    []*Enum
	AlteredEnums    []*EnumDiff
	
	CreatedFunctions []*Function
	DroppedFunctions []*Function
	AlteredFunctions []*FunctionDiff
	
	CreatedViews    []*View
	DroppedViews    []*View
	AlteredViews    []*ViewDiff
	
	CreatedTriggers []*Trigger
	DroppedTriggers []*Trigger
}

// TableDiff represents changes to a table
type TableDiff struct {
	TableName      string
	AddedColumns   []*Column
	DroppedColumns []*Column
	AlteredColumns []*ColumnDiff
}

// ColumnDiff represents changes to a column
type ColumnDiff struct {
	Name            string
	OldColumn       *Column
	NewColumn       *Column
	TypeChanged     bool
	NullableChanged bool
	DefaultChanged  bool
}

// EnumDiff represents changes to an enum
type EnumDiff struct {
	Name         string
	AddedValues  []string
	RemovedValues []string
}

// FunctionDiff represents changes to a function
type FunctionDiff struct {
	Name          string
	OldDefinition string
	NewDefinition string
}

// ViewDiff represents changes to a view
type ViewDiff struct {
	Name          string
	OldDefinition string
	NewDefinition string
}

// GeneratedScript contains the migration scripts
type GeneratedScript struct {
	UpScript     string
	DownScript   string
	GeneratedAt  time.Time
	Description  string
	HasBreaking  bool
	Warnings     []string
}
```

### 3. Schema Reader (`pkg/diff/reader.go`)

```go
package diff

import (
	"database/sql"
	"fmt"
	"strings"
)

// SchemaReader reads database schema information
type SchemaReader struct {
	db     *sql.DB
	schema string
}

// NewSchemaReader creates a new schema reader
func NewSchemaReader(db *sql.DB, schema string) *SchemaReader {
	if schema == "" {
		schema = "public"
	}
	return &SchemaReader{
		db:     db,
		schema: schema,
	}
}

// ReadSchema reads the complete database schema
func (r *SchemaReader) ReadSchema() (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Tables:      make(map[string]*Table),
		Sequences:   make(map[string]*Sequence),
		Enums:       make(map[string]*Enum),
		Functions:   make(map[string]*Function),
		Indexes:     make(map[string]*Index),
		Constraints: make(map[string]*Constraint),
		Views:       make(map[string]*View),
		Triggers:    make(map[string]*Trigger),
	}

	var err error
	
	if err = r.readTables(schema); err != nil {
		return nil, fmt.Errorf("reading tables: %w", err)
	}
	
	if err = r.readIndexes(schema); err != nil {
		return nil, fmt.Errorf("reading indexes: %w", err)
	}
	
	if err = r.readConstraints(schema); err != nil {
		return nil, fmt.Errorf("reading constraints: %w", err)
	}
	
	if err = r.readSequences(schema); err != nil {
		return nil, fmt.Errorf("reading sequences: %w", err)
	}
	
	if err = r.readEnums(schema); err != nil {
		return nil, fmt.Errorf("reading enums: %w", err)
	}
	
	if err = r.readFunctions(schema); err != nil {
		return nil, fmt.Errorf("reading functions: %w", err)
	}
	
	if err = r.readViews(schema); err != nil {
		return nil, fmt.Errorf("reading views: %w", err)
	}
	
	if err = r.readTriggers(schema); err != nil {
		return nil, fmt.Errorf("reading triggers: %w", err)
	}

	return schema, nil
}

// readTables reads all tables and their columns
func (r *SchemaReader) readTables(schema *DatabaseSchema) error {
	query := `
		SELECT 
			t.table_name,
			t.table_schema,
			obj_description((t.table_schema||'.'||t.table_name)::regclass, 'pg_class') as table_comment
		FROM information_schema.tables t
		WHERE t.table_schema = $1
		AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_name
	`

	rows, err := r.db.Query(query, r.schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName, tableSchema string
		var comment sql.NullString
		
		if err := rows.Scan(&tableName, &tableSchema, &comment); err != nil {
			return err
		}

		table := &Table{
			Name:    tableName,
			Schema:  tableSchema,
			Columns: make(map[string]*Column),
		}
		
		if comment.Valid {
			table.Comment = comment.String
		}

		// Read columns for this table
		if err := r.readColumns(table); err != nil {
			return err
		}

		key := fmt.Sprintf("%s.%s", tableSchema, tableName)
		schema.Tables[key] = table
	}

	return rows.Err()
}

// readColumns reads columns for a specific table
func (r *SchemaReader) readColumns(table *Table) error {
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale,
			c.ordinal_position,
			pgd.description
		FROM information_schema.columns c
		LEFT JOIN pg_catalog.pg_statio_all_tables st 
			ON c.table_schema = st.schemaname 
			AND c.table_name = st.relname
		LEFT JOIN pg_catalog.pg_description pgd 
			ON pgd.objoid = st.relid 
			AND pgd.objsubid = c.ordinal_position
		WHERE c.table_schema = $1
		AND c.table_name = $2
		ORDER BY c.ordinal_position
	`

	rows, err := r.db.Query(query, table.Schema, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name                                 string
			dataType                             string
			isNullable                           string
			defaultValue                         sql.NullString
			charMaxLength, numPrecision, numScale sql.NullInt64
			ordinalPosition                      int
			comment                              sql.NullString
		)

		err := rows.Scan(
			&name,
			&dataType,
			&isNullable,
			&defaultValue,
			&charMaxLength,
			&numPrecision,
			&numScale,
			&ordinalPosition,
			&comment,
		)
		if err != nil {
			return err
		}

		col := &Column{
			Name:            name,
			DataType:        dataType,
			IsNullable:      isNullable == "YES",
			OrdinalPosition: ordinalPosition,
		}

		if defaultValue.Valid {
			col.DefaultValue = &defaultValue.String
		}
		if charMaxLength.Valid {
			length := int(charMaxLength.Int64)
			col.CharMaxLength = &length
		}
		if numPrecision.Valid {
			precision := int(numPrecision.Int64)
			col.NumericPrecision = &precision
		}
		if numScale.Valid {
			scale := int(numScale.Int64)
			col.NumericScale = &scale
		}
		if comment.Valid {
			col.Comment = &comment.String
		}

		table.Columns[name] = col
	}

	return rows.Err()
}

// readIndexes reads all indexes
func (r *SchemaReader) readIndexes(schema *DatabaseSchema) error {
	query := `
		SELECT
			i.indexname,
			i.tablename,
			i.schemaname,
			ix.indisunique,
			ix.indisprimary,
			pg_get_indexdef(ix.indexrelid) as definition,
			array_agg(a.attname ORDER BY a.attnum) as columns
		FROM pg_indexes i
		JOIN pg_class c ON c.relname = i.indexname
		JOIN pg_index ix ON ix.indexrelid = c.oid
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE i.schemaname = $1
		GROUP BY i.indexname, i.tablename, i.schemaname, ix.indisunique, ix.indisprimary, ix.indexrelid
		ORDER BY i.tablename, i.indexname
	`

	rows, err := r.db.Query(query, r.schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name, tableName, schemaName, definition string
			isUnique, isPrimary                     bool
			columns                                 string
		)

		if err := rows.Scan(&name, &tableName, &schemaName, &isUnique, &isPrimary, &definition, &columns); err != nil {
			return err
		}

		// Parse column array string (format: {col1,col2})
		colSlice := parsePostgresArray(columns)

		index := &Index{
			Name:       name,
			TableName:  tableName,
			Schema:     schemaName,
			Columns:    colSlice,
			IsUnique:   isUnique,
			IsPrimary:  isPrimary,
			Definition: definition,
		}

		key := fmt.Sprintf("%s.%s", schemaName, name)
		schema.Indexes[key] = index
	}

	return rows.Err()
}

// readConstraints reads all constraints
func (r *SchemaReader) readConstraints(schema *DatabaseSchema) error {
	query := `
		SELECT
			con.conname as constraint_name,
			rel.relname as table_name,
			ns.nspname as schema_name,
			con.contype as constraint_type,
			pg_get_constraintdef(con.oid) as definition,
			frel.relname as foreign_table,
			con.confdeltype as on_delete,
			con.confupdtype as on_update
		FROM pg_constraint con
		JOIN pg_class rel ON rel.oid = con.conrelid
		JOIN pg_namespace ns ON ns.oid = rel.relnamespace
		LEFT JOIN pg_class frel ON frel.oid = con.confrelid
		WHERE ns.nspname = $1
		AND con.contype IN ('c', 'f', 'u', 'p')
		ORDER BY rel.relname, con.conname
	`

	rows, err := r.db.Query(query, r.schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name, tableName, schemaName, conType, definition string
			foreignTable                                     sql.NullString
			onDelete, onUpdate                               sql.NullString
		)

		err := rows.Scan(
			&name,
			&tableName,
			&schemaName,
			&conType,
			&definition,
			&foreignTable,
			&onDelete,
			&onUpdate,
		)
		if err != nil {
			return err
		}

		constraint := &Constraint{
			Name:       name,
			TableName:  tableName,
			Schema:     schemaName,
			Type:       mapConstraintType(conType),
			Definition: definition,
		}

		if foreignTable.Valid {
			constraint.ForeignTable = &foreignTable.String
		}
		if onDelete.Valid {
			action := mapForeignKeyAction(onDelete.String)
			constraint.OnDelete = &action
		}
		if onUpdate.Valid {
			action := mapForeignKeyAction(onUpdate.String)
			constraint.OnUpdate = &action
		}

		key := fmt.Sprintf("%s.%s", schemaName, name)
		schema.Constraints[key] = constraint
	}

	return rows.Err()
}

// readSequences reads all sequences
func (r *SchemaReader) readSequences(schema *DatabaseSchema) error {
	query := `
		SELECT
			sequence_name,
			sequence_schema,
			start_value::bigint,
			increment::bigint
		FROM information_schema.sequences
		WHERE sequence_schema = $1
		ORDER BY sequence_name
	`

	rows, err := r.db.Query(query, r.schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, schemaName string
		var startValue, increment int64

		if err := rows.Scan(&name, &schemaName, &startValue, &increment); err != nil {
			return err
		}

		seq := &Sequence{
			Name:       name,
			Schema:     schemaName,
			StartValue: startValue,
			Increment:  increment,
		}

		key := fmt.Sprintf("%s.%s", schemaName, name)
		schema.Sequences[key] = seq
	}

	return rows.Err()
}

// readEnums reads all enum types
func (r *SchemaReader) readEnums(schema *DatabaseSchema) error {
	query := `
		SELECT
			t.typname as enum_name,
			n.nspname as schema_name,
			array_agg(e.enumlabel ORDER BY e.enumsortorder) as enum_values
		FROM pg_type t
		JOIN pg_enum e ON t.oid = e.enumtypid
		JOIN pg_namespace n ON t.typnamespace = n.oid
		WHERE n.nspname = $1
		GROUP BY t.typname, n.nspname
		ORDER BY t.typname
	`

	rows, err := r.db.Query(query, r.schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, schemaName, values string

		if err := rows.Scan(&name, &schemaName, &values); err != nil {
			return err
		}

		enum := &Enum{
			Name:   name,
			Schema: schemaName,
			Values: parsePostgresArray(values),
		}

		key := fmt.Sprintf("%s.%s", schemaName, name)
		schema.Enums[key] = enum
	}

	return rows.Err()
}

// readFunctions reads all functions
func (r *SchemaReader) readFunctions(schema *DatabaseSchema) error {
	query := `
		SELECT
			p.proname as function_name,
			n.nspname as schema_name,
			pg_get_functiondef(p.oid) as definition,
			pg_get_function_result(p.oid) as return_type,
			l.lanname as language
		FROM pg_proc p
		JOIN pg_namespace n ON p.pronamespace = n.oid
		JOIN pg_language l ON p.prolang = l.oid
		WHERE n.nspname = $1
		AND p.prokind = 'f'
		ORDER BY p.proname
	`

	rows, err := r.db.Query(query, r.schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, schemaName, definition, returnType, language string

		if err := rows.Scan(&name, &schemaName, &definition, &returnType, &language); err != nil {
			return err
		}

		function := &Function{
			Name:       name,
			Schema:     schemaName,
			Definition: definition,
			ReturnType: returnType,
			Language:   language,
		}

		key := fmt.Sprintf("%s.%s", schemaName, name)
		schema.Functions[key] = function
	}

	return rows.Err()
}

// readViews reads all views
func (r *SchemaReader) readViews(schema *DatabaseSchema) error {
	query := `
		SELECT
			table_name,
			table_schema,
			view_definition
		FROM information_schema.views
		WHERE table_schema = $1
		ORDER BY table_name
	`

	rows, err := r.db.Query(query, r.schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, schemaName, definition string

		if err := rows.Scan(&name, &schemaName, &definition); err != nil {
			return err
		}

		view := &View{
			Name:       name,
			Schema:     schemaName,
			Definition: definition,
		}

		key := fmt.Sprintf("%s.%s", schemaName, name)
		schema.Views[key] = view
	}

	return rows.Err()
}

// readTriggers reads all triggers
func (r *SchemaReader) readTriggers(schema *DatabaseSchema) error {
	query := `
		SELECT
			t.tgname as trigger_name,
			c.relname as table_name,
			n.nspname as schema_name,
			CASE t.tgtype & 2
				WHEN 0 THEN 'AFTER'
				ELSE 'BEFORE'
			END as timing,
			CASE t.tgtype & 28
				WHEN 4 THEN 'INSERT'
				WHEN 8 THEN 'DELETE'
				WHEN 16 THEN 'UPDATE'
				ELSE 'UNKNOWN'
			END as event,
			pg_get_triggerdef(t.oid) as definition
		FROM pg_trigger t
		JOIN pg_class c ON t.tgrelid = c.oid
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE n.nspname = $1
		AND NOT t.tgisinternal
		ORDER BY c.relname, t.tgname
	`

	rows, err := r.db.Query(query, r.schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, tableName, schemaName, timing, event, definition string

		if err := rows.Scan(&name, &tableName, &schemaName, &timing, &event, &definition); err != nil {
			return err
		}

		trigger := &Trigger{
			Name:       name,
			TableName:  tableName,
			Schema:     schemaName,
			Timing:     timing,
			Event:      event,
			Definition: definition,
		}

		key := fmt.Sprintf("%s.%s.%s", schemaName, tableName, name)
		schema.Triggers[key] = trigger
	}

	return rows.Err()
}

// Helper functions

func parsePostgresArray(s string) []string {
	s = strings.Trim(s, "{}")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

func mapConstraintType(t string) string {
	switch t {
	case "c":
		return "CHECK"
	case "f":
		return "FOREIGN KEY"
	case "p":
		return "PRIMARY KEY"
	case "u":
		return "UNIQUE"
	default:
		return "UNKNOWN"
	}
}

func mapForeignKeyAction(action string) string {
	switch action {
	case "a":
		return "NO ACTION"
	case "r":
		return "RESTRICT"
	case "c":
		return "CASCADE"
	case "n":
		return "SET NULL"
	case "d":
		return "SET DEFAULT"
	default:
		return "NO ACTION"
	}
}
```

### 4. Comparator (`pkg/diff/comparator.go`)

```go
package diff

import (
	"fmt"
)

// Comparator compares two database schemas
type Comparator struct{}

// NewComparator creates a new comparator
func NewComparator() *Comparator {
	return &Comparator{}
}

// Compare compares source and target schemas and returns differences
func (c *Comparator) Compare(source, target *DatabaseSchema) *DiffResult {
	result := &DiffResult{}

	c.compareTables(source, target, result)
	c.compareIndexes(source, target, result)
	c.compareConstraints(source, target, result)
	c.compareSequences(source, target, result)
	c.compareEnums(source, target, result)
	c.compareFunctions(source, target, result)
	c.compareViews(source, target, result)
	c.compareTriggers(source, target, result)

	return result
}

// compareTables compares tables between schemas
func (c *Comparator) compareTables(source, target *DatabaseSchema, result *DiffResult) {
	// Find created tables (in target but not in source)
	for key, targetTable := range target.Tables {
		if _, exists := source.Tables[key]; !exists {
			result.CreatedTables = append(result.CreatedTables, targetTable)
		}
	}

	// Find dropped tables (in source but not in target)
	for key, sourceTable := range source.Tables {
		if _, exists := target.Tables[key]; !exists {
			result.DroppedTables = append(result.DroppedTables, sourceTable)
		}
	}

	// Find altered tables (exist in both)
	for key, sourceTable := range source.Tables {
		if targetTable, exists := target.Tables[key]; exists {
			tableDiff := c.compareTableStructure(sourceTable, targetTable)
			if tableDiff != nil {
				result.AlteredTables = append(result.AlteredTables, tableDiff)
			}
		}
	}
}

// compareTableStructure compares the structure of two tables
func (c *Comparator) compareTableStructure(source, target *Table) *TableDiff {
	diff := &TableDiff{
		TableName: target.Name,
	}

	hasChanges := false

	// Find added columns
	for colName, targetCol := range target.Columns {
		if _, exists := source.Columns[colName]; !exists {
			diff.AddedColumns = append(diff.AddedColumns, targetCol)
			hasChanges = true
		}
	}

	// Find dropped columns
	for colName, sourceCol := range source.Columns {
		if _, exists := target.Columns[colName]; !exists {
			diff.DroppedColumns = append(diff.DroppedColumns, sourceCol)
			hasChanges = true
		}
	}

	// Find altered columns
	for colName, sourceCol := range source.Columns {
		if targetCol, exists := target.Columns[colName]; exists {
			colDiff := c.compareColumns(sourceCol, targetCol)
			if colDiff != nil {
				diff.AlteredColumns = append(diff.AlteredColumns, colDiff)
				hasChanges = true
			}
		}
	}

	if !hasChanges {
		return nil
	}

	return diff
}

// compareColumns compares two columns
func (c *Comparator) compareColumns(source, target *Column) *ColumnDiff {
	diff := &ColumnDiff{
		Name:      target.Name,
		OldColumn: source,
		NewColumn: target,
	}

	hasChanges := false

	// Check data type
	if source.DataType != target.DataType {
		diff.TypeChanged = true
		hasChanges = true
	}

	// Check nullable
	if source.IsNullable != target.IsNullable {
		diff.NullableChanged = true
		hasChanges = true
	}

	// Check default value
	if !compareNullableStrings(source.DefaultValue, target.DefaultValue) {
		diff.DefaultChanged = true
		hasChanges = true
	}

	if !hasChanges {
		return nil
	}

	return diff
}

// compareIndexes compares indexes
func (c *Comparator) compareIndexes(source, target *DatabaseSchema, result *DiffResult) {
	// Created indexes
	for key, targetIndex := range target.Indexes {
		if _, exists := source.Indexes[key]; !exists {
			// Skip primary key indexes (they're handled by constraints)
			if !targetIndex.IsPrimary {
				result.CreatedIndexes = append(result.CreatedIndexes, targetIndex)
			}
		}
	}

	// Dropped indexes
	for key, sourceIndex := range source.Indexes {
		if _, exists := target.Indexes[key]; !exists {
			if !sourceIndex.IsPrimary {
				result.DroppedIndexes = append(result.DroppedIndexes, sourceIndex)
			}
		}
	}
}

// compareConstraints compares constraints
func (c *Comparator) compareConstraints(source, target *DatabaseSchema, result *DiffResult) {
	// Created constraints
	for key, targetConstraint := range target.Constraints {
		if _, exists := source.Constraints[key]; !exists {
			result.CreatedConstraints = append(result.CreatedConstraints, targetConstraint)
		}
	}

	// Dropped constraints
	for key, sourceConstraint := range source.Constraints {
		if _, exists := target.Constraints[key]; !exists {
			result.DroppedConstraints = append(result.DroppedConstraints, sourceConstraint)
		}
	}
}

// compareSequences compares sequences
func (c *Comparator) compareSequences(source, target *DatabaseSchema, result *DiffResult) {
	for key, targetSeq := range target.Sequences {
		if _, exists := source.Sequences[key]; !exists {
			result.CreatedSequences = append(result.CreatedSequences, targetSeq)
		}
	}

	for key, sourceSeq := range source.Sequences {
		if _, exists := target.Sequences[key]; !exists {
			result.DroppedSequences = append(result.DroppedSequences, sourceSeq)
		}
	}
}

// compareEnums compares enum types
func (c *Comparator) compareEnums(source, target *DatabaseSchema, result *DiffResult) {
	// Created enums
	for key, targetEnum := range target.Enums {
		if _, exists := source.Enums[key]; !exists {
			result.CreatedEnums = append(result.CreatedEnums, targetEnum)
		}
	}

	// Dropped enums
	for key, sourceEnum := range source.Enums {
		if _, exists := target.Enums[key]; !exists {
			result.DroppedEnums = append(result.DroppedEnums, sourceEnum)
		}
	}

	// Altered enums
	for key, sourceEnum := range source.Enums {
		if targetEnum, exists := target.Enums[key]; exists {
			enumDiff := c.compareEnumValues(sourceEnum, targetEnum)
			if enumDiff != nil {
				result.AlteredEnums = append(result.AlteredEnums, enumDiff)
			}
		}
	}
}

// compareEnumValues compares enum values
func (c *Comparator) compareEnumValues(source, target *Enum) *EnumDiff {
	diff := &EnumDiff{
		Name: target.Name,
	}

	sourceMap := make(map[string]bool)
	for _, val := range source.Values {
		sourceMap[val] = true
	}

	targetMap := make(map[string]bool)
	for _, val := range target.Values {
		targetMap[val] = true
	}

	hasChanges := false

	// Find added values
	for _, val := range target.Values {
		if !sourceMap[val] {
			diff.AddedValues = append(diff.AddedValues, val)
			hasChanges = true
		}
	}

	// Find removed values
	for _, val := range source.Values {
		if !targetMap[val] {
			diff.RemovedValues = append(diff.RemovedValues, val)
			hasChanges = true
		}
	}

	if !hasChanges {
		return nil
	}

	return diff
}

// compareFunctions compares functions
func (c *Comparator) compareFunctions(source, target *DatabaseSchema, result *DiffResult) {
	// Created functions
	for key, targetFunc := range target.Functions {
		if _, exists := source.Functions[key]; !exists {
			result.CreatedFunctions = append(result.CreatedFunctions, targetFunc)
		}
	}

	// Dropped functions
	for key, sourceFunc := range source.Functions {
		if _, exists := target.Functions[key]; !exists {
			result.DroppedFunctions = append(result.DroppedFunctions, sourceFunc)
		}
	}

	// Altered functions
	for key, sourceFunc := range source.Functions {
		if targetFunc, exists := target.Functions[key]; exists {
			if sourceFunc.Definition != targetFunc.Definition {
				result.AlteredFunctions = append(result.AlteredFunctions, &FunctionDiff{
					Name:          targetFunc.Name,
					OldDefinition: sourceFunc.Definition,
					NewDefinition: targetFunc.Definition,
				})
			}
		}
	}
}

// compareViews compares views
func (c *Comparator) compareViews(source, target *DatabaseSchema, result *DiffResult) {
	// Created views
	for key, targetView := range target.Views {
		if _, exists := source.Views[key]; !exists {
			result.CreatedViews = append(result.CreatedViews, targetView)
		}
	}

	// Dropped views
	for key, sourceView := range source.Views {
		if _, exists := target.Views[key]; !exists {
			result.DroppedViews = append(result.DroppedViews, sourceView)
		}
	}

	// Altered views
	for key, sourceView := range source.Views {
		if targetView, exists := target.Views[key]; exists {
			if sourceView.Definition != targetView.Definition {
				result.AlteredViews = append(result.AlteredViews, &ViewDiff{
					Name:          targetView.Name,
					OldDefinition: sourceView.Definition,
					NewDefinition: targetView.Definition,
				})
			}
		}
	}
}

// compareTriggers compares triggers
func (c *Comparator) compareTriggers(source, target *DatabaseSchema, result *DiffResult) {
	for key, targetTrigger := range target.Triggers {
		if _, exists := source.Triggers[key]; !exists {
			result.CreatedTriggers = append(result.CreatedTriggers, targetTrigger)
		}
	}

	for key, sourceTrigger := range source.Triggers {
		if _, exists := target.Triggers[key]; !exists {
			result.DroppedTriggers = append(result.DroppedTriggers, sourceTrigger)
		}
	}
}

// Helper functions

func compareNullableStrings(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
```

### 5. Script Generator (`pkg/diff/generator.go`)

```go
package diff

import (
	"fmt"
	"strings"
	"time"
)

// Generator generates migration scripts from diff results
type Generator struct {
	schema string
}

// NewGenerator creates a new script generator
func NewGenerator(schema string) *Generator {
	if schema == "" {
		schema = "public"
	}
	return &Generator{schema: schema}
}

// Generate creates UP and DOWN migration scripts
func (g *Generator) Generate(diff *DiffResult, description string) *GeneratedScript {
	script := &GeneratedScript{
		GeneratedAt: time.Now(),
		Description: description,
		Warnings:    []string{},
	}

	var upBuilder, downBuilder strings.Builder

	// Header
	upBuilder.WriteString(fmt.Sprintf("-- Migration: %s\n", description))
	upBuilder.WriteString(fmt.Sprintf("-- Generated: %s\n", script.GeneratedAt.Format(time.RFC3339)))
	upBuilder.WriteString("-- UP Migration\n\n")

	downBuilder.WriteString(fmt.Sprintf("-- Migration: %s\n", description))
	downBuilder.WriteString(fmt.Sprintf("-- Generated: %s\n", script.GeneratedAt.Format(time.RFC3339)))
	downBuilder.WriteString("-- DOWN Migration\n\n")

	// Generate in dependency order
	g.generateEnums(&upBuilder, &downBuilder, diff, script)
	g.generateSequences(&upBuilder, &downBuilder, diff)
	g.generateTables(&upBuilder, &downBuilder, diff, script)
	g.generateConstraints(&upBuilder, &downBuilder, diff)
	g.generateIndexes(&upBuilder, &downBuilder, diff)
	g.generateFunctions(&upBuilder, &downBuilder, diff)
	g.generateViews(&upBuilder, &downBuilder, diff)
	g.generateTriggers(&upBuilder, &downBuilder, diff)

	script.UpScript = upBuilder.String()
	script.DownScript = downBuilder.String()

	return script
}

// generateEnums generates enum-related SQL
func (g *Generator) generateEnums(up, down *strings.Builder, diff *DiffResult, script *GeneratedScript) {
	// Create enums
	for _, enum := range diff.CreatedEnums {
		up.WriteString(fmt.Sprintf("CREATE TYPE %s.%s AS ENUM (\n", g.schema, enum.Name))
		for i, val := range enum.Values {
			up.WriteString(fmt.Sprintf("    '%s'", val))
			if i < len(enum.Values)-1 {
				up.WriteString(",\n")
			} else {
				up.WriteString("\n")
			}
		}
		up.WriteString(");\n\n")

		// Down
		down.WriteString(fmt.Sprintf("DROP TYPE IF EXISTS %s.%s;\n\n", g.schema, enum.Name))
	}

	// Alter enums
	for _, enumDiff := range diff.AlteredEnums {
		for _, val := range enumDiff.AddedValues {
			up.WriteString(fmt.Sprintf("ALTER TYPE %s.%s ADD VALUE '%s';\n", g.schema, enumDiff.Name, val))
		}
		
		if len(enumDiff.RemovedValues) > 0 {
			script.Warnings = append(script.Warnings, 
				fmt.Sprintf("Cannot automatically remove enum values from %s. Manual intervention required.", enumDiff.Name))
			script.HasBreaking = true
		}
		up.WriteString("\n")
	}

	// Drop enums
	for _, enum := range diff.DroppedEnums {
		up.WriteString(fmt.Sprintf("DROP TYPE IF EXISTS %s.%s;\n\n", g.schema, enum.Name))
		
		// Down - recreate
		down.WriteString(fmt.Sprintf("CREATE TYPE %s.%s AS ENUM (\n", g.schema, enum.Name))
		for i, val := range enum.Values {
			down.WriteString(fmt.Sprintf("    '%s'", val))
			if i < len(enum.Values)-1 {
				down.WriteString(",\n")
			} else {
				down.WriteString("\n")
			}
		}
		down.WriteString(");\n\n")
	}
}

// generateSequences generates sequence-related SQL
func (g *Generator) generateSequences(up, down *strings.Builder, diff *DiffResult) {
	for _, seq := range diff.CreatedSequences {
		up.WriteString(fmt.Sprintf(
			"CREATE SEQUENCE %s.%s START WITH %d INCREMENT BY %d;\n\n",
			g.schema, seq.Name, seq.StartValue, seq.Increment,
		))
		down.WriteString(fmt.Sprintf("DROP SEQUENCE IF EXISTS %s.%s;\n\n", g.schema, seq.Name))
	}

	for _, seq := range diff.DroppedSequences {
		up.WriteString(fmt.Sprintf("DROP SEQUENCE IF EXISTS %s.%s;\n\n", g.schema, seq.Name))
		down.WriteString(fmt.Sprintf(
			"CREATE SEQUENCE %s.%s START WITH %d INCREMENT BY %d;\n\n",
			g.schema, seq.Name, seq.StartValue, seq.Increment,
		))
	}
}

// generateTables generates table-related SQL
func (g *Generator) generateTables(up, down *strings.Builder, diff *DiffResult, script *GeneratedScript) {
	// Create tables
	for _, table := range diff.CreatedTables {
		up.WriteString(fmt.Sprintf("CREATE TABLE %s.%s (\n", g.schema, table.Name))
		
		columns := make([]*Column, 0, len(table.Columns))
		for _, col := range table.Columns {
			columns = append(columns, col)
		}
		
		// Sort by ordinal position
		for i := 0; i < len(columns); i++ {
			for j := i + 1; j < len(columns); j++ {
				if columns[i].OrdinalPosition > columns[j].OrdinalPosition {
					columns[i], columns[j] = columns[j], columns[i]
				}
			}
		}
		
		for i, col := range columns {
			up.WriteString(fmt.Sprintf("    %s %s", col.Name, g.formatColumnType(col)))
			
			if !col.IsNullable {
				up.WriteString(" NOT NULL")
			}
			
			if col.DefaultValue != nil {
				up.WriteString(fmt.Sprintf(" DEFAULT %s", *col.DefaultValue))
			}
			
			if i < len(columns)-1 {
				up.WriteString(",\n")
			} else {
				up.WriteString("\n")
			}
		}
		
		up.WriteString(");\n\n")
		
		// Down
		down.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s CASCADE;\n\n", g.schema, table.Name))
	}

	// Alter tables
	for _, tableDiff := range diff.AlteredTables {
		// Add columns
		for _, col := range tableDiff.AddedColumns {
			up.WriteString(fmt.Sprintf(
				"ALTER TABLE %s.%s ADD COLUMN %s %s",
				g.schema, tableDiff.TableName, col.Name, g.formatColumnType(col),
			))
			
			if !col.IsNullable {
				up.WriteString(" NOT NULL")
			}
			
			if col.DefaultValue != nil {
				up.WriteString(fmt.Sprintf(" DEFAULT %s", *col.DefaultValue))
			}
			
			up.WriteString(";\n")
			
			// Down
			down.WriteString(fmt.Sprintf(
				"ALTER TABLE %s.%s DROP COLUMN IF EXISTS %s;\n",
				g.schema, tableDiff.TableName, col.Name,
			))
		}
		
		// Drop columns
		for _, col := range tableDiff.DroppedColumns {
			up.WriteString(fmt.Sprintf(
				"ALTER TABLE %s.%s DROP COLUMN IF EXISTS %s;\n",
				g.schema, tableDiff.TableName, col.Name,
			))
			script.Warnings = append(script.Warnings, 
				fmt.Sprintf("Dropping column %s.%s will result in data loss", tableDiff.TableName, col.Name))
			script.HasBreaking = true
			
			// Down
			down.WriteString(fmt.Sprintf(
				"ALTER TABLE %s.%s ADD COLUMN %s %s",
				g.schema, tableDiff.TableName, col.Name, g.formatColumnType(col),
			))
			
			if !col.IsNullable {
				down.WriteString(" NOT NULL")
			}
			
			if col.DefaultValue != nil {
				down.WriteString(fmt.Sprintf(" DEFAULT %s", *col.DefaultValue))
			}
			
			down.WriteString(";\n")
		}
		
		// Alter columns
		for _, colDiff := range tableDiff.AlteredColumns {
			if colDiff.TypeChanged {
				up.WriteString(fmt.Sprintf(
					"ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s;\n",
					g.schema, tableDiff.TableName, colDiff.Name, g.formatColumnType(colDiff.NewColumn),
				))
				down.WriteString(fmt.Sprintf(
					"ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s;\n",
					g.schema, tableDiff.TableName, colDiff.Name, g.formatColumnType(colDiff.OldColumn),
				))
				script.Warnings = append(script.Warnings, 
					fmt.Sprintf("Type change on %s.%s may cause data loss or conversion errors", tableDiff.TableName, colDiff.Name))
			}
			
			if colDiff.NullableChanged {
				if colDiff.NewColumn.IsNullable {
					up.WriteString(fmt.Sprintf(
						"ALTER TABLE %s.%s ALTER COLUMN %s DROP NOT NULL;\n",
						g.schema, tableDiff.TableName, colDiff.Name,
					))
					down.WriteString(fmt.Sprintf(
						"ALTER TABLE %s.%s ALTER COLUMN %s SET NOT NULL;\n",
						g.schema, tableDiff.TableName, colDiff.Name,
					))
				} else {
					up.WriteString(fmt.Sprintf(
						"ALTER TABLE %s.%s ALTER COLUMN %s SET NOT NULL;\n",
						g.schema, tableDiff.TableName, colDiff.Name,
					))
					down.WriteString(fmt.Sprintf(
						"ALTER TABLE %s.%s ALTER COLUMN %s DROP NOT NULL;\n",
						g.schema, tableDiff.TableName, colDiff.Name,
					))
					script.Warnings = append(script.Warnings, 
						fmt.Sprintf("Setting NOT NULL on %s.%s may fail if existing NULL values exist", tableDiff.TableName, colDiff.Name))
				}
			}
			
			if colDiff.DefaultChanged {
				if colDiff.NewColumn.DefaultValue != nil {
					up.WriteString(fmt.Sprintf(
						"ALTER TABLE %s.%s ALTER COLUMN %s SET DEFAULT %s;\n",
						g.schema, tableDiff.TableName, colDiff.Name, *colDiff.NewColumn.DefaultValue,
					))
				} else {
					up.WriteString(fmt.Sprintf(
						"ALTER TABLE %s.%s ALTER COLUMN %s DROP DEFAULT;\n",
						g.schema, tableDiff.TableName, colDiff.Name,
					))
				}
				
				if colDiff.OldColumn.DefaultValue != nil {
					down.WriteString(fmt.Sprintf(
						"ALTER TABLE %s.%s ALTER COLUMN %s SET DEFAULT %s;\n",
						g.schema, tableDiff.TableName, colDiff.Name, *colDiff.OldColumn.DefaultValue,
					))
				} else {
					down.WriteString(fmt.Sprintf(
						"ALTER TABLE %s.%s ALTER COLUMN %s DROP DEFAULT;\n",
						g.schema, tableDiff.TableName, colDiff.Name,
					))
				}
			}
		}
		
		up.WriteString("\n")
		down.WriteString("\n")
	}

	// Drop tables
	for _, table := range diff.DroppedTables {
		up.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s CASCADE;\n\n", g.schema, table.Name))
		script.Warnings = append(script.Warnings, 
			fmt.Sprintf("Dropping table %s will result in complete data loss", table.Name))
		script.HasBreaking = true
		
		// Down - recreate (simplified, real implementation should preserve all details)
		down.WriteString(fmt.Sprintf("-- TODO: Recreate table %s.%s with all columns and data\n\n", g.schema, table.Name))
	}
}

// generateConstraints generates constraint-related SQL
func (g *Generator) generateConstraints(up, down *strings.Builder, diff *DiffResult) {
	// Drop constraints first (for dependencies)
	for _, constraint := range diff.DroppedConstraints {
		up.WriteString(fmt.Sprintf(
			"ALTER TABLE %s.%s DROP CONSTRAINT IF EXISTS %s;\n",
			g.schema, constraint.TableName, constraint.Name,
		))
	}
	if len(diff.DroppedConstraints) > 0 {
		up.WriteString("\n")
	}

	// Create constraints
	for _, constraint := range diff.CreatedConstraints {
		up.WriteString(fmt.Sprintf(
			"ALTER TABLE %s.%s ADD CONSTRAINT %s %s;\n",
			g.schema, constraint.TableName, constraint.Name, constraint.Definition,
		))
		
		// Down
		down.WriteString(fmt.Sprintf(
			"ALTER TABLE %s.%s DROP CONSTRAINT IF EXISTS %s;\n",
			g.schema, constraint.TableName, constraint.Name,
		))
	}
	if len(diff.CreatedConstraints) > 0 {
		up.WriteString("\n")
		down.WriteString("\n")
	}

	// Recreate dropped constraints in down script
	for _, constraint := range diff.DroppedConstraints {
		down.WriteString(fmt.Sprintf(
			"ALTER TABLE %s.%s ADD CONSTRAINT %s %s;\n",
			g.schema, constraint.TableName, constraint.Name, constraint.Definition,
		))
	}
	if len(diff.DroppedConstraints) > 0 {
		down.WriteString("\n")
	}
}

// generateIndexes generates index-related SQL
func (g *Generator) generateIndexes(up, down *strings.Builder, diff *DiffResult) {
	for _, index := range diff.CreatedIndexes {
		up.WriteString(fmt.Sprintf("%s;\n", index.Definition))
		down.WriteString(fmt.Sprintf("DROP INDEX IF EXISTS %s.%s;\n", g.schema, index.Name))
	}
	if len(diff.CreatedIndexes) > 0 {
		up.WriteString("\n")
		down.WriteString("\n")
	}

	for _, index := range diff.DroppedIndexes {
		up.WriteString(fmt.Sprintf("DROP INDEX IF EXISTS %s.%s;\n", g.schema, index.Name))
		down.WriteString(fmt.Sprintf("%s;\n", index.Definition))
	}
	if len(diff.DroppedIndexes) > 0 {
		up.WriteString("\n")
		down.WriteString("\n")
	}
}

// generateFunctions generates function-related SQL
func (g *Generator) generateFunctions(up, down *strings.Builder, diff *DiffResult) {
	for _, fn := range diff.CreatedFunctions {
		up.WriteString(fn.Definition)
		up.WriteString(";\n\n")
		down.WriteString(fmt.Sprintf("DROP FUNCTION IF EXISTS %s.%s;\n\n", g.schema, fn.Name))
	}

	for _, fnDiff := range diff.AlteredFunctions {
		up.WriteString(fmt.Sprintf("-- Replacing function %s\n", fnDiff.Name))
		up.WriteString(fnDiff.NewDefinition)
		up.WriteString(";\n\n")
		
		down.WriteString(fmt.Sprintf("-- Restoring function %s\n", fnDiff.Name))
		down.WriteString(fnDiff.OldDefinition)
		down.WriteString(";\n\n")
	}

	for _, fn := range diff.DroppedFunctions {
		up.WriteString(fmt.Sprintf("DROP FUNCTION IF EXISTS %s.%s;\n\n", g.schema, fn.Name))
		down.WriteString(fn.Definition)
		down.WriteString(";\n\n")
	}
}

// generateViews generates view-related SQL
func (g *Generator) generateViews(up, down *strings.Builder, diff *DiffResult) {
	for _, view := range diff.CreatedViews {
		up.WriteString(fmt.Sprintf("CREATE VIEW %s.%s AS\n%s;\n\n", g.schema, view.Name, view.Definition))
		down.WriteString(fmt.Sprintf("DROP VIEW IF EXISTS %s.%s;\n\n", g.schema, view.Name))
	}

	for _, viewDiff := range diff.AlteredViews {
		up.WriteString(fmt.Sprintf("CREATE OR REPLACE VIEW %s.%s AS\n%s;\n\n", g.schema, viewDiff.Name, viewDiff.NewDefinition))
		down.WriteString(fmt.Sprintf("CREATE OR REPLACE VIEW %s.%s AS\n%s;\n\n", g.schema, viewDiff.Name, viewDiff.OldDefinition))
	}

	for _, view := range diff.DroppedViews {
		up.WriteString(fmt.Sprintf("DROP VIEW IF EXISTS %s.%s;\n\n", g.schema, view.Name))
		down.WriteString(fmt.Sprintf("CREATE VIEW %s.%s AS\n%s;\n\n", g.schema, view.Name, view.Definition))
	}
}

// generateTriggers generates trigger-related SQL
func (g *Generator) generateTriggers(up, down *strings.Builder, diff *DiffResult) {
	for _, trigger := range diff.CreatedTriggers {
		up.WriteString(fmt.Sprintf("%s;\n\n", trigger.Definition))
		down.WriteString(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s.%s;\n\n", 
			trigger.Name, g.schema, trigger.TableName))
	}

	for _, trigger := range diff.DroppedTriggers {
		up.WriteString(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s.%s;\n\n", 
			trigger.Name, g.schema, trigger.TableName))
		down.WriteString(fmt.Sprintf("%s;\n\n", trigger.Definition))
	}
}

// formatColumnType formats a column's data type
func (g *Generator) formatColumnType(col *Column) string {
	dataType := col.DataType
	
	if col.CharMaxLength != nil {
		return fmt.Sprintf("%s(%d)", dataType, *col.CharMaxLength)
	}
	
	if col.NumericPrecision != nil && col.NumericScale != nil {
		return fmt.Sprintf("%s(%d,%d)", dataType, *col.NumericPrecision, *col.NumericScale)
	}
	
	if col.NumericPrecision != nil {
		return fmt.Sprintf("%s(%d)", dataType, *col.NumericPrecision)
	}
	
	return dataType
}
```

### 6. Main Differ Orchestrator (`pkg/diff/differ.go`)

```go
package diff

import (
	"database/sql"
	"fmt"
)

// Differ orchestrates the diff process
type Differ struct {
	sourceDB *sql.DB
	targetDB *sql.DB
	schema   string
}

// NewDiffer creates a new differ
func NewDiffer(sourceDB, targetDB *sql.DB, schema string) *Differ {
	return &Differ{
		sourceDB: sourceDB,
		targetDB: targetDB,
		schema:   schema,
	}
}

// Diff performs the complete diff operation
func (d *Differ) Diff(description string) (*GeneratedScript, error) {
	// Read source schema
	sourceReader := NewSchemaReader(d.sourceDB, d.schema)
	sourceSchema, err := sourceReader.ReadSchema()
	if err != nil {
		return nil, fmt.Errorf("reading source schema: %w", err)
	}

	// Read target schema
	targetReader := NewSchemaReader(d.targetDB, d.schema)
	targetSchema, err := targetReader.ReadSchema()
	if err != nil {
		return nil, fmt.Errorf("reading target schema: %w", err)
	}

	// Compare schemas
	comparator := NewComparator()
	diffResult := comparator.Compare(sourceSchema, targetSchema)

	// Generate scripts
	generator := NewGenerator(d.schema)
	script := generator.Generate(diffResult, description)

	return script, nil
}
```

### 7. CLI Integration (`cmd/dbman/main.go`)

```go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"yourproject/pkg/diff"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "diff":
		runDiff()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func runDiff() {
	diffCmd := flag.NewFlagSet("diff", flag.ExitOnError)
	
	sourceURL := diffCmd.String("source", "", "Source database URL (postgres