package domain

// Template represents an HTML document template that can contain Handlebars-style variables.
type Template struct {
	ID         IDType
	Name       NameType
	Code       CodeType
	Version    VersionType
	Stylesheet *IDType
	Content    HTMLType
	Subject    SubjectType
}
