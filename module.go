// package module provides a way to recursively initialize fields of a struct.
package module

import (
	"errors"
	"reflect"
)

func Load(m interface{}) error {
	// m must be a pointer to struct.
	t := reflect.TypeOf(m)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return errors.New("only pointer to struct can be passed to Load")
	}

	// Initialized values will be saved in visited so that modules are singletons.
	// Key is the type of the module, of kind struct.
	// Value is the initialized value of the module, of kind struct.
	visited := make(map[reflect.Type]reflect.Value)
	v := reflect.ValueOf(m).Elem()
	visited[t] = v

	tree := &depTree{}
	current, ok := m.(loadable)
	if ok {
		tree.root = &depNode{m: current}
	}

	err := populate(v, visited, tree.root)
	if err != nil {
		return err
	}

	err = tree.load()
	if err != nil {
		return err
	}

	return nil
}

type loadable interface {
	Load() error
}

var loadableType = reflect.TypeOf((*loadable)(nil)).Elem()

type depTree struct {
	root   *depNode
	loaded map[reflect.Type]bool
}

type depNode struct {
	m    loadable
	deps []*depNode
}

func (t *depTree) load() error {
	if t.root == nil {
		return nil
	}
	t.loaded = make(map[reflect.Type]bool)
	return t.loadNode(t.root)
}

func (t *depTree) loadNode(n *depNode) error {
	// Load dependencies first.
	for _, n := range n.deps {
		err := t.loadNode(n)
		if err != nil {
			return err
		}
	}

	typ := reflect.TypeOf(n.m)
	loaded := t.loaded[typ]
	if loaded {
		return nil
	}
	err := n.m.Load()
	if err != nil {
		return err
	}
	t.loaded[typ] = true
	return nil
}

func populate(v reflect.Value, visited map[reflect.Type]reflect.Value, dep *depNode) error {
	if v.Kind() != reflect.Struct {
		return errors.New("only struct can be passed to populate")
	}
	for i := 0; i != v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := fieldValue.Type()
		if !fieldValue.CanSet() {
			continue
		}

		switch {
		case fieldType.Kind() == reflect.Struct:
			populate(fieldValue, visited, dep)
		case fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct:
			fieldTypeElem := fieldType.Elem()           // From Ptr to Struct.
			newFieldValue, ok := visited[fieldTypeElem] // Check if it's already initialized. newFieldValue will be of type Struct at this point.
			if ok {
				newFieldValue = newFieldValue.Addr() // From Struct to Ptr.
			} else {
				newFieldValue = reflect.New(fieldTypeElem)
				visited[fieldTypeElem] = newFieldValue.Elem()
			}
			fieldValue.Set(newFieldValue)

			if dep != nil && fieldType.Implements(loadableType) {
				newDep := &depNode{m: fieldValue.Interface().(loadable)}
				dep.deps = append(dep.deps, newDep)
				dep = newDep
			}

			populate(newFieldValue.Elem(), visited, dep)
		}
	}
	return nil
}
