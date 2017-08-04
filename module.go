// package module provides a way to recursively initialize fields of a struct.
package module

import (
	"errors"
	"reflect"
)

type loadable interface {
	Load() error
}

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

	err := populate(v, visited)
	if err != nil {
		return err
	}

	return nil
}

func populate(v reflect.Value, visited map[reflect.Type]reflect.Value) error {
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
			populate(fieldValue, visited)
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
			populate(newFieldValue.Elem(), visited)
		}
	}
	return nil
}
