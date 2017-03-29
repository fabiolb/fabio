package verify

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Values verifies that got has all the content, and only the content, defined by want.
func Values(t *testing.T, name string, got, want interface{}) (ok bool) {
	tr := travel{}
	tr.values(reflect.ValueOf(got), reflect.ValueOf(want), nil)

	fail := tr.report(name)
	if fail != "" {
		t.Error(fail)
		return false
	}

	return true
}

func (t *travel) values(got, want reflect.Value, path []*segment) {
	if !want.IsValid() {
		if got.IsValid() {
			t.differ(path, "Unwanted %s", got.Type())
		}
		return
	}
	if !got.IsValid() {
		t.differ(path, "Missing %s", want.Type())
		return
	}

	if got.Type() != want.Type() {
		t.differ(path, "Got type %s, want %s", got.Type(), want.Type())
		return
	}

	switch got.Kind() {

	case reflect.Struct:
		seg := &segment{format: "/%s"}
		path = append(path, seg)
		for i, n := 0, got.NumField(); i < n; i++ {
			seg.x = got.Type().Field(i).Name
			t.values(got.Field(i), want.Field(i), path)
		}
		path = path[:len(path)-1]

	case reflect.Slice, reflect.Array:
		n := got.Len()
		if n != want.Len() {
			t.differ(path, "Got %d elements, want %d", n, want.Len())
			return
		}

		seg := &segment{format: "[%d]"}
		path = append(path, seg)
		for i := 0; i < n; i++ {
			seg.x = i
			t.values(got.Index(i), want.Index(i), path)
		}
		path = path[:len(path)-1]

	case reflect.Ptr, reflect.Interface:
		t.values(got.Elem(), want.Elem(), path)

	case reflect.Map:
		seg := &segment{}
		path = append(path, seg)
		for _, key := range want.MapKeys() {
			applyKeySeg(seg, key)
			t.values(got.MapIndex(key), want.MapIndex(key), path)
		}

		for _, key := range got.MapKeys() {
			v := want.MapIndex(key)
			if v.IsValid() {
				continue
			}
			applyKeySeg(seg, key)
			t.values(got.MapIndex(key), v, path)
		}
		path = path[:len(path)-1]

	case reflect.Func:
		t.differ(path, "Can't compare functions")

	default:
		a, b := asInterface(got), asInterface(want)
		if a != b {
			t.differ(path, differMsg(a, b))
		}

	}
}

func applyKeySeg(dst *segment, key reflect.Value) {
	if key.Kind() == reflect.String {
		dst.format = "[%q]"
	} else {
		dst.format = "[%v]"
	}
	dst.x = asInterface(key)
}

func asInterface(v reflect.Value) interface{} {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Complex64, reflect.Complex128:
		return v.Complex()
	case reflect.String:
		return v.String()
	default:
		return v.Interface()
	}
}

func differMsg(got, want interface{}) string {
	switch got.(type) {
	case int64:
		g, w := got.(int64), want.(int64)
		if g < 0xA && g > -0xA && w < 0xA && w > -0xA {
			return fmt.Sprintf("Got %d, want %d", got, want)
		}
		return fmt.Sprintf("Got %d (0x%x), want %d (0x%x)", got, got, want, want)
	case uint64:
		g, w := got.(uint64), want.(uint64)
		if g < 0xA && w < 0xA {
			return fmt.Sprintf("Got %d, want %d", got, want)
		}
		return fmt.Sprintf("Got %d (0x%x), want %d (0x%x)", got, got, want, want)
	case float64, complex128:
		return fmt.Sprintf("Got %f (%e), want %f (%e)", got, got, want, want)
	case string:
		a, b := got.(string), want.(string)
		if len(a) > len(b) {
			a, b = b, a
		}
		r := strings.NewReader(b)

		var differAt int
		for i, c := range a {
			o, _, _ := r.ReadRune()
			if c != o {
				differAt = i
				break
			}
		}

		format := "Got %q, want %q"
		if differAt > 0 {
			format += fmt.Sprintf("\n     %s^", strings.Repeat(" ", differAt))
		}
		return fmt.Sprintf(format, got, want)
	default:
		return fmt.Sprintf("Got %v, want %v", got, want)
	}
}
