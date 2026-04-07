package allrequired

// AllRequired has all fields marked as required - no optional fields.
type AllRequired struct {
	Name string `opt:"-"`
	Age  int    `opt:"-"`
}
