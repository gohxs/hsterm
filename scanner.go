package termu

type Scanner struct {
	term   *Term
	output string
	err    error
}

func NewScanner(t *Term) *Scanner {
	return &Scanner{t, "", nil}
}

func (s *Scanner) Scan() bool {
	line, err := s.term.ReadLine()
	if err != nil {
		s.err = err
		return false
	}
	s.output = line
	return true
}

func (s *Scanner) Text() string {
	return s.output
}

func (s *Scanner) Err() error {
	return s.err
}
