package dgrl

import (
	"bufio"
	"io"
	"strings"
)

// Parser parses Doggerel text into a node tree
type Parser struct {
	tree    *Branch
	branch  *Branch
	leaf    *Leaf
	level   int
	context int
}

// NewParser constructs a Parser instance
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses input into a tree of nodes
func (p *Parser) Parse(input io.Reader) *Branch {
	in := bufio.NewReader(input)
	if in == nil {
		return nil
	}
	p.tree = NewBranch("")
	p.branch = p.tree
	for line, err := in.ReadString('\n'); err != io.EOF; line, err = in.ReadString('\n') {
		if err == nil {
			p.parseLine(line)
		}
	}
	p.parseLine("-\n") // force dangling leaf to terminate
	return p.tree
}

func (p *Parser) parseLine(line string) {
	typ := lineType(line)
	switch {
	case p.context == LongLeafContext && typ != TextType: // new element ends long leaf
		p.leaf.val = strings.TrimSpace(p.leaf.val) + "\n"
		p.context = DefaultContext
	case p.context == CommentContext && typ != CommentType: // non-comment ends comment leaf
		p.context = DefaultContext
	}
	switch typ {
	case BranchType:
		p.parseBranch(line)
	case LeafType:
		p.parseLeaf(line)
	case CommentType:
		p.parseComment(line)
	default:
		p.parseText(line)
	}
}

func (p *Parser) parseBranch(line string) {
	level := branchLevel(line)
	delta := level - p.level
	if delta > 1 { // too far down
		return
	}
	for i := delta; i <= 0 && p.level > 0; i++ {
		p.branch = p.branch.Parent()
		p.level--
	}
	name := strings.TrimSpace(line[level:])
	if name != "" {
		branch := NewBranch(name)
		p.branch.Append(branch)
		p.branch = branch
		p.level++
	}
}

func (p *Parser) parseLeaf(line string) {
	line = line[1:]
	idx := strings.Index(line, ":")
	hasSuffix := (idx >= 0)
	if idx < 0 { // just a key
		idx = len(line) - 1
		if idx < 0 {
			idx = 0
		}
	}
	key, val := strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:])
	if key == "" && val == "" { // drop empty leaf
		return
	}
	typ := LeafType
	if hasSuffix && val == "" {
		typ = LongLeafType
		p.context = LongLeafContext
	}
	leaf := NewLeaf(key, val)
	leaf.typ = typ
	p.branch.Append(leaf)
	p.leaf = leaf
}

func (p *Parser) parseComment(line string) {
	switch p.context {
	case CommentContext:
		p.leaf.AppendValue(line)
	default:
		p.context = CommentContext
		leaf := NewComment(line)
		p.branch.Append(leaf)
		p.leaf = leaf
	}
}

func (p *Parser) parseText(line string) {
	switch p.context {
	case LongLeafContext:
		p.leaf.AppendValue(line)
	default:
		if strings.TrimSpace(line) == "" { // ignore whitespace-only lines
			break
		}
		leaf := NewLeaf("", line)
		leaf.typ = TextType
		p.branch.Append(leaf)
		p.leaf = leaf
		p.context = LongLeafContext
	}
}

func lineType(line string) int {
	typ := TextType
	switch {
	case strings.HasPrefix(line, "="):
		typ = BranchType
	case strings.HasPrefix(line, "-"):
		typ = LeafType
	case strings.HasPrefix(line, "#"):
		typ = CommentType
	}
	return typ
}

func branchLevel(line string) int {
	i := 0
	for ; i < len(line) && line[i] == '='; i++ {
	}
	return i
}
