package domain

// Binding links a template to a specific data source query.
type Binding struct {
	ID       IDType
	Name     NameType
	Query    SQLQueryType
	Template IDType
}
