package pool

import (
	"testing"
)

type TestObject struct {
	ID    int
	Value string
}

func TestPool(t *testing.T) {
	counter := 0
	factory := func() *TestObject {
		counter++

		return &TestObject{ID: counter}
	}

	reset := func(o *TestObject) {
		o.Value = ""
	}

	p := New(factory, reset)

	// Test Get (Factory)
	obj1 := p.Get()
	if obj1.ID != 1 {
		t.Errorf("Expected ID 1, got %d", obj1.ID)
	}

	obj2 := p.Get()
	if obj2.ID != 2 {
		t.Errorf("Expected ID 2, got %d", obj2.ID)
	}

	// Test Put
	obj1.Value = "dirty"
	p.Put(obj1)

	if p.Size() != 1 {
		t.Errorf("Expected size 1, got %d", p.Size())
	}

	// Test Get (Reuse)
	obj3 := p.Get()
	if obj3.ID != 1 {
		t.Errorf("Expected reused ID 1, got %d", obj3.ID)
	}

	if obj3.Value != "" {
		t.Errorf("Expected reset value empty, got %s", obj3.Value)
	}

	// Test Size
	if p.Size() != 0 {
		t.Errorf("Expected size 0, got %d", p.Size())
	}
}
