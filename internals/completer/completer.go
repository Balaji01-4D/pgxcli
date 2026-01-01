package completer

import "strings"

type Completer struct {
	metadata *MetaData

	executor DatabaseExecutor

	smartCompletion bool
}

func New() *Completer {
	return &Completer{
		metadata: NewMetaData(),
	}
}

func (c *Completer) ExtendDatabases(databases []string) {
	c.metadata.mu.Lock()
	defer c.metadata.mu.Unlock()

	for _, db := range databases {
		c.metadata.Databases = append(c.metadata.Databases, db)
		c.metadata.AllCompletions[db] = true
	}
}

func (c *Completer) ExtendSchemas(schemas []string) {
	c.metadata.mu.Lock()
	defer c.metadata.mu.Unlock()

	for _, schema := range schemas {
		escapedSchema := c.escapeName(schema)

		if c.metadata.Tables[escapedSchema] == nil {
			c.metadata.Tables[escapedSchema] = make(map[string]*TableMetadata)
		}
		if c.metadata.Views[escapedSchema] == nil {
			c.metadata.Views[escapedSchema] = make(map[string]*TableMetadata)
		}
		if c.metadata.Functions[escapedSchema] == nil {
			c.metadata.Functions[escapedSchema] = make(map[string][]*FunctionMetadata)
		}

		if c.metadata.DataTypes[escapedSchema] == nil {
			c.metadata.DataTypes[escapedSchema] = make(map[string]bool)
		}

		c.metadata.AllCompletions[escapedSchema] = true
	}
}

func (c *Completer) ExtendTables(tables []Relation) {
	c.metadata.mu.Lock()
	defer c.metadata.mu.Unlock()

	for _, table := range tables {
		escapedSchemaName := c.escapeName(table.Schema)
		escpapedTableName := c.escapeName(table.Name)

		if c.metadata.Tables[escapedSchemaName] == nil {
			c.metadata.Tables[escapedSchemaName] = make(map[string]*TableMetadata)
		}

		c.metadata.Tables[escapedSchemaName][escpapedTableName] = &TableMetadata{
			Name:    table.Name,
			Columns: make(map[string]*ColumnMetadata),
		}
		c.metadata.AllCompletions[escpapedTableName] = true
	}
}

func (c *Completer) ExtendColumns(columns []ColumnInfo, isView bool) {
	c.metadata.mu.Lock()
	defer c.metadata.mu.Unlock()

	var targetMap map[string]map[string]*TableMetadata
	if isView {
		targetMap = c.metadata.Views
	} else {
		targetMap = c.metadata.Tables
	}

	for _, column := range columns {
		escapedSchemaName := c.escapeName(column.Schema)
		escapedTableName := c.escapeName(column.Table)
		escapedColumnName := c.escapeName(column.Column)

		// get or create schema
		if targetMap[escapedSchemaName] == nil {
			targetMap[escapedSchemaName] = make(map[string]*TableMetadata)
		}

		// get or create table
		if targetMap[escapedSchemaName][escapedTableName] == nil {
			targetMap[escapedSchemaName][escapedTableName] = &TableMetadata{
				Name:    column.Table,
				Columns: make(map[string]*ColumnMetadata),
			}
		}

		// add column
		targetMap[escapedSchemaName][escapedTableName].Columns[escapedColumnName] = &ColumnMetadata{
			Name:       column.Column,
			DataType:   column.DataType,
			ForeignKey: []ForeignKey{},
			HasDefault: column.HasDefault,
			Default:    column.Default,
		}
		c.metadata.AllCompletions[escapedColumnName] = true
	}
}

func (c *Completer) ExtendForeignKeys(foreignKeys []ForeignKey) {
	c.metadata.mu.Lock()
	defer c.metadata.mu.Unlock()

	for _, fk := range foreignKeys {
		escapedChildSchema := c.escapeName(fk.ChildSchema)
		escapedChildTable := c.escapeName(fk.ChildTable)
		escapedChildColumn := c.escapeName(fk.ChildColumn)

		// check if table exists
		if tableMeta, ok := c.metadata.Tables[escapedChildSchema][escapedChildTable]; ok {
			// check if column exists
			if columnMeta, ok := tableMeta.Columns[escapedChildColumn]; ok {
				columnMeta.ForeignKey = append(columnMeta.ForeignKey, fk)
			}
		}
	}
}

func (c *Completer) ExtendFunctions(funcs []*FunctionMetadata) {
	c.metadata.mu.Lock()
	defer c.metadata.mu.Unlock()

	for _, fn := range funcs {
		escShema := c.escapeName(fn.SchemaName)
		escfunc := c.escapeName(fn.FuncName)

		// set isPublic flag
		fn.IsPublic = (escShema == "public")

		// get or create schema
		if c.metadata.Functions[escShema] == nil {
			c.metadata.Functions[escShema] = make(map[string][]*FunctionMetadata)
		}

		c.metadata.Functions[escShema][escfunc] = append(c.metadata.Functions[escShema][escfunc], fn)

		c.metadata.AllCompletions[escfunc] = true
	}
}

func (c *Completer) ExtendDataTypes(dataTypes []DatatypeName) {
	c.metadata.mu.Lock()
	defer c.metadata.mu.Unlock()

	for _, dt := range dataTypes {
		escapedSchemaName := c.escapeName(dt.Schema)
		escapedTypeName := c.escapeName(dt.Name)

		// get or create schema
		if c.metadata.DataTypes[escapedSchemaName] == nil {
			c.metadata.DataTypes[escapedSchemaName] = make(map[string]bool)
		}

		c.metadata.DataTypes[escapedSchemaName][escapedTypeName] = true
		c.metadata.AllCompletions[escapedTypeName] = true
	}
}

func (c *Completer) unescapeName(name string) string {
	if len(name) >= 2 && name[0] == '"' && name[len(name)-1] == '"' {
		return name[1 : len(name)-1]
	}
	return name
}

func (c *Completer) escapeName(name string) string {
	if name == "" {
		return name
	}

	needsQuoting := false

	if c.metadata.ReservedWords[strings.ToUpper(name)] {
		needsQuoting = true
	}

	if needsQuoting && !strings.HasPrefix(name, `"`) {
		return `"` + name + `"`
	}

	return name
}
