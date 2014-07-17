package dgrl

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// Node is the interface common to branches and leaves
type Node interface {
	Parent() *Branch
	SetParent(parent *Branch)
	Type() int
	Key() string
	SetKey(key string)
	String() string
	ToJSON() string

	write(w *bufio.Writer)
	writeJSON(w *bufio.Writer)
}

// Branch is a branch tree node
type Branch struct {
	parent *Branch
	key    string
	kids   []Node
}

// NewRoot constructs a top level Branch instance
func NewRoot() *Branch {
	return NewBranch("")
}

// NewBranch constructs a Branch instance
func NewBranch(key string) *Branch {
	return &Branch{
		key:  key,
		kids: []Node{},
	}
}

// String returns the text represenation of the Branch
func (b *Branch) String() string {
	str := ""
	prevTyp := NoType
	lvl := b.Level()
	if lvl > 0 {
		str = strings.Repeat("=", lvl) + " " + b.key + "\n"
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

// Write writes the text reperesentation of the Branch to an io.Writer
func (b *Branch) Write(w io.Writer) error {
	buf := bufio.NewWriter(w)
	if buf == nil {
		return errors.New("error creating bufio.Writer")
	}
	b.write(buf)
	buf.Flush()
	return nil
}

func (b *Branch) write(w *bufio.Writer) {
	prevTyp := NoType
	lvl := b.Level()
	if lvl > 0 {
		for i := 0; i < lvl; i++ {
			w.WriteByte('=')
		}
		w.WriteByte(' ')
		w.WriteString(b.key)
		w.WriteByte('\n')
		prevTyp = CommentType // fake type to get a newline between branch and first kid
	}
	for _, kid := range b.kids {
		typ := kid.Type()
		if prevTyp == BranchType && typ != BranchType {
			w.WriteByte('\n')
			for i := 0; i < lvl+1; i++ {
				w.WriteByte('=')
			}
			w.WriteByte('\n')
		}
		if prevTyp != NoType && !(typ == LeafType && prevTyp == LeafType) {
			w.WriteByte('\n')
		}
		if typ == TextType && (prevTyp == TextType || prevTyp == LongLeafType) {
			w.WriteString("-\n\n")
		}
		kid.write(w)
		prevTyp = typ
	}
}

// ToJSON returns a JSON representation of the Branch
func (b *Branch) ToJSON() string {
	str := "{ \"" + b.key + "\": [ "
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

// WriteJSON writes a JSON representation of the Branch to an io.Writer
func (b *Branch) WriteJSON(w io.Writer) error {
	buf := bufio.NewWriter(w)
	if buf == nil {
		return errors.New("error creating bufio.Writer")
	}
	b.writeJSON(buf)
	buf.Flush()
	return nil
}

func (b *Branch) writeJSON(w *bufio.Writer) {
	w.WriteString("{ \"")
	w.WriteString(b.key)
	w.WriteString("\": [ ")
	for i, kid := range b.kids {
		kid.writeJSON(w)
		if i < len(b.kids)-1 {
			w.WriteByte(',')
		}
		w.WriteByte(' ')
	}
	w.WriteString("] }")
}

// Type returns the node type of Branch
func (b *Branch) Type() int {
	return BranchType
}

// Level returns the nesting level of the Branch
func (b *Branch) Level() int {
	i := 0
	for ; b.parent != nil; i++ {
		b = b.parent
	}
	return i
}

func (b *Branch) Key() string       { return b.key }
func (b *Branch) SetKey(key string) { b.key = key }

func (b *Branch) Parent() *Branch          { return b.parent }
func (b *Branch) SetParent(parent *Branch) { b.parent = parent }

func (b *Branch) NumKids() int { return len(b.kids) }
func (b *Branch) Kids() []Node { return b.kids }

// Append appends a node to the Branch
func (b *Branch) Append(node Node) {
	b.kids = append(b.kids, node)
	node.SetParent(b)
}

// First returns the first child node in the Branch
func (b *Branch) First() Node {
	if len(b.kids) == 0 {
		return nil
	}
	return b.kids[0]
}

// Insert inserts a node into the branch at the given position
func (b *Branch) Insert(node Node, pos int) bool {
	if pos > len(b.kids) {
		return false
	}
	b.kids = append(b.kids[:pos], append([]Node{node}, b.kids[pos:]...)...)
	return true
}

// Leaf is a leaf tree node
type Leaf struct {
	typ    int
	key    string
	val    string
	parent *Branch
}

// NewLeaf constructs a Leaf instance
func NewLeaf(key, val string) *Leaf {
	return &Leaf{
		key: key,
		val: val,
		typ: LeafType,
	}
}

// NewLongLeaf constructs a long type Leaf
func NewLongLeaf(key, val string) *Leaf {
	val = strings.TrimRight(val, "\n") + "\n"
	leaf := NewLeaf(key, val)
	leaf.typ = LongLeafType
	return leaf
}

// NewLongLeaf constructs a text type Leaf
func NewText(val string) *Leaf {
	val = strings.TrimRight(val, "\n") + "\n"
	leaf := NewLeaf("", val)
	leaf.typ = TextType
	return leaf
}

// NewLongLeaf constructs a comment type Leaf
func NewComment(val string) *Leaf {
	val = strings.TrimRight(val, "\n") + "\n"
	leaf := NewLeaf("#", val)
	leaf.typ = CommentType
	return leaf
}

// String returns the text representation of the Leaf
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

func (l *Leaf) write(w *bufio.Writer) {
	switch l.typ {
	case LeafType:
		w.WriteByte('-')
		if l.key != "" {
			w.WriteByte(' ')
			w.WriteString(l.key)
		}
		if l.val != "" {
			w.WriteString(": ")
			w.WriteString(l.val)
		}
		w.WriteByte('\n')
	case LongLeafType:
		w.WriteString("- ")
		w.WriteString(l.key)
		w.WriteString(":\n\n")
		w.WriteString(l.val)
	case TextType, CommentType:
		w.WriteString(l.val)
	}
}

// ToJSON returns a JSON representation of the Leaf
func (l *Leaf) ToJSON() string {
	val := strings.Replace(l.val, "\n", "\\n", -1)
	val = strings.Replace(val, "\"", "\\\"", -1)
	return "{ \"" + l.key + "\": \"" + val + "\" }"
}

func (l *Leaf) writeJSON(w *bufio.Writer) {
	val := strings.Replace(l.val, "\n", "\\n", -1)
	val = strings.Replace(val, "\"", "\\\"", -1)
	w.WriteString("{ \"")
	w.WriteString(l.key)
	w.WriteString("\": \"")
	w.WriteString(val)
	w.WriteString("\" }")
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
