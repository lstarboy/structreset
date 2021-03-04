package testdata


func (p *people)Reset() {
	p.a = 0
	p.b = ""

	if len(p.c) > 0 {
		p.c = p.c[:0]
	}

	if len(p.d) > 0 {
		for key := range p.d {
			delete(p.d, key)
		}
	}

	if p.e != nil {
		p.e = nil
	}
	p.f.Reset()

	for i := range p.g {
		p.g[i] = false
		p.h[i] = false
	}

	p.clearXYZ()
}
