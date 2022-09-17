# go-strcmd

The reflection-based package allows users to call methods with automatically parsed parameters, with single line of string.

```go
func test(id int) error {
	fmt.Print(id)
}

func main () {
    strcmd := strcmd.NewStrCmd()
    strcmd.Call(test, "123")
}
```