package repository

// TagAttachResult reports which tags were newly attached vs already present when attaching a
// batch of tags to a command or variable.
type TagAttachResult struct {
	Added    []string
	Existing []string
}

// TagDetachResult reports which tags were actually removed vs weren't attached in the first
// place when detaching a batch of tags from a command or variable.
type TagDetachResult struct {
	Removed     []string
	NotAttached []string
}
