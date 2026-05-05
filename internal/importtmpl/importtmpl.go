package importtmpl

type ColumnType string

const (
	ColumnTypeString  ColumnType = "string"
	ColumnTypeInt     ColumnType = "int"
	ColumnTypeDate    ColumnType = "date"
	ColumnTypeEnum    ColumnType = "enum"
	ColumnTypeFK      ColumnType = "fk"
)

type Column struct {
	Name      string
	Type      ColumnType
	Required  bool
	EnumValues []string
	FKEntity  string
}

type Template struct {
	Code    string
	Label   string
	Entity  string
	Columns []Column
}

type RowError struct {
	Row     int
	Column  string
	Message string
}
