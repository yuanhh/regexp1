package regexp1

import (
	"bytes"
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
		rd  = bytes.NewReader([]byte(re))
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
				n int
			)
			for n = 1;; {
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

// compiling
type stateType int

const (
	matchState stateType = iota
	splitState
	literState
)

type state struct {
	char      rune
	out, out1 *state
	typ       stateType
	index     int
}

var matchstate = &state{0, nil, nil, matchState, 0}

type frag struct {
	start *state
	olist []**state
}

func (f *frag) patch(s *state) {
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

func post2nfa(postfix string) *state {
	var (
		rd = bytes.NewReader([]byte(postfix))
		sp stack
	)
	for {
		r, _, err := rd.ReadRune()
		if err != nil {
			break
		}
		switch r {
		default:
			s := &state{r, nil, nil, literState, 0}
			sp.push(&frag{s, []**state{&s.out}})
		case '.':
			e2 := sp.pop()
			e1 := sp.pop()
			e1.patch(e2.start)
			sp.push(&frag{e1.start, e2.olist})
		case '|':
			e2 := sp.pop()
			e1 := sp.pop()
			s := &state{0, e1.start, e2.start, splitState, 0}
			sp.push(&frag{s, append(e1.olist, e2.olist...)})
		case '+':
			e := sp.pop()
			s := &state{0, e.start, nil, splitState, 0}
			e.patch(s)
			sp.push(&frag{e.start, []**state{&s.out1}})
		case '*':
			e := sp.pop()
			s := &state{0, e.start, nil, splitState, 0}
			e.patch(s)
			sp.push(&frag{s, []**state{&s.out1}})
		case '?':
			e := sp.pop()
			s := &state{0, e.start, nil, splitState, 0}
			sp.push(&frag{e.start, append(e.olist, &s.out1)})
		}
	}

	e := sp.pop()
	if len(sp) != 0 {
		return nil
	}
	e.patch(matchstate)
	return e.start
}

// matching
type list []*state

var listid = 1

func (l *list) addstate(s *state) {
	if s == nil || s.index == listid {
		return
	}
	s.index = listid
	if s.typ == splitState {
		l.addstate(s.out)
		l.addstate(s.out1)
		return
	}
	*l = append(*l, s)
	return
}

func ismatch(l list) bool {
	for _, e := range l {
		if e.typ == matchState {
			return true
		}
	}
	return false
}

func step(clist, nlist *list, c rune) {
	listid++
	*nlist = (*nlist)[:0]
	for _, s := range *clist {
		if s.char == c {
			nlist.addstate(s.out)
		}
	}
}

func match(start *state, input string) bool {
	rd := bytes.NewReader([]byte(input))
	l1 := new(list)
	l2 := new(list)
        listid++
        l1.addstate(start)

	for clist, nlist := l1, l2;; clist, nlist = nlist, clist {
		r, _, err := rd.ReadRune()
		if err != nil {
                        return ismatch(*clist)
		}
		step(clist, nlist, r)
	}
        panic("unreachable")
}
