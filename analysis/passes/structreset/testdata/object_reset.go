package testdata

type skill struct {
	ss string
}

func (s *skill)Reset() {

}

type reuse interface {
	Reset()
}

type st1 struct {
	stt int
}

type refCount struct {

}


type people struct {
	st1
	refCount

	a int
	b string
	c []byte
	d map[int]int
	e *skill
	f skill
	g [4]bool
	h [4]bool

	x [3]bool
	y [3]bool
	z [3]bool
}

func (p *people) clearXYZ() {
	for i, length := 0, len(p.x); i < length; i++ {
		p.x[i] = false
		//p.y[i] = false
		p.z[i] = false
	}
}