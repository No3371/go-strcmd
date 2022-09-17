package internal

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type StrCmd struct {
    parsers map[string]func(from string) (any, error)
    strBuilder strings.Builder
}

func NewStrCmd () StrCmd {
	return StrCmd{
		parsers: map[string]func(from string) (any, error){
			reflect.TypeOf(int(0)).Name(): func(from string) (any, error) {
				parsed, err := strconv.ParseInt(from, 10, 64)
				if err != nil {
					return nil, err
				} else {
					return int(parsed), nil
				}
			},
			reflect.TypeOf(uint(0)).Name(): func(from string) (any, error) {
				parsed, err := strconv.ParseUint(from, 10, 64)
				if err != nil {
					return nil, err
				} else {
					return uint(parsed), nil
				}
			},
			reflect.TypeOf(int8(0)).Name(): func(from string) (any, error) {
				return strconv.ParseInt(from, 10, 8)
			},
			reflect.TypeOf(int32(0)).Name(): func(from string) (any, error) {
				return strconv.ParseInt(from, 10, 32)
			},
			reflect.TypeOf(int64(0)).Name(): func(from string) (any, error) {
				return strconv.ParseInt(from, 10, 64)
			},
			reflect.TypeOf(uint8(0)).Name(): func(from string) (any, error) {
				return strconv.ParseUint(from, 10, 8)
			},
			reflect.TypeOf(uint32(0)).Name(): func(from string) (any, error) {
				return strconv.ParseUint(from, 10, 32)
			},
			reflect.TypeOf(uint64(0)).Name(): func(from string) (any, error) {
				return strconv.ParseUint(from, 10, 64)
			},
			reflect.TypeOf(float32(0)).Name(): func(from string) (any, error) {
				return strconv.ParseFloat(from, 32)
			},
			reflect.TypeOf(float64(0)).Name(): func(from string) (any, error) {
				return strconv.ParseFloat(from, 64)
			},
			reflect.TypeOf(byte(0)).Name(): func(from string) (any, error) {
				return strconv.ParseUint(from, 10, 8)
			},
			reflect.TypeOf("").Name(): func(from string) (any, error) {
				return from, nil
			},
		},
	}
}

func (strcmd *StrCmd) ClearParsers () {
    strcmd.parsers = make(map[string]func(from string) (any, error))
}

func (strcmd *StrCmd) RemoveParser (typeName string) {
    delete(strcmd.parsers, typeName)
}

func (strcmd *StrCmd) SetParser (typeName string, parserFunc func(from string) (any, error)) error {
    fType := reflect.TypeOf(parserFunc)
    if fType.Kind() != reflect.Func {
        return fmt.Errorf("supplied parserFunc is not a func")
    }
    if fType.NumOut() != 1 {
        return fmt.Errorf("parserFunc must has only 1 return value")
    }
    if fType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
        return fmt.Errorf("retrun value of parserFunc must be of interface 'error'")
    }
    strcmd.parsers[typeName] = parserFunc
    return nil
}

func (strcmd *StrCmd) CallNamed(cmd string, functions map[string]any) error {
	splitted, err := strcmd.split(cmd)
	if err != nil {
		return err
	}

	if f, registered := functions[splitted[0]]; registered {
		return strcmd.Call(f, splitted[1:])
	} else {
		return fmt.Errorf("function %s not found", splitted[0])
	}
}

// Call calls the function with the provided arguments automatically parsed.
// Supported types of arguments can be extended by adding parser func to [Parsers].
func (strcmd *StrCmd) Call(function any, args []string) (err error) {
    defer func () {
        panicValue := recover()
        if panicValue != nil {
            err = fmt.Errorf("Call: recovered: %v", panicValue)
        }
    } ()

	fType := reflect.TypeOf(function)
	if fType.Kind() != reflect.Func {
		return fmt.Errorf("provided function is actually not a func! (%s)", fType.Kind())
	}

	numIn := fType.NumIn()
	if len(args) != numIn {
		return fmt.Errorf("function argument count mismatch: %d/%d", len(args), numIn)
	}

	preparedArgs := make([]reflect.Value, 0, numIn)
	for i := 0; i < numIn; i++ {
		if parser, hasParser := strcmd.parsers[fType.In(i).Name()]; !hasParser {
			return fmt.Errorf("no parser available for arg#%d (%s)", i, fType.In(i).Name())
		} else {
			parsed, pErr := parser(args[i])
			if pErr != nil {
				return fmt.Errorf("an error occured when parsing arg#%d: %s", i, args[i])
			}
			preparedArgs = append(preparedArgs, reflect.ValueOf(parsed))
		}
	}

	rvs := reflect.ValueOf(function).Call(preparedArgs)
	for _, rv := range rvs {
		if rv.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) && !rv.IsNil() {
			return rv.Interface().(error)
		}
	}
	return nil
}

func (strcmd *StrCmd) split(cmd string) ([]string, error) {
	cmd = strings.TrimSpace(cmd)

	var splitted []string = make([]string, 0)

	left := 0
    expectingSpace := false

	for left < len(cmd) {
        if expectingSpace && cmd[left] != ' ' {
            return nil, fmt.Errorf("malformed cmd, epxecting space")
        }

		if cmd[left] == '\'' || cmd[left] == '"'{
			seg, right, err := strcmd.findEnclosingSeg(cmd[left:])
			if err != nil {
				return splitted, err
			}
			splitted = append(splitted, seg[1: len(seg) - 1])
			left += right
            expectingSpace = true
		} else if cmd[left] == ' ' {
            expectingSpace = false
			left++
			continue
		} else {
			right := findSegEnd(cmd[left:])
			right += left
			splitted = append(splitted, cmd[left:right])
			left = right + 1
		}
	}

	return splitted, nil
}

func (strcmd *StrCmd) findEnclosingSeg(cmd string) (seg string, offset int, err error) {
    if !strings.ContainsRune(cmd, '\\') {
        border := ' '
        prev := ' '
        for i, c := range cmd {
            if i == 0 {
                border = c
                continue
            }
		    if prev == border {
				return cmd[:i], i, nil
            }
            prev = c
        }
    }

    // \ will only be written when it's escaped

    r := strings.NewReader(cmd)
    strcmd.strBuilder.Reset()

    border, read, rErr := r.ReadRune()
    if rErr != nil {
        return "", 0, rErr
    }
    tRead := read
    c := border
    n := ' '

    _, bErr := strcmd.strBuilder.WriteRune(border)
    if bErr != nil {
        return "", 0, bErr
    }
    n, read, rErr = r.ReadRune()
    if rErr != nil {
        return "", 0, rErr
    }
    tRead += read

    escaping := false
    for rErr == nil {
        c = n
        n, read, rErr = r.ReadRune()

        // if rErr == nil && c == '\\' && n != '\\' && n != border { // c is \
        //     _, bErr = strcmd.strBuilder.WriteRune(c)
        //     if bErr != nil {
        //         return strcmd.strBuilder.String(), tRead, bErr
        //     }
        //     return strcmd.strBuilder.String(), tRead, fmt.Errorf("malformed cmd, the next rune should not be escaped: %c", n)
        // }

        if c == border && !escaping { // it's border rune and not escaped
            _, bErr = strcmd.strBuilder.WriteRune(c)
            if bErr != nil {
                return strcmd.strBuilder.String(), tRead, bErr
            }
            return strcmd.strBuilder.String(), tRead, nil
        } else if c != '\\'  { // c is not \
            if escaping && c != border {
                return strcmd.strBuilder.String(), tRead, fmt.Errorf("malformed cmd, this rune should not be escaped: %c", c)
            }
            _, bErr = strcmd.strBuilder.WriteRune(c)
            if bErr != nil {
                return strcmd.strBuilder.String(), tRead, bErr
            }
            escaping = false
        } else if escaping { // c is a escaped \
            _, bErr = strcmd.strBuilder.WriteRune(c)
            if bErr != nil {
                return strcmd.strBuilder.String(), tRead, bErr
            }
            escaping = false
        } else {
            escaping = true
        }
        tRead += read
    }

	return strcmd.strBuilder.String(), tRead, fmt.Errorf("border rune %c not found in \"%s\"", border, cmd)
}

func findSegEnd(cmd string) int {
	for i, c := range cmd {
		if c == ' ' {
			return i
		}
	}

	return len(cmd)
}
