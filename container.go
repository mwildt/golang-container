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

func NewContainer() *Container {
	container := new(Container)
	container.providers = make(map[reflect.Type]*provider)
	container.Singleton(func() *Container {
		return container
	})
	return container
}

/**
 * registers a new provider to the containers context using a producer function
 */
func (container *Container) Singleton(producer interface{}) error {
	if reflect.TypeOf(producer).Kind() != reflect.Func {
		return errors.New("producer is not of Kind reflect.Func")
	}
	provider := newProvider(producer)
	container.providers[provider.producedType] = provider
	return nil
}

/**
 * Invokes a target-function with params resolved from the containers context
 */
func(container *Container) With(target interface{}) error {
	if reflect.TypeOf(target).Kind() != reflect.Func {
		return errors.New("target is not of Kind reflect.Func")
	}
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

func (container *Container) find(t reflect.Type) (reflect.Value, error) {
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