package main

import (
	"testing"
)

type Repository struct {}

func (repository Repository) getValue() string {
	return "1100101"
}

func TestContainerShouldCallwithFunction (t *testing.T) {
	c := NewContainer()
	called := false
	c.With(func(container *Container) {
		called = true
	})
	if !called {
		t.Error("Handler should have been called but was not")
	}
}

func TestContainerProvidesASelfReference (t *testing.T) {
	c := NewContainer()
	c.With(func(container *Container) {
		if container == nil {
			t.Error("Container should be set")
		}
	})
}

func TestContainerConstructorFunction(t *testing.T) {
	container := NewContainer()

	container.Provide(func() *Repository {
		return new(Repository)
	})

	container.With(func(repo *Repository) {
		if repo.getValue() != "1100101" {
			t.Error("repository fn was not called correctly")
		}
	})
}

func TestContainerCachesProvidedComponents(t *testing.T) {
	container := NewContainer()
	type A struct {}
	counter := 0
	container.Provide(func() *A {
		counter ++
		return new(A)
	})
	container.With(func(a *A) {})
	container.With(func(a *A) {})
	if counter != 1 {
		t.Error("producer is should ne called exactly 1 time")
	}
}

func TestContainerResolvesTransitiveDependencies(t *testing.T) {
	container := NewContainer()

	container.Provide(func(context *Container) *Repository {
		return new(Repository)
	})

	err := container.With(func(repo *Repository) {
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
	container.Provide(func(a *A) *B {
		return new(B)
	})
	container.Provide(func(a *B) *A {
		return new(A)
	})
	err := container.With(func(a *A) {})
	if err == nil {
		t.Error("call to cyclic dependend compoenent should return an error")
	}
}