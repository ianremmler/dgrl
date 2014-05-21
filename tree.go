package dgrl

import (
	"strings"
)

type Node interface {
	Parent() *Branch
	SetParent(*Branch)
	Type() int
	String() string
	ToJSON() string
}

type Branch struct {
	parent *Branch
	name   string
	kids   []Node
}

func NewRoot() *Branch {
	return NewBranch("")
}

func NewBranch(name string) *Branch {
	return &Branch{
		name: name,
		kids: []Node{},
	}
}

func (b *Branch) String() string {
	str := ""
	prevTyp := NoType
	lvl := b.Level()
	if lvl > 0 {
		str = strings.Repeat("=", lvl) + " " + b.name + "\n"
		prevTyp = CommentType // fake type to get a newline between branch and first kid
	}
	for _, kid := range b.kids {
		typ := kid.Type()
		if prevTyp == BranchType && typ != BranchType {
			str += "\n" + strings.Repeat("=", lvl+1) + "\n"
		}
		if prevTyp != NoType && !(typ == LeafType && prevTyp == LeafType) {
			str += "\n"
		}
		if typ == TextType && (prevTyp == TextType || prevTyp == LongLeafType) {
			str += "-\n\n"
		}
		str += kid.String()
		prevTyp = typ
	}
	return str
}

func (b *Branch) ToJSON() string {
	str := "{ \"" + b.name + "\": [ "
	for i, kid := range b.kids {
		str += kid.ToJSON()
		if i < len(b.kids)-1 {
			str += ","
		}
		str += " "
	}
	str += "] }"
	return str
}

func (b *Branch) Type() int {
	return BranchType
}

func (b *Branch) Level() int {
	i := 0
	for ; b.parent != nil; i++ {
		b = b.parent
	}
	return i
}

func (b *Branch) Name() string        { return b.name }
func (b *Branch) SetName(name string) { b.name = name }

func (b *Branch) Parent() *Branch          { return b.parent }
func (b *Branch) SetParent(parent *Branch) { b.parent = parent }

func (b *Branch) NumKids() int { return len(b.kids) }
func (b *Branch) Kids() []Node { return b.kids }

func (b *Branch) Append(node Node) {
	b.kids = append(b.kids, node)
	node.SetParent(b)
}

func (b *Branch) First() Node {
	if len(b.kids) == 0 {
		return nil
	}
	return b.kids[0]
}

func (b *Branch) Insert(node Node, pos int) bool {
	if pos > len(b.kids) {
		return false
	}
	kids := append(b.kids[:pos], node)
	kids = append(kids, b.kids[pos:]...)
	b.kids = kids
	return true
}

type Leaf struct {
	typ    int
	key    string
	val    string
	parent *Branch
}

func NewLeaf(key, val string) *Leaf {
	return &Leaf{
		key: key,
		val: val,
		typ: LeafType,
	}
}

func NewLongLeaf(key, val string) *Leaf {
	leaf := NewLeaf(key, val)
	leaf.typ = LongLeafType
	return leaf
}

func NewComment(val string) *Leaf {
	leaf := NewLeaf("#", val)
	leaf.typ = CommentType
	return leaf
}

func (l *Leaf) String() string {
	str := ""
	switch l.typ {
	case LeafType:
		str += "-"
		if l.key != "" {
			str += " " + l.key
		}
		if l.val != "" {
			str += ": " + l.val
		}
		str += "\n"
	case LongLeafType:
		str = "- " + l.key + ":\n\n"
		str += l.val
	case TextType, CommentType:
		str += l.val
	}
	return str
}

func (l *Leaf) ToJSON() string {
	val := strings.Replace(l.val, "\n", "\\n", -1)
	val = strings.Replace(val, "\"", "\\\"", -1)
	return "{ \"" + l.key + "\": \"" + val + "\" }"
}

func (l *Leaf) Parent() *Branch          { return l.parent }
func (l *Leaf) SetParent(branch *Branch) { l.parent = branch }

func (l *Leaf) Type() int       { return l.typ }
func (l *Leaf) SetType(typ int) { l.typ = typ }

func (l *Leaf) Key() string       { return l.key }
func (l *Leaf) SetKey(key string) { l.key = key }

func (l *Leaf) Value() string          { return l.val }
func (l *Leaf) SetValue(val string)    { l.val = val }
func (l *Leaf) AppendValue(val string) { l.val += val }
