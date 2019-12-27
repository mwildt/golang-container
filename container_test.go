package container

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type Repository struct {}

func (repository Repository) getValue() string {
	return "1100101"
}

func TestGetTypeForErrorReturningFuncs(t *testing.T) {
	type A struct {}
	res := getReturnType(func() (error, *A) {
		return nil, nil
	})
	if res != reflect.TypeOf(&A{}) {
		t.Errorf("type should be %T but was %s", &A{}, res)
	}
}

func TestGetTypeForErrorReturningFuncsExecuteTypeFirst(t *testing.T) {
	type A struct {}
	res := getReturnType(func() (*A, error) {
		return nil, nil
	})
	if res != reflect.TypeOf(&A{}) {
		t.Errorf("type should be %T but was %s", &A{}, res)
	}
}

func TestContainerHandlesMultiReturningProviders(t *testing.T) {
	type A struct {}
	container := NewContainer()
	container.Singleton(func() (error, *A){
		return nil, &A{}
	})
	called := false
	container.Execute(func(a *A) {
		called = true
	})
	if !called {
		t.Error("Handler should have been called but was not")
	}
}

func TestContainerHandlesMultiReturningProvidersExecuteTypeFirst(t *testing.T) {
	type A struct {}
	container := NewContainer()
	container.Singleton(func() (*A, error){
		return &A{}, nil
	})
	called := false
	container.Execute(func(a *A) {
		called = true
	})
	if !called {
		t.Error("Handler should have been called but was not")
	}
}

func TestContainerPropagatesErrorInProvider(t *testing.T) {
	type A struct {}
	container := NewContainer()
	container.Singleton(func() (*A, error) {
		return nil, errors.New("FEHLER")
	})
	err := container.Execute(func(a *A) {})
	if nil == err {
		t.Error("Execute-Execute should have returned an error but didnt ")
		}
	if err.Error() != "FEHLER" {
		t.Error(fmt.Sprintf("message should be FEHLER but was %s", err.Error()))
	}
}

func TestContainerShouldCallwithFunction (t *testing.T) {
	c := NewContainer()
	called := false
	c.Execute(func(container *Container) {
		called = true
	})
	if !called {
		t.Error("Handler should have been called but was not")
	}
}

func TestContainerProvidesASelfReference (t *testing.T) {
	c := NewContainer()
	c.Execute(func(container *Container) {
		if container == nil {
			t.Error("Container should be set")
		}
	})
}

func TestContainerConstructorFunction(t *testing.T) {
	container := NewContainer()
	container.Singleton(func() *Repository {
		return new(Repository)
	})
	container.Execute(func(repo *Repository) {
		if repo.getValue() != "1100101" {
			t.Error("repository fn was not called correctly")
		}
	})
}

func TestContainerCachesProvidedComponents(t *testing.T) {
	container := NewContainer()
	type A struct {}
	counter := 0
	container.Singleton(func() *A {
		counter ++
		return new(A)
	})
	container.Execute(func(a *A) {})
	container.Execute(func(a *A) {})
	if counter != 1 {
		t.Error("producer is should ne called exactly 1 time")
	}
}

func TestContainerResolvesTransitiveDependencies(t *testing.T) {
	container := NewContainer()
	container.Singleton(func(context *Container) *Repository {
		return new(Repository)
	})
	err := container.Execute(func(repo *Repository) {
		if repo.getValue() != "1100101" {
			t.Error("repository fn was not called correctly")
		}
	})
	if err != nil {
		t.Error("err should ne nil, but is not")
	}
}

func TestContaineShouldFailOnCyclicDependency(t *testing.T) {
	container := NewContainer()
	type A struct {}
	type B struct {}
	container.Singleton(func(a *A) *B {
		return new(B)
	})
	container.Singleton(func(a *B) *A {
		return new(A)
	})
	err := container.Execute(func(a *A) {})
	if err == nil {
		t.Error("call to cyclic dependent component should return an error")
	}
}

func TestContainerPropagatesErrorsFromCalles(t *testing.T) {
	msg := "FEHLER"
	container := NewContainer()
	err := container.Execute(func(c *Container) error {
		return errors.New(msg)
	})
	if nil == err {
		t.Error("Execute-Execute should have returned an error but didnt ")
	}
	if err.Error() != msg {
		t.Error(fmt.Sprintf("message should be %s but was %s", msg, err.Error()))
	}
}

func TestContainerCallCallsNeedToBeOnFunctions(t *testing.T) {
	type A struct {}
	container := NewContainer()
	err := container.Execute(&A{})
	if nil == err {
		t.Error("Execute-Execute should have returned an error but didnt ")
	}
	if err.Error() != "target is not of Kind reflect.Func" {
		t.Error(fmt.Sprintf("message should be 'target is not of Kind reflect.Func' but was %s", err.Error()))
	}
}

func TestContainerProvideCallsNeedToBeOnFunctions(t *testing.T) {
	type A struct {}
	container := NewContainer()
	err := container.Singleton(&A{})
	if nil == err {
		t.Error("Execute-Execute should have returned an error but didnt ")
	}
	if err.Error() != "producer is not of Kind reflect.Func" {
		t.Error(fmt.Sprintf("message should be 'producer is not of Kind reflect.Func' but was %s", err.Error()))
	}
}