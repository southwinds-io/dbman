package diff

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
