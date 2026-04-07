package singlechar

// Example tests single-letter field names that could conflict with closure receivers.
type Example struct {
	Name string `opt:"-"`
	E    int
	X    string
	A    bool
}
