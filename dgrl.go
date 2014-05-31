// Package dgrl implements parsing and generation of the Doggerel language.
package dgrl

// Node types
const (
	NoType = iota
	TextType
	LeafType
	LongLeafType
	CommentType
	BranchType
)

// Parse contexts
const (
	DefaultContext = iota
	LongLeafContext
	CommentContext
)
