# go-strcmd

The reflection-based package allows users to call methods with automatically parsed parameters, with strings.

The most common usecase is interactive cli that execute methods according to input from stdin.

```go
func testInt(i1 int, i2 int) error {
	fmt.Print(i1 + i2)
    return nil
}

func Call () {
    strcmd := strcmd.NewStrCmd()
    strcmd.Call(testInt, []string {"123", "456"})
    strcmd.SplitAndCall(testInt, "123 456")
}

func NamedCall () {
    strcmd := strcmd.NewStrCmd()
    functions := map[string]any{
        "testInt": testInt,
        "testStr": func (arg string) error {
            _, err := fmt.Print(arg)
            return err
        },
    }
    strcmd.CallNamed("testInt 123 456", functions)
    strcmd.CallNamed("testStr '2 whitespaces here'", functions)
    strcmd.CallNamed("testStr \"2 whitespaces here\"", functions)
    strcmd.CallNamed("testStr '\\\''", functions)
}
```

## Parsers

By default, most of the basic types are supported as parameters.

You can also add custom parser by `strcmd.SetParser`.

```go
type testStruct struct {
    a int
}

func addParser () {
    strcmd := strcmd.NewStrCmd()
    strcmd.SetParser(reflect.TypeOf(testStruct{}).Name(),  func(from string) (any, error) {
        parsed, err := strconv.ParseInt(from, 10, 64)
        if err != nil {
            return testStruct { a: int(parsed) }, nil
        } else {
            return int(parsed), nil
        }
    })
}
```