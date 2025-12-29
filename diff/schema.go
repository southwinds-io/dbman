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
	Name             string
	DataType         string
	IsNullable       bool
	DefaultValue     *string
	CharMaxLength    *int
	NumericPrecision *int
	NumericScale     *int
	OrdinalPosition  int
	Comment          *string
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
	Name       string
	Schema     string
	StartValue int64
	Increment  int64
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
	CreatedTables []*Table
	DroppedTables []*Table
	AlteredTables []*TableDiff

	CreatedIndexes []*Index
	DroppedIndexes []*Index

	CreatedConstraints []*Constraint
	DroppedConstraints []*Constraint

	CreatedSequences []*Sequence
	DroppedSequences []*Sequence

	CreatedEnums []*Enum
	DroppedEnums []*Enum
	AlteredEnums []*EnumDiff

	CreatedFunctions []*Function
	DroppedFunctions []*Function
	AlteredFunctions []*FunctionDiff

	CreatedViews []*View
	DroppedViews []*View
	AlteredViews []*ViewDiff

	CreatedTriggers []*Trigger
	DroppedTriggers []*Trigger
}

// HasChanges returns true if there are any schema differences
func (d *DiffResult) HasChanges() bool {
	return len(d.CreatedTables) > 0 ||
		len(d.DroppedTables) > 0 ||
		len(d.AlteredTables) > 0 ||
		len(d.CreatedIndexes) > 0 ||
		len(d.DroppedIndexes) > 0 ||
		len(d.CreatedConstraints) > 0 ||
		len(d.DroppedConstraints) > 0 ||
		len(d.CreatedViews) > 0 ||
		len(d.DroppedViews) > 0 ||
		len(d.AlteredViews) > 0 ||
		len(d.CreatedFunctions) > 0 ||
		len(d.DroppedFunctions) > 0 ||
		len(d.AlteredFunctions) > 0 ||
		len(d.CreatedTriggers) > 0 ||
		len(d.DroppedTriggers) > 0 ||
		len(d.CreatedEnums) > 0 ||
		len(d.DroppedEnums) > 0 ||
		len(d.CreatedSequences) > 0 ||
		len(d.DroppedSequences) > 0
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
	Name          string
	AddedValues   []string
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
	UpScript    string
	DownScript  string
	GeneratedAt time.Time
	Description string
	HasBreaking bool
	Warnings    []string
}
