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
