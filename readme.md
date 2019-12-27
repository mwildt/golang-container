# go-container

A very simple DI-Container for Golang.

## Register Factory Function

```go

type MyService struct {}

func main() {
    container := NewContainer()
    container.Singleton(func() (*MyService, error) {
    	return &MyService{}, nil
    })
}
```