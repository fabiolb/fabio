package config

import (
	"reflect"
	"testing"
)

func TestParseKVSlice(t *testing.T) {
	tests := []struct {
		desc string
		s    string
		m    []map[string]string
		err  error
	}{
		{"empty", "", nil, nil},
		{"key=val", "a=b", []map[string]string{{"a": "b"}}, nil},
		{"key with spaces", " a =b", []map[string]string{{"a": "b"}}, nil},
		{"quoted value", "a=\"b\"", []map[string]string{{"a": "b"}}, nil},
		{"single quoted value", "a='b'", []map[string]string{{"a": "b"}}, nil},
		{"quoted value with backslash", `a="b\\\""`, []map[string]string{{"a": `b\"`}}, nil},
		{"ignore empty map front", ",a=b", []map[string]string{{"a": "b"}}, nil},
		{"ignore empty map back", "a=b,", []map[string]string{{"a": "b"}}, nil},
		{"ignore empty value front", ";a=b", []map[string]string{{"a": "b"}}, nil},
		{"ignore empty value back", "a=b;", []map[string]string{{"a": "b"}}, nil},
		{"multiple values", "a=b;c=d", []map[string]string{{"a": "b", "c": "d"}}, nil},
		{"multiple maps", "a=b,c=d", []map[string]string{{"a": "b"}, {"c": "d"}}, nil},
		{"multiple values and maps", "a=b;c=d,e=f;g=h", []map[string]string{{"a": "b", "c": "d"}, {"e": "f", "g": "h"}}, nil},
		{"first key empty", "b", []map[string]string{{"": "b"}}, nil},
		{"first key empty and more values", "b;c=d", []map[string]string{{"": "b", "c": "d"}}, nil},
		{"first key empty and more maps", "b,c", []map[string]string{{"": "b"}, {"": "c"}}, nil},
		{"first key empty and more maps and values", "b;c=d,e;f=g", []map[string]string{{"": "b", "c": "d"}, {"": "e", "f": "g"}}, nil},
		{"issue 305", "a=b=c,d=e=f", []map[string]string{{"a": "b=c"}, {"d": "e=f"}}, nil},
		{"issue 305", "a=b=c;d=e=f", []map[string]string{{"a": "b=c", "d": "e=f"}}, nil},
		{"issue 305", "a=b;d=e=f", []map[string]string{{"a": "b", "d": "e=f"}}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m, err := parseKVSlice(tt.s)
			if got, want := err, tt.err; !reflect.DeepEqual(got, want) {
				t.Fatalf("got error %v want %v", got, want)
			}
			if got, want := m, tt.m; !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v want %#v", got, want)
			}
		})
	}
}
