package container

import (
	"errors"
	"fmt"
	"reflect"
)

/**
 * util states for detecting cyclic dependencies
 */
type providerState int

const (
	UNRESOLVED providerState = iota
	RESOLVING
	RESOLVED
)

/**
 * An internal structure für handling single producer-functions (caching, state-handling)
 */
type provider struct {
	producer interface{}
	producedType reflect.Type
	state providerState
	value reflect.Value
}

func (provider *provider) get(container *Container) (reflect.Value, error) {
	switch provider.state {
		case UNRESOLVED: {
			return provider.resolve(container)
		}
		case RESOLVING: {
			return reflect.NewAt(provider.producedType, nil), errors.New("provider is already resolving: cyclic dependency detected")
		}
		case RESOLVED: {
			return provider.value, nil
		}
	}
	return reflect.NewAt(provider.producedType, nil), errors.New("illegal state: a provider should always have defined state")
}

func (provider *provider) resolve(container *Container) (reflect.Value, error) {
	provider.state = RESOLVING
	providerCallResults, err := container.call(provider.producer)
	if nil != err {
		return reflect.NewAt(provider.producedType, nil), err
	}

	for _, returnElement := range providerCallResults {
		if returnElement.Type().AssignableTo(provider.producedType) && !returnElement.IsNil() {
			provider.value = returnElement
			provider.state = RESOLVED
			return provider.value, nil
		} else if returnElement.Type().Implements(errorInterface) && !returnElement.IsNil() {
			return reflect.NewAt(provider.producedType, nil), returnElement.Interface().(error)
		}
	}
	return reflect.NewAt(provider.producedType, nil), errors.New(fmt.Sprintf("unable to identity correct return value for type %s\n", provider.producedType))
}

func newProvider(producer interface{}) *provider{
	producedType := getReturnType(producer)
	p := new(provider)
	p.producer = producer
	p.producedType = producedType
	p.state = UNRESOLVED
	p.value = reflect.NewAt(producedType, nil)
	return p
}
