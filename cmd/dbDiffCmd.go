/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"southwinds.dev/dbman/diff"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

type DbDiffCmd struct {
	cmd *cobra.Command

	// Connection flags
	SourceURL string
	TargetURL string

	// Or individual connection parameters
	SourceHost     string
	SourcePort     int
	SourceDB       string
	SourceUser     string
	SourcePassword string

	TargetHost     string
	TargetPort     int
	TargetDB       string
	TargetUser     string
	TargetPassword string

	// Schema options
	Schema string

	// Output options
	OutputDir   string
	Description string
	Format      string // "split" or "single"

	// Behavior options
	IgnoreTables  []string
	IgnoreColumns []string
	IncludeData   bool
	DryRun        bool
	Verbose       bool

	// Safety options
	RequireConfirm bool
	SkipWarnings   bool
}

func NewDbDiffCmd() *DbDiffCmd {
	c := &DbDiffCmd{
		cmd: &cobra.Command{
			Use:   "diff",
			Short: "Generate migration scripts by comparing two database schemas",
			Long: `Compare two PostgreSQL databases and generate UP/DOWN migration scripts.
        
Examples:
  # Compare using full connection URLs
  dbman diff --source "postgres://user:pass@localhost/db1" \
             --target "postgres://user:pass@localhost/db2"
  
  # Compare using individual parameters
  dbman diff --source-host localhost --source-db old_version \
             --target-host localhost --target-db new_version
  
  # Compare specific schema with custom output
  dbman diff --source $DB1 --target $DB2 \
             --schema public \
             --output ./migrations \
             --description "Add user authentication"`,
			Example: `
# Basic usage with connection URLs
dbman diff \
  --source "postgres://user:pass@localhost:5432/old_db" \
  --target "postgres://user:pass@localhost:5432/new_db"

# With description and custom output
dbman diff \
  --source "postgres://user:pass@localhost:5432/v1" \
  --target "postgres://user:pass@localhost:5432/v2" \
  --description "Add user authentication" \
  --output ./migrations

# Using individual connection parameters
dbman diff \
  --source-host localhost \
  --source-db production_v1 \
  --source-user admin \
  --source-password secret \
  --target-host localhost \
  --target-db production_v2 \
  --target-user admin \
  --target-password secret \
  --schema public

# Dry run to preview changes
dbman diff \
  --source $SOURCE_DB \
  --target $TARGET_DB \
  --dry-run \
  --verbose

# Ignore specific tables and columns
dbman diff \
  --source $SOURCE_DB \
  --target $TARGET_DB \
  --ignore-tables "temp_table,log_table" \
  --ignore-columns "users.legacy_field,orders.old_status"

# Single file output format
dbman diff \
  --source $SOURCE_DB \
  --target $TARGET_DB \
  --format single \
  --output ./migrations

# With confirmation for breaking changes
dbman diff \
  --source $SOURCE_DB \
  --target $TARGET_DB \
  --confirm \
  --verbose
`,
		}}

	c.cmd.RunE = c.Run

	// Connection flags - URL style
	c.cmd.Flags().StringVar(&c.SourceURL, "source", "",
		"Source database URL (postgres://user:pass@host:port/dbname)")
	c.cmd.Flags().StringVar(&c.TargetURL, "target", "",
		"Target database URL (postgres://user:pass@host:port/dbname)")

	// Connection flags - Individual parameters
	c.cmd.Flags().StringVar(&c.SourceHost, "source-host", "localhost",
		"Source database host")
	c.cmd.Flags().IntVar(&c.SourcePort, "source-port", 5432,
		"Source database port")
	c.cmd.Flags().StringVar(&c.SourceDB, "source-db", "",
		"Source database name")
	c.cmd.Flags().StringVar(&c.SourceUser, "source-user", "postgres",
		"Source database user")
	c.cmd.Flags().StringVar(&c.SourcePassword, "source-password", "",
		"Source database password")

	c.cmd.Flags().StringVar(&c.TargetHost, "target-host", "localhost",
		"Target database host")
	c.cmd.Flags().IntVar(&c.TargetPort, "target-port", 5432,
		"Target database port")
	c.cmd.Flags().StringVar(&c.TargetDB, "target-db", "",
		"Target database name")
	c.cmd.Flags().StringVar(&c.TargetUser, "target-user", "postgres",
		"Target database user")
	c.cmd.Flags().StringVar(&c.TargetPassword, "target-password", "",
		"Target database password")

	// Schema options
	c.cmd.Flags().StringVar(&c.Schema, "schema", "public",
		"Database schema to compare")

	// Output options
	c.cmd.Flags().StringVarP(&c.OutputDir, "output", "o", "./migrations",
		"Output directory for migration files")
	c.cmd.Flags().StringVarP(&c.Description, "description", "d", "",
		"Migration description")
	c.cmd.Flags().StringVar(&c.Format, "format", "split",
		"Output format: 'split' (separate up/down files) or 'single'")

	// Behavior options
	c.cmd.Flags().StringSliceVar(&c.IgnoreTables, "ignore-tables", []string{},
		"Comma-separated list of tables to ignore")
	c.cmd.Flags().StringSliceVar(&c.IgnoreColumns, "ignore-columns", []string{},
		"Comma-separated list of columns to ignore (format: table.column)")
	c.cmd.Flags().BoolVar(&c.IncludeData, "include-data", false,
		"Include data migration scripts (experimental)")
	c.cmd.Flags().BoolVar(&c.DryRun, "dry-run", false,
		"Show what would be generated without writing files")
	c.cmd.Flags().BoolVarP(&c.Verbose, "verbose", "v", false,
		"Verbose output")

	// Safety options
	c.cmd.Flags().BoolVar(&c.RequireConfirm, "confirm", false,
		"Require confirmation before generating scripts with breaking changes")
	c.cmd.Flags().BoolVar(&c.SkipWarnings, "skip-warnings", false,
		"Skip warning prompts")

	return c
}

func (c *DbDiffCmd) Run(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Step 1: Validate flags
	if err := c.validateFlags(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Build connection strings
	sourceConnStr, err := c.buildSourceConnectionString()
	if err != nil {
		return fmt.Errorf("invalid source connection: %w", err)
	}

	targetConnStr, err := c.buildTargetConnectionString()
	if err != nil {
		return fmt.Errorf("invalid target connection: %w", err)
	}

	if c.Verbose {
		fmt.Printf("🔌 Connecting to source: %s\n", c.maskPassword(sourceConnStr))
		fmt.Printf("🔌 Connecting to target: %s\n", c.maskPassword(targetConnStr))
	}

	// Step 3: Connect to databases
	sourceDB, err := c.connectDatabase(sourceConnStr, "source")
	if err != nil {
		return err
	}
	defer sourceDB.Close()

	targetDB, err := c.connectDatabase(targetConnStr, "target")
	if err != nil {
		return err
	}
	defer targetDB.Close()

	// Step 4: Verify schema exists
	if err := c.verifySchema(sourceDB, "source"); err != nil {
		return err
	}
	if err := c.verifySchema(targetDB, "target"); err != nil {
		return err
	}

	fmt.Println("📊 Reading source database schema...")

	// Step 5: Read source schema
	sourceReader := diff.NewSchemaReader(sourceDB, c.Schema)
	sourceReader.SetIgnoreTables(c.IgnoreTables)
	sourceReader.SetIgnoreColumns(c.IgnoreColumns)

	sourceSchema, err := sourceReader.ReadSchema()
	if err != nil {
		return fmt.Errorf("failed to read source schema: %w", err)
	}

	if c.Verbose {
		fmt.Printf("  ✅ Found %d tables, %d indexes, %d constraints\n",
			len(sourceSchema.Tables),
			len(sourceSchema.Indexes),
			len(sourceSchema.Constraints))
	}

	fmt.Println("📊 Reading target database schema...")

	// Step 6: Read target schema
	targetReader := diff.NewSchemaReader(targetDB, c.Schema)
	targetReader.SetIgnoreTables(c.IgnoreTables)
	targetReader.SetIgnoreColumns(c.IgnoreColumns)

	targetSchema, err := targetReader.ReadSchema()
	if err != nil {
		return fmt.Errorf("failed to read target schema: %w", err)
	}

	if c.Verbose {
		fmt.Printf("  ✅ Found %d tables, %d indexes, %d constraints\n",
			len(targetSchema.Tables),
			len(targetSchema.Indexes),
			len(targetSchema.Constraints))
	}

	fmt.Println("🔍 Comparing schemas...")

	// Step 7: Compare schemas
	comparator := diff.NewComparator()
	diffResult := comparator.Compare(sourceSchema, targetSchema)

	// Step 8: Display summary
	c.printDiffSummary(diffResult)

	// Step 9: Check if there are any changes
	if !diffResult.HasChanges() {
		fmt.Println("✨ No differences found. Databases are identical.")
		return nil
	}

	// Step 10: Generate migration scripts
	fmt.Println("📝 Generating migration scripts...")

	generator := diff.NewGenerator(c.Schema)
	script := generator.Generate(diffResult, c.getDescription())

	// Step 11: Display warnings
	if len(script.Warnings) > 0 && !c.SkipWarnings {
		fmt.Println("\n⚠️  WARNINGS:")
		for _, warning := range script.Warnings {
			fmt.Printf("   %s\n", warning)
		}

		if script.HasBreaking && c.RequireConfirm && !c.DryRun {
			if !c.confirmProceed() {
				return fmt.Errorf("operation cancelled by user")
			}
		}
	}

	// Step 12: Dry run - just display
	if c.DryRun {
		fmt.Println("\n🔍 DRY RUN - Scripts would be generated:")
		c.printScriptPreview(script)
		return nil
	}

	// Step 13: Write migration files
	files, err := c.writeMigrationFiles(script)
	if err != nil {
		return fmt.Errorf("failed to write migration files: %w", err)
	}

	// Step 14: Success message
	duration := time.Since(startTime)
	fmt.Printf("\n✅ Migration scripts generated successfully in %v\n", duration)
	fmt.Println("\n📁 Generated files:")
	for _, file := range files {
		fmt.Printf("   %s\n", file)
	}

	// Step 15: Next steps
	c.printNextSteps(files)

	return nil
}

// validateFlags ensures all required flags are set
func (c *DbDiffCmd) validateFlags() error {
	// Check if we have connection info
	hasSourceURL := c.SourceURL != ""
	hasTargetURL := c.TargetURL != ""
	hasSourceParams := c.SourceDB != ""
	hasTargetParams := c.TargetDB != ""

	if !hasSourceURL && !hasSourceParams {
		return fmt.Errorf("either --source or --source-db must be provided")
	}

	if !hasTargetURL && !hasTargetParams {
		return fmt.Errorf("either --target or --target-db must be provided")
	}

	// Validate format
	if c.Format != "split" && c.Format != "single" {
		return fmt.Errorf("format must be 'split' or 'single', got: %s", c.Format)
	}

	return nil
}

// buildSourceConnectionString constructs the source database connection string
func (c *DbDiffCmd) buildSourceConnectionString() (string, error) {
	if c.SourceURL != "" {
		return c.SourceURL, nil
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.SourceHost,
		c.SourcePort,
		c.SourceUser,
		c.SourcePassword,
		c.SourceDB,
	), nil
}

// buildTargetConnectionString constructs the target database connection string
func (c *DbDiffCmd) buildTargetConnectionString() (string, error) {
	if c.TargetURL != "" {
		return c.TargetURL, nil
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.TargetHost,
		c.TargetPort,
		c.TargetUser,
		c.TargetPassword,
		c.TargetDB,
	), nil
}

// connectDatabase establishes a database connection
func (c *DbDiffCmd) connectDatabase(connStr, name string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %w", name, err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping %s database: %w", name, err)
	}

	if c.Verbose {
		fmt.Printf("  ✅ Connected to %s database\n", name)
	}

	return db, nil
}

// verifySchema checks if the schema exists
func (c *DbDiffCmd) verifySchema(db *sql.DB, dbName string) error {
	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1 FROM information_schema.schemata 
            WHERE schema_name = $1
        )
    `

	err := db.QueryRow(query, c.Schema).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify schema in %s: %w", dbName, err)
	}

	if !exists {
		return fmt.Errorf("schema '%s' does not exist in %s database", c.Schema, dbName)
	}

	return nil
}

// printDiffSummary displays a summary of the differences found
func (c *DbDiffCmd) printDiffSummary(result *diff.DiffResult) {
	fmt.Println("\n📋 Diff Summary:")

	// Tables
	if len(result.CreatedTables) > 0 {
		fmt.Printf("  ➕ Tables to create: %d\n", len(result.CreatedTables))
		if c.Verbose {
			for _, t := range result.CreatedTables {
				fmt.Printf("     - %s\n", t.Name)
			}
		}
	}

	if len(result.DroppedTables) > 0 {
		fmt.Printf("  ➖ Tables to drop: %d\n", len(result.DroppedTables))
		if c.Verbose {
			for _, t := range result.DroppedTables {
				fmt.Printf("     - %s\n", t.Name)
			}
		}
	}

	if len(result.AlteredTables) > 0 {
		fmt.Printf("  🔄 Tables to alter: %d\n", len(result.AlteredTables))
		if c.Verbose {
			for _, t := range result.AlteredTables {
				fmt.Printf("     - %s (%d columns added, %d dropped, %d altered)\n",
					t.TableName,
					len(t.AddedColumns),
					len(t.DroppedColumns),
					len(t.AlteredColumns))
			}
		}
	}

	// Indexes
	if len(result.CreatedIndexes) > 0 {
		fmt.Printf("  ➕ Indexes to create: %d\n", len(result.CreatedIndexes))
	}
	if len(result.DroppedIndexes) > 0 {
		fmt.Printf("  ➖ Indexes to drop: %d\n", len(result.DroppedIndexes))
	}

	// Constraints
	if len(result.CreatedConstraints) > 0 {
		fmt.Printf("  ➕ Constraints to create: %d\n", len(result.CreatedConstraints))
	}
	if len(result.DroppedConstraints) > 0 {
		fmt.Printf("  ➖ Constraints to drop: %d\n", len(result.DroppedConstraints))
	}

	// Views
	if len(result.CreatedViews) > 0 {
		fmt.Printf("  ➕ Views to create: %d\n", len(result.CreatedViews))
	}
	if len(result.DroppedViews) > 0 {
		fmt.Printf("  ➖ Views to drop: %d\n", len(result.DroppedViews))
	}

	// Functions
	if len(result.CreatedFunctions) > 0 {
		fmt.Printf("  ➕ Functions to create: %d\n", len(result.CreatedFunctions))
	}
	if len(result.DroppedFunctions) > 0 {
		fmt.Printf("  ➖ Functions to drop: %d\n", len(result.DroppedFunctions))
	}

	// Triggers
	if len(result.CreatedTriggers) > 0 {
		fmt.Printf("  ➕ Triggers to create: %d\n", len(result.CreatedTriggers))
	}
	if len(result.DroppedTriggers) > 0 {
		fmt.Printf("  ➖ Triggers to drop: %d\n", len(result.DroppedTriggers))
	}

	// Enums
	if len(result.CreatedEnums) > 0 {
		fmt.Printf("  ➕ Enums to create: %d\n", len(result.CreatedEnums))
	}
	if len(result.DroppedEnums) > 0 {
		fmt.Printf("  ➖ Enums to drop: %d\n", len(result.DroppedEnums))
	}

	// Sequences
	if len(result.CreatedSequences) > 0 {
		fmt.Printf("  ➕ Sequences to create: %d\n", len(result.CreatedSequences))
	}
	if len(result.DroppedSequences) > 0 {
		fmt.Printf("  ➖ Sequences to drop: %d\n", len(result.DroppedSequences))
	}
}

// printScriptPreview shows a preview of the generated scripts
func (c *DbDiffCmd) printScriptPreview(script *diff.GeneratedScript) {
	fmt.Println("\n--- UP SCRIPT PREVIEW (first 50 lines) ---")
	c.printLines(script.UpScript, 50)

	fmt.Println("\n--- DOWN SCRIPT PREVIEW (first 50 lines) ---")
	c.printLines(script.DownScript, 50)
}

// printLines prints the first n lines of a string
func (c *DbDiffCmd) printLines(content string, maxLines int) {
	lines := strings.Split(content, "\n")
	count := min(len(lines), maxLines)

	for i := 0; i < count; i++ {
		fmt.Println(lines[i])
	}

	if len(lines) > maxLines {
		fmt.Printf("... (%d more lines)\n", len(lines)-maxLines)
	}
}

// writeMigrationFiles writes the migration scripts to files
func (c *DbDiffCmd) writeMigrationFiles(script *diff.GeneratedScript) ([]string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate migration file name
	timestamp := time.Now().Format("20060102150405")
	baseName := c.generateMigrationName(timestamp)

	var files []string

	if c.Format == "split" {
		// Write separate UP and DOWN files
		upFile := filepath.Join(c.OutputDir, baseName+".up.sql")
		downFile := filepath.Join(c.OutputDir, baseName+".down.sql")

		if err := os.WriteFile(upFile, []byte(script.UpScript), 0644); err != nil {
			return nil, fmt.Errorf("failed to write UP script: %w", err)
		}
		files = append(files, upFile)

		if err := os.WriteFile(downFile, []byte(script.DownScript), 0644); err != nil {
			return nil, fmt.Errorf("failed to write DOWN script: %w", err)
		}
		files = append(files, downFile)
	} else {
		// Write single file with both UP and DOWN
		singleFile := filepath.Join(c.OutputDir, baseName+".sql")
		content := fmt.Sprintf("-- +migrate Up\n%s\n-- +migrate Down\n%s",
			script.UpScript, script.DownScript)

		if err := os.WriteFile(singleFile, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write migration script: %w", err)
		}
		files = append(files, singleFile)
	}

	// Write warnings file if there are any
	if len(script.Warnings) > 0 {
		warningsFile := filepath.Join(c.OutputDir, baseName+".warnings.txt")
		warningsContent := strings.Join(script.Warnings, "\n")

		if err := os.WriteFile(warningsFile, []byte(warningsContent), 0644); err != nil {
			return files, fmt.Errorf("failed to write warnings file: %w", err)
		}
		files = append(files, warningsFile)
	}

	return files, nil
}

// generateMigrationName creates a migration file name
func (c *DbDiffCmd) generateMigrationName(timestamp string) string {
	if c.Description != "" {
		// Convert description to snake_case
		desc := strings.ToLower(c.Description)
		desc = strings.ReplaceAll(desc, " ", "_")
		desc = strings.ReplaceAll(desc, "-", "_")
		return fmt.Sprintf("%s_%s", timestamp, desc)
	}
	return timestamp + "_schema_diff"
}

// getDescription returns the migration description
func (c *DbDiffCmd) getDescription() string {
	if c.Description != "" {
		return c.Description
	}
	return "Schema differences"
}

// confirmProceed asks the user for confirmation
func (c *DbDiffCmd) confirmProceed() bool {
	fmt.Print("\n⚠️  This migration contains breaking changes. Proceed? (yes/no): ")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "yes" || response == "y"
}

// maskPassword masks the password in connection string for display
func (c *DbDiffCmd) maskPassword(connStr string) string {
	// Simple password masking for display
	if strings.Contains(connStr, "password=") {
		parts := strings.Split(connStr, "password=")
		if len(parts) > 1 {
			afterPassword := strings.SplitN(parts[1], " ", 2)
			if len(afterPassword) > 1 {
				return parts[0] + "password=***** " + afterPassword[1]
			}
			return parts[0] + "password=*****"
		}
	}
	return connStr
}

// printNextSteps shows the user what to do next
func (c *DbDiffCmd) printNextSteps(files []string) {
	fmt.Println("\n📚 Next steps:")
	fmt.Println("   1. Review the generated migration scripts")

	if len(files) > 0 {
		fmt.Printf("   2. Apply the migration: psql -f %s\n", files[0])
	}

	fmt.Println("   3. Test the migration in a development environment first")
	fmt.Println("   4. Backup your database before applying to production")
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
