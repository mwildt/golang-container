# go-container

A very simple DI-Container for Golang.

```go

type MyService struct {}

func (s *MyService) someAction(){
	...
}

func myServiceFactory() (*MyService, error) {
    return &MyService{}, nil
}
```

##  Register a factory function and execute a function with registered components as parameters

```go
func main() {
    container := NewContainer()
    container.Singleton(myServiceFactory)
    container.Execute(func(service *MyService) {
    	myService.someAction()
    })
}
```