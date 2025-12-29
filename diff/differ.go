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
