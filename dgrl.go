package dgrl

const (
	NoType = iota
	TextType
	LeafType
	LongLeafType
	CommentType
	BranchType
)

const (
	DefaultContext = iota
	LongLeafContext
	CommentContext
)
