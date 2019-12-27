package container

import (
	"errors"
	"fmt"
	"reflect"
)

/**
 * The Dependency-Container
 */
type Container struct {
	providers map[reflect.Type]*provider
}

/**
 * util states for detecting cyclic dependencies
 */
type providerState int
const (
	UNRESOLVED providerState = iota
	RESOLVING
	RESOLVED
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

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
			return provider.resovle(container)
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

func (provider *provider) resovle(container *Container) (reflect.Value, error) {
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

func getType(producer interface{}) reflect.Type {
	t := reflect.TypeOf(producer)
	for i := 0; i < t.NumOut(); i++ {
		out := t.Out(i)
		if !out.Implements(errorInterface) {
			return out
		}
	}
	return nil
}

func newProvider(producer interface{}) *provider{
	producedType := getType(producer)
	p := new(provider)
	p.producer = producer
	p.producedType = producedType
	p.state = UNRESOLVED
	p.value = reflect.NewAt(producedType, nil)
	return p
}

func NewContainer() *Container {
	container := new(Container)
	container.providers = make(map[reflect.Type]*provider)
	container.Provide(func() *Container {
		return container
	})
	return container
}

func (container *Container) Provide(producer interface{}) {
	provider := newProvider(producer)
	container.providers[provider.producedType] = provider
}

func (container *Container) call(target interface{}) ([]reflect.Value, error) {
	targetType := reflect.TypeOf(target)
	parameter := make([]reflect.Value, targetType.NumIn())
	for i := 0;  i< targetType.NumIn(); i++ {
		paramType := targetType.In(i)
		var err error
		parameter[i], err = container.find(paramType)
		if err != nil {
			return nil, err
		}
	}
	return reflect.ValueOf(target).Call(parameter), nil
}

func(container *Container) With(target interface{}) error {
	values, err := container.call(target)
	if nil != err {
		return err
	}
	for _, returnValue := range values {
		if returnValue.Type().Implements(errorInterface) && !returnValue.IsNil() {
			return returnValue.Interface().(error)
		}
	}
	return nil
}

func (container *Container) find(t reflect.Type) (reflect.Value, error) {
	//log.Printf("Container::find %s\n", t)
	provider, providerExists := container.providers[t]
	if !providerExists {
		return reflect.NewAt(t.Elem(), nil), errors.New(fmt.Sprintf("no Provider found for type, %s", t))
	} else {
		return provider.get(container)
	}
}

func  (container *Container)  findProviderFor(t reflect.Type) interface{} {
	for _, provider := range container.providers {
		if reflect.TypeOf(provider).Out(0) == t {
			return provider
		}
	}
	return nil
}