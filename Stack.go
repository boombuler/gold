package gold

// Stack is a basic LIFO stack that resizes as needed.
type stack struct {
	nodes []interface{}
	count int
}

func newStack() *stack {
	return &stack{nodes: make([]interface{}, 10)}
}

// Push adds a node to the stack.
func (s *stack) Push(n interface{}) {
	if s.count >= len(s.nodes) {
		nodes := make([]interface{}, len(s.nodes)*2)
		copy(nodes, s.nodes)
		s.nodes = nodes
	}
	s.nodes[s.count] = n
	s.count++
}

// Pop removes and returns a node from the stack in last to first order.
func (s *stack) Pop() interface{} {
	if s.count == 0 {
		return nil
	}
	node := s.nodes[s.count-1]
	s.count--
	return node
}

func (s *stack) Len() int {
	return s.count
}

func (s *stack) Peek() interface{} {
	return s.nodes[s.count-1]
}
