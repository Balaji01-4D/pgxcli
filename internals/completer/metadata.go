package completer

import (
	"sync"
	"time"
)

type RelationKind string

const (
	RelationKindTable RelationKind = "table"
	RelationKindView  RelationKind = "view"
)

// Reference: pgcli/pgcompleter.py

type MetaData struct {
	mu sync.RWMutex

	Databases []string

	SearchPath []string

	Tables map[string]map[string]*TableMetadata // schema -> table -> metadata

	Views map[string]map[string]*TableMetadata // schema -> view -> metadata

	Functions map[string]map[string][]*FunctionMetadata // schema -> functions(overload) -> metadata

	// we are not storing full enum metadata, just the names, suffices for completion
	// reference : pgcli/pgcompleter.py line 288 - 297
	DataTypes map[string]map[string]bool // schema -> datatype -> exists

	// keyword tree: a map mapping keywords to well known following keywords
	// ex: create -> map[table, user, database ...]
	KeyWordsTree map[string]any

	KeyWords []string

	BuiltinFunctions []string

	AllCompletions map[string]bool

	Casing map[string]string

	ReservedWords map[string]bool

	LastRefreshed time.Time
}

func NewMetaData() *MetaData {
	return &MetaData{
		Databases:        make([]string, 0),
		SearchPath:       make([]string, 0),
		Tables:           make(map[string]map[string]*TableMetadata),
		Views:            make(map[string]map[string]*TableMetadata),
		Functions:        make(map[string]map[string][]*FunctionMetadata),
		DataTypes:        make(map[string]map[string]bool),
		KeyWordsTree:     make(map[string]any),
		KeyWords:         make([]string, 0),
		BuiltinFunctions: make([]string, 0),
		AllCompletions:   make(map[string]bool),
		Casing:           make(map[string]string),
		ReservedWords:    make(map[string]bool),
		LastRefreshed:    time.Now(),
	}
}

// Reference: pgcli/completion_refresher.py (lines 1-80)

// MetaDataRefresher handles asynchronous refreshing of metadata
type MetaDataRefresher struct {
	mu sync.Mutex

	// channel to signal refresh requests
	refreshChan chan RefreshRequest

	// channel to signal shutdown
	stopChan chan struct{}

	// indicates if a refresh is currently ongoing
	isRefreshing bool

	// metadata to update
	meta *MetaData

	// database executor to use for fetching metadata
	executor DatabaseExecutor
}

// RefreshRequest represents a request to refresh certain types of metadata
type RefreshRequest struct {
	// Type of metadata to refresh
	// options: "schemata", "tables", "views", "functions", "datatypes", "etc"
	RefreshTypes []string

	// Callback to invoke after refresh is done
	Callback func(*MetaData)
}

// DatabaseExecutor defines the interface for executing database queries to fetch metadata
type DatabaseExecutor interface {

	// query schema names
	Schemas() ([]string, error)

	// Query table names along with their schema
	Tables() ([]Relation, error)

	// Query view names along with their schema
	Views() ([]Relation, error)

	// Query table column metadata
	TableColumns() ([]ColumnInfo, error)

	// Query view column metadata
	ViewColumns() ([]ColumnInfo, error)

	// Query function metadata
	Functions() ([]*FunctionMetadata, error)

	// Query data type names
	DataTypes() ([]DatatypeName, error)

	// Query foreign key relationships
	ForeignKeys() ([]ForeignKey, error)

	// Query database names
	Databases() ([]string, error)

	// Query the current search path
	SearchPath() ([]string, error)
}

// Python Reference: pgcli/packages/parseutils/meta.py

type ColumnMetadata struct {
	Name       string
	DataType   string
	ForeignKey []ForeignKey
	Default    *string
	HasDefault bool
}

// example school.id -> students.school_id
// parenttable = school
// parentcolumn = id
// childtable = students
// childcolumn = school_id
type ForeignKey struct {
	ParentSchema string
	ParentTable  string
	ParentColumn string
	ChildSchema  string
	ChildTable   string
	ChildColumn  string
}

type TableMetadata struct {
	Name    string
	Columns map[string]*ColumnMetadata
}

type FunctionMetadata struct {
	SchemaName     string
	FuncName       string
	ArgNames       []string
	ArgTypes       []string
	ArgModes       []string
	ArgDefaults    []string
	ReturnType     string
	IsAggregate    bool
	IsWindow       bool
	IsSetReturning bool
	IsExtension    bool
	IsPublic       bool
}

type Relation struct {
	Schema string
	Name   string
	Kind   RelationKind
}

type ColumnInfo struct {
	Schema     string
	Table      string
	Column     string
	DataType   string
	HasDefault bool
	Default    *string
}

type DatatypeName struct {
	Schema string
	Name   string
}
