package gold

type stack struct {
	top  *element
	size int
}

type element struct {
	value interface{}
	next  *element
}

func (s *stack) Len() int {
	return s.size
}

func (s *stack) Push(value interface{}) {
	s.top = &element{value, s.top}
	s.size++
}

func (s *stack) Pop() (value interface{}) {
	if s.size > 0 {
		value, s.top = s.top.value, s.top.next
		s.size--
		return
	}
	return nil
}

func (s *stack) Peek() interface{} {
	return s.top.value
}
