package dgrl

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

type Node interface {
	Parent() *Branch
	SetParent(parent *Branch)
	Type() int
	Key() string
	SetKey(key string)
	String() string
	ToJSON() string

	write(w *bufio.Writer)
}

type Branch struct {
	parent *Branch
	key    string
	kids   []Node
}

func NewRoot() *Branch {
	return NewBranch("")
}

func NewBranch(key string) *Branch {
	return &Branch{
		key:  key,
		kids: []Node{},
	}
}

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

func (b *Branch) Key() string       { return b.key }
func (b *Branch) SetKey(key string) { b.key = key }

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
	val = strings.TrimRight(val, "\n") + "\n"
	leaf := NewLeaf(key, val)
	leaf.typ = LongLeafType
	return leaf
}

func NewText(val string) *Leaf {
	val = strings.TrimRight(val, "\n") + "\n"
	leaf := NewLeaf("", val)
	leaf.typ = TextType
	return leaf
}

func NewComment(val string) *Leaf {
	val = strings.TrimRight(val, "\n") + "\n"
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
