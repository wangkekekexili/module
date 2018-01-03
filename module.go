// package module provides a way to recursively initialize fields of a struct.
package module

import (
	"errors"
	"reflect"
)

// Load recursively initializes a module.
// A module is defined as a pointer to a struct implementing loadable.
// It recursively checks structs or pointers to structs and any module will be initialized
// and Load() will be called.
// Modules are singletons.
func Load(m interface{}) error {
	// m must be a pointer to struct and implements loadable.
	t := reflect.TypeOf(m)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return errors.New("expecting pointer to a struct")
	}

	g := newGraph(reflect.ValueOf(m))
	g.init()
	err := g.load()
	if err != nil {
		return err
	}

	return nil
}

type loadable interface {
	Load() error
}

type node struct {
	m    reflect.Value // Either Struct or Ptr to Struct.
	deps []*node
}

type graph struct {
	root *node

	// initialized saves initialized module values.
	// Key is the type of the module, of kind ptr to struct.
	// Value is the initialized value of the module, of kind ptr to struct.
	initialized map[reflect.Type]reflect.Value

	// loaded saves a set of loaded modules.
	// Key is the type of the module, of kind struct.
	loaded map[reflect.Type]bool
}

func newGraph(m reflect.Value) *graph {
	g := &graph{
		root:        &node{m: m},
		initialized: make(map[reflect.Type]reflect.Value),
		loaded:      make(map[reflect.Type]bool),
	}
	return g
}

func (g *graph) init() {
	if g.root == nil {
		return
	}
	g.initInternal(g.root)
}

func (g *graph) initInternal(n *node) {
	v := n.m
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i != v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := fieldValue.Type()
		if !fieldValue.CanSet() {
			continue
		}

		switch {
		case fieldType.Kind() == reflect.Struct:
			next := &node{m: fieldValue}
			n.deps = append(n.deps, next)
			g.initInternal(next)
		case fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct:
			newFieldValue, ok := g.initialized[fieldType]
			if !ok {
				newFieldValue = reflect.New(fieldType.Elem())
				g.initialized[fieldType] = newFieldValue
			}
			fieldValue.Set(newFieldValue)

			next := &node{m: fieldValue}
			n.deps = append(n.deps, next)
			g.initInternal(next)
		}
	}
}

func (g *graph) load() error {
	if g.root == nil {
		return nil
	}
	return g.loadInternal(g.root)
}

func (g *graph) loadInternal(n *node) error {
	// Load dependencies first.
	for _, n := range n.deps {
		err := g.loadInternal(n)
		if err != nil {
			return err
		}
	}

	m, ok := n.m.Interface().(loadable)
	if !ok {
		return nil
	}
	typ := reflect.TypeOf(m).Elem()
	loaded := g.loaded[typ]
	if loaded {
		return nil
	}
	err := m.Load()
	if err != nil {
		return err
	}
	g.loaded[typ] = true

	return nil
}
