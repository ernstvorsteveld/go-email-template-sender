package domain

// Context represents an arbitrary JSON payload and metadata used for data binding.
type Context struct {
	ID           IDType
	Reference    ReferenceType
	Customer     CustomerType
	Payload      JSONPayloadType
	EmailAddress JSONPathType
}
