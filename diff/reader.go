package diff

import (
	"database/sql"
	"fmt"
	"strings"
)

// SchemaReader reads database schema information
// Add these fields to SchemaReader
type SchemaReader struct {
	db            *sql.DB
	schema        string
	ignoreTables  map[string]bool
	ignoreColumns map[string]bool // key format: "table.column"
}

// SetIgnoreTables sets tables to ignore during schema reading
func (r *SchemaReader) SetIgnoreTables(tables []string) {
	r.ignoreTables = make(map[string]bool)
	for _, table := range tables {
		r.ignoreTables[table] = true
	}
}

// SetIgnoreColumns sets columns to ignore during schema reading
func (r *SchemaReader) SetIgnoreColumns(columns []string) {
	r.ignoreColumns = make(map[string]bool)
	for _, col := range columns {
		r.ignoreColumns[col] = true
	}
}

// shouldIgnoreTable checks if a table should be ignored
func (r *SchemaReader) shouldIgnoreTable(tableName string) bool {
	return r.ignoreTables[tableName]
}

// shouldIgnoreColumn checks if a column should be ignored
func (r *SchemaReader) shouldIgnoreColumn(tableName, columnName string) bool {
	return r.ignoreColumns[fmt.Sprintf("%s.%s", tableName, columnName)]
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
			name                                  string
			dataType                              string
			isNullable                            string
			defaultValue                          sql.NullString
			charMaxLength, numPrecision, numScale sql.NullInt64
			ordinalPosition                       int
			comment                               sql.NullString
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
