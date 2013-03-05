package dgrl

import (
	"bufio"
	"io"
	"strings"
)

type Parser struct {
	tree    *Branch
	branch  *Branch
	leaf    *Leaf
	level   int
	context int
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(input io.Reader) Node {
	in := bufio.NewReader(input)
	if in == nil {
		return nil
	}
	p.tree = NewBranch("")
	p.branch = p.tree
	for line, err := in.ReadString('\n'); err != io.EOF;
		line, err = in.ReadString('\n') {

		if err == nil {
			p.parseLine(line)
		}
	}
	p.parseLine("::\n") // force dangling leaf to terminate
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
	if idx < 0 { // didn't find end of key
		return
	}
	key, val := line[:idx], line[idx+1:]
	if key == "" && strings.TrimSpace(val) == "" { // drop empty leaf
		return
	}
	typ := LeafType
	if strings.HasPrefix(val, ":") {
		val = val[1:]
		typ = LongLeafType
		p.context = LongLeafContext
	}
	leaf := NewLeaf(key, strings.TrimSpace(val))
	leaf.typ = typ
	p.branch.Append(leaf)
	p.leaf = leaf
}

func (p *Parser) parseComment(line string) {
	switch p.context {
	case CommentContext:
		p.leaf.AppendVal(line)
	default:
		p.context = CommentContext
		leaf := NewLeaf("#", line)
		leaf.typ = CommentType
		p.branch.Append(leaf)
		p.leaf = leaf
	}
}

func (p *Parser) parseText(line string) {
	switch p.context {
	case LongLeafContext:
		p.leaf.AppendVal(line)
	default:
		if strings.TrimSpace(line) == "" { // ignore whitespace-only lines
			break
		}
		leaf := NewLeaf(".", line)
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
	case strings.HasPrefix(line, ":"):
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
