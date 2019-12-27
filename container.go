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
		provider.state = RESOLVING
		producerType := reflect.TypeOf(provider.producer)
		producerCallArgs := make([]reflect.Value, producerType.NumIn())
		for i := 0;  i<producerType.NumIn(); i++ {
			paramType := producerType.In(i)
			var err error
			producerCallArgs[i], err = container.find(paramType)
			if nil != err {
				return reflect.NewAt(provider.producedType, nil), err
			}
		}
		providerCallResults := reflect.ValueOf(provider.producer).Call(producerCallArgs)
		provider.value = providerCallResults[0]
		provider.state = RESOLVED
		return provider.value, nil
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


func getType(producer interface{}) reflect.Type {
	return reflect.TypeOf(producer).Out(0)
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

func(container *Container) With(target interface{}) error {
	t := reflect.TypeOf(target)
	v := reflect.ValueOf(target)
	parameter := make([]reflect.Value, 1)
	for i := 0;  i<t.NumIn(); i++ {
		paramType := t.In(i)
		var err error
		parameter[i], err = container.find(paramType)
		if err != nil {
			return err
		}
	}
	v.Call(parameter)
	return nil
}

func (container *Container) find(t reflect.Type) (reflect.Value, error) {
	provider, providerExists := container.providers[t]
	if !providerExists {
		return reflect.NewAt(t.Elem(), nil), errors.New(fmt.Sprintf("no Provider found for type, %s", t))
	} else {
		value, err :=  provider.get(container)
		if err != nil {
			return reflect.NewAt(t.Elem(), nil), err
		}
		return value, nil
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

