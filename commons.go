package container

import "reflect"

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

/**
 * finds a non error Return-Value
 */
func getReturnType(producer interface{}) reflect.Type {
	t := reflect.TypeOf(producer)
	for i := 0; i < t.NumOut(); i++ {
		out := t.Out(i)
		if !out.Implements(errorInterface) {
			return out
		}
	}
	return nil
}