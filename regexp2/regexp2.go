package regexp2

import (
	"strings"
)

// parsing
type sub string

type expr []sub

func (e *expr) pop() (s sub) {
	s = (*e)[len(*e)-1]
	*e = (*e)[:len(*e)-1]
	return
}

func (e *expr) push(s sub) {
	*e = append(*e, s)
}

func (e expr) String() string {
	var b strings.Builder

	for i, s := range e {
		b.WriteString(string(s))
		if i > 0 {
			b.WriteRune('.')
		}
	}
	return b.String()
}

func re2post(re string) string {
	var (
		rd  = strings.NewReader(re)
		buf expr
		alt int
	)

	for {
		r, _, err := rd.ReadRune()
		if err != nil {
			break
		}
		switch r {
		case '(':
			var (
				s string
				n = 1
			)
			for {
				r, _, err := rd.ReadRune()
				if err != nil {
					panic("paren")
				}
				switch r {
				case '(':
					n++
				case ')':
					n--
				}
				if n < 0 {
					panic("paren")
				}
				if n == 0 {
					break
				}
				s += string(r)
			}
			buf.push(sub(re2post(s)))
		case '|':
			alt++
		case '+', '*', '?':
			s := buf.pop()
			buf.push(sub(string(s) + string(r)))
		default:
			buf.push(sub(string(r)))
		}
	}

	for ; alt > 0; alt-- {
		s2 := buf.pop()
		s1 := buf.pop()
		buf.push(sub(string(s1) + string(s2) + "|"))
	}
	return buf.String()
}

const (
	Char = iota
	Match
	Jmp
	Split
)

// compiling
type inst struct {
	char   rune
	opcode int
	x, y   *inst
	tid    int
}

type frag struct {
	start *inst
	olist []**inst
}

func (f *frag) patch(s *inst) {
	for _, o := range f.olist {
		*o = s
	}
}

type stack []*frag

func (t *stack) pop() (f *frag) {
	f = (*t)[len(*t)-1]
	*t = (*t)[:len(*t)-1]
	return
}

func (t *stack) push(f *frag) {
	*t = append(*t, f)
}

type prog struct {
	start *inst
}

func post2nfa(postfix string) prog {
	var (
		rd = strings.NewReader(postfix)
		sp stack
	)
	for {
		r, _, err := rd.ReadRune()
		if err != nil {
			break
		}
		switch r {
		default:
			s := &inst{r, Char, nil, nil, 0}
			sp.push(&frag{s, []**inst{&s.x}})
		case '.':
			e2 := sp.pop()
			e1 := sp.pop()
			e1.patch(e2.start)
			sp.push(&frag{e1.start, e2.olist})
		case '|':
			e2 := sp.pop()
			e1 := sp.pop()
			s := &inst{0, Split, e1.start, e2.start, 0}
			sp.push(&frag{s, append(e1.olist, e2.olist...)})
		case '+':
			e := sp.pop()
			s := &inst{0, Split, e.start, nil, 0}
			e.patch(s)
			sp.push(&frag{e.start, []**inst{&s.y}})
		case '*':
			e := sp.pop()
			s := &inst{0, Split, e.start, nil, 0}
			e.patch(s)
			sp.push(&frag{s, []**inst{&s.y}})
		case '?':
			e := sp.pop()
			s := &inst{0, Split, e.start, nil, 0}
			sp.push(&frag{e.start, append(e.olist, &s.y)})
		}
	}

	e := sp.pop()
	if len(sp) != 0 {
                panic("bug")
	}
	e.patch(&inst{0, Match, nil, nil, 0})
	return prog{e.start}
}

// matching
type thread struct {
	pc *inst
}

type threads []thread

func (ts *threads) add(t thread, tid int) {
	if t.pc.tid == tid {
		return
	}

	t.pc.tid = tid
	*ts = append(*ts, t)

	switch t.pc.opcode {
	case Jmp:
		ts.add(thread{t.pc.x}, tid)
	case Split:
		ts.add(thread{t.pc.x}, tid)
		ts.add(thread{t.pc.y}, tid)
	}
	return
}

var tid int

func execute(p prog, input string) bool {
	var (
		clist, nlist *threads
		pc           *inst
	)

	tid++
	clist = new(threads)
	nlist = new(threads)
	clist.add(thread{p.start}, tid)

	rd := strings.NewReader(input)
	for {
		r, _, err := rd.ReadRune()
		tid++
		for i := range *clist {
			pc = (*clist)[i].pc
			switch pc.opcode {
			case Char:
				if pc.char == r {
					nlist.add(thread{pc.x}, tid)
				}
			case Match:
				return true
			}
		}
		clist, nlist = nlist, clist
		*nlist = (*nlist)[:0]
		if err != nil {
			break
		}
	}
	return false
}
