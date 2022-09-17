package internal

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_split(t *testing.T) {
	strcmd := StrCmd{}
	tests := []struct {
		cmd     string
		want    []string
		wantErr bool
	}{
		{cmd: "test1 12344", want: []string{"test1", "12344"}, wantErr: false},
		{cmd: "test2 12344 \"C:/Windows/A Folder\"", want: []string{"test2", "12344", "C:/Windows/A Folder"}, wantErr: false},
		{cmd: "test3 '12344' 'C:/Windows/A Folder'", want: []string{"test3", "12344", "C:/Windows/A Folder"}, wantErr: false},
		{cmd: "test4 \"C:/Windows/A Folder\\\"", want: []string{"test4"}, wantErr: true},
		{cmd: "test5 \"C:/Windows/A\\Folder\"", want: []string{"test5"}, wantErr: true},
		{cmd: "test6 \"C:/Windows/A\\\\Folder\"", want: []string{"test6", "C:/Windows/A\\Folder"}, wantErr: false},
		{cmd: "test7 \"C:/Windows/A\\\"\\\"Folder\\\\\\\"", want: []string{"test7"}, wantErr: true},
		{cmd: "test8 \"C:/Windows/A Folder\\", want: []string{"test8"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got, err := strcmd.split(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("split() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("split() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_autoCall(t *testing.T) {
	strcmd := StrCmd{}
	tests := []struct {
		cmd       string
		functions map[string]any
		wantErr   bool
	}{
		{
			cmd: "sv.add 123",
			functions: map[string]any{
				"sv.add": func(id int) error {
					if id == 123 {
						return nil
					} else {
						return fmt.Errorf("wrong value: %d", id)
					}
				},
			},
			wantErr: false,
		},
		{
			cmd: "sv.add 123 randomUser412",
			functions: map[string]any{
				"sv.add": func(id int, owner string) error {
					if id == 123 && owner == "randomUser412" {
						return nil
					} else {
						return fmt.Errorf("wrong value: %d %s", id, owner)
					}
				},
			},
			wantErr: false,
		},
		{
			cmd: "test \"123 randomUser412\"",
			functions: map[string]any{
				"test": func(str string) error {
					if str == "123 randomUser412" {
						return nil
					} else {
						return fmt.Errorf("wrong value: %s", str)
					}
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if err := strcmd.CallNamed(tt.cmd, tt.functions); (err != nil) != tt.wantErr {
				t.Errorf("autoCall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStrCmd_findEnclosingSeg(t *testing.T) {
	strcmd := StrCmd{}
	tests := []struct {
		cmd        string
		wantSeg    string
		wantOffset int
		wantErr    bool
	}{
		{cmd: "'\\\\'", wantSeg: "'\\'", wantOffset: 4, wantErr: false},
		{cmd: "'a\\\\b'", wantSeg: "'a\\b'", wantOffset: 6, wantErr: false},
		{cmd: "'a\\'b c' 123", wantSeg: "'a'b c'", wantOffset: 8, wantErr: false},
		{cmd: "'12344' '12345' ", wantSeg: "'12344'", wantOffset: 7, wantErr: false},
		{cmd: "''", wantSeg: "''", wantOffset: 2, wantErr: false},
		{cmd: "\"12 3\"", wantSeg: "\"12 3\"", wantOffset: 6, wantErr: false},
		{cmd: "\"12 3\" \"12 3\"", wantSeg: "\"12 3\"", wantOffset: 6, wantErr: false},
		{cmd: "\" \\b", wantSeg: "\" ", wantOffset: 4, wantErr: true},  // invalid escaping
		{cmd: "\" \\\"", wantSeg: "\" \"", wantOffset: 4, wantErr: true}, // border not found
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			gotSeg, gotOffset, err := strcmd.findEnclosingSeg(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrCmd.findEnclosingSeg() error = %v, wantErr %v, got: %v", err, tt.wantErr, gotSeg)
				return
			}
			if gotSeg != tt.wantSeg {
				t.Errorf("StrCmd.findEnclosingSeg() gotSeg = %v, want %v,  error = %v", gotSeg, tt.wantSeg, err)
			}
			if gotOffset != tt.wantOffset {
				t.Errorf("StrCmd.findEnclosingSeg() gotOffset = %v, want %v,  error = %v", gotOffset, tt.wantOffset, err)
			}
		})
	}
}
