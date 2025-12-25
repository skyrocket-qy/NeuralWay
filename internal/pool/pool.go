package pool

// Pool manages a pool of objects of type T.
// It is designed to reduce memory allocation overhead by reusing objects.
// T should be a pointer type for most effective use.
type Pool[T any] struct {
	items   []T
	factory func() T
	reset   func(T)
}

// New creates a new Pool for type T.
// factory: function to create a new instance of T when the pool is empty.
// reset: optional function to reset an instance of T when retrieved from the pool (or returned).
func New[T any](factory func() T, reset func(T)) *Pool[T] {
	return &Pool[T]{
		items:   make([]T, 0, 64),
		factory: factory,
		reset:   reset,
	}
}

// Get retrieves an item from the pool.
// If the pool is empty, a new item is created using the factory.
func (p *Pool[T]) Get() T {
	if len(p.items) == 0 {
		return p.factory()
	}

	// Pop from stack (end of slice)
	lastIdx := len(p.items) - 1
	item := p.items[lastIdx]
	p.items = p.items[:lastIdx]

	return item
}

// Put returns an item to the pool.
// The reset function is called on the item before storing it.
func (p *Pool[T]) Put(item T) {
	if p.reset != nil {
		p.reset(item)
	}

	p.items = append(p.items, item)
}

// Size returns the current number of items in the pool.
func (p *Pool[T]) Size() int {
	return len(p.items)
}

// Clear removes all items from the pool.
func (p *Pool[T]) Clear() {
	p.items = p.items[:0]
}
