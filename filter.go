package filter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/qri-io/dataset/vals"
)

// Apply executes a filter string against a given source, returning a filtered result
func Apply(filterStr string, source interface{}) (val interface{}, err error) {
	// fmt.Printf("parse %s\n", filterStr)
	r := strings.NewReader(filterStr)
	p := parser{s: newScanner(r)}
	filters, err := p.filters()
	if err != nil {
		return nil, err
	}

	val = source
	for _, f := range filters {
		// fmt.Printf("run filter: %#v\n", f)
		// TODO - resolve links here
		val, err = f.apply(val)
		// fmt.Printf("result: %#v\n", val)
		if err != nil {
			return val, err
		}
	}

	return unpackIterators(val)
}

func unpackIterators(in value) (val interface{}, err error) {
	if it, ok := in.(vals.Iterator); ok {
		vs := []interface{}{}
		defer it.Done()
		for {
			ent, done := it.Next()
			if done {
				return vs, nil
			}
			if val, err = unpackIterators(ent.Value); err != nil {
				return nil, err
			}
			vs = append(vs, val)
		}
	}

	return in, nil
}

type filter interface {
	apply(in value) (out value, err error)
}

type value interface {
}

type fStringLiteral string

func (f fStringLiteral) apply(in value) (out value, err error) {
	return string(f), nil
}

type fNumericLiteral float64

func (f fNumericLiteral) apply(in value) (out value, err error) {
	return f, nil
}

type fLength byte

func (f fLength) apply(in value) (out value, err error) {
	if in == nil {
		return 0, nil
	}
	target := reflect.ValueOf(in)
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}
	switch target.Kind() {
	case reflect.Struct:
		return target.NumField(), nil
	default:
		return target.Len(), nil
	}
}

type selector interface {
	filter
	isSelector()
}

type fSelector []selector

func (f fSelector) apply(in value) (out value, err error) {
	out = in
	for _, sel := range f {
		out, err = sel.apply(out)
		if err != nil {
			return out, err
		}
	}
	return out, err
}

// fIdentity is the identity filter, it returns whatever it's given
type fIdentity byte

func (f fIdentity) isSelector() {}

func (f fIdentity) apply(in value) (out value, err error) {
	return in, nil
}

type fKeySelector string

func (f fKeySelector) isSelector() {}

func (f fKeySelector) apply(in value) (out value, err error) {
	// if keyable, ok := in.(vals.Keyable); ok {
	// 	return keyable.ValueForKey(string(f))
	// }

	if it, ok := in.(vals.Iterator); ok {
		vals := []interface{}{}
		for {
			e, done := it.Next()
			if done {
				return vals, nil
			}

			val, err := f.apply(e.Value)
			if err != nil {
				return nil, err
			}
			vals = append(vals, val)
		}
	}

	target := reflect.ValueOf(in)
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}

	if target.Kind() == reflect.Slice {
		vals := []interface{}{}
		l := target.Len()
		for i := 0; i < l; i++ {
			v, err := f.applySingle(target.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			vals = append(vals, v)
		}
		// fmt.Printf("returning array application: %v\n", vals)
		return vals, nil
	}

	return f.applySingle(in)
}

func (f fKeySelector) applySingle(in value) (out value, err error) {
	if in == nil {
		return nil, nil
	}

	if keyable, ok := in.(vals.Keyable); ok {
		return keyable.ValueForKey(string(f))
	}

	// fmt.Printf("key selector input: %#v\n", in)
	t := reflect.ValueOf(in)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	} else if t.Kind() == reflect.Interface {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return nil, nil
	case reflect.Struct:
		return f.selectStructField(t)
	case reflect.Map:
		val := t.MapIndex(reflect.ValueOf(string(f)))
		return val.Interface(), nil
	}

	return nil, fmt.Errorf("unexpected key selector value: %#v", in)
}

func (f fKeySelector) selectStructField(target reflect.Value) (out value, err error) {
	str := string(f)
	for i := 0; i < target.NumField(); i++ {
		// Lowercase the key in order to make matching case-insensitive.
		fieldName := target.Type().Field(i).Name
		// lowerName := strings.ToLower(fieldName)

		fieldTag := target.Type().Field(i).Tag
		if fieldTag != "" && fieldTag.Get("json") != "" {
			jsonName := fieldTag.Get("json")
			pos := strings.Index(jsonName, ",")
			if pos != -1 {
				jsonName = jsonName[:pos]
			}
			// lowerName = strings.ToLower(jsonName)
			fieldName = jsonName
		}

		if fieldName == str {
			return target.Field(i).Interface(), nil
		}
	}

	// TODO (b5) - is not finding a key an error?
	return nil, nil
}

type fIndexSelector int

func (f fIndexSelector) isSelector() {}

func (f fIndexSelector) apply(in value) (out value, err error) {
	if it, ok := in.(*iterator); ok {
		vals := []interface{}{}
		for {
			e, done := it.Next()
			if done {
				return vals, nil
			}
			v, err := f.applySingle(e.Value)
			if err != nil {
				return nil, err
			}
			vals = append(vals, v)
		}
	}

	// target := reflect.ValueOf(in)
	// if target.Kind() == reflect.Ptr {
	// 	target = target.Elem()
	// }

	// if target.Kind() == reflect.Slice {
	// 	vals := []interface{}{}
	// 	l := target.Len()
	// 	for i := 0; i < l; i++ {
	// 		v, err := f.applySingle(target.Index(i))
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		vals = append(vals, v)
	// 	}
	// 	// fmt.Printf("returning array application: %v\n", vals)
	// 	return vals, nil
	// }

	return f.applySingle(in)
}

func (f fIndexSelector) applySingle(in value) (out value, err error) {
	if indexer, ok := in.(vals.Indexable); ok {
		return indexer.ValueForIndex(int(f))
	}

	target := reflect.ValueOf(in)
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}

	if target.Kind() == reflect.Interface {
		target = target.Elem()
	}

	switch target.Kind() {
	case reflect.Slice:
		return target.Index(int(f)).Interface(), nil
	}

	return nil, fmt.Errorf("select index of non array type %#v", target)
}

type fIndexRangeSelector struct {
	start int
	stop  int
	all   bool
}

func (f *fIndexRangeSelector) isSelector() {}

func (f *fIndexRangeSelector) apply(in value) (out value, err error) {
	if it, ok := in.(vals.Iterator); ok {
		// TODO (b5) - can't use this trick b/c the internal iterater implementation
		// is being used as a signal to the next filter that it needs to iterate
		// should probs find another way to pass this message along
		// if f.all {
		// 	return it, nil
		// }

		vals := []interface{}{}
		for {
			e, done := it.Next()
			if done {
				return &iterator{v: reflect.ValueOf(vals)}, nil
			}
			if e.Index < f.start {
				continue
			}
			if e.Index == f.stop && !f.all {
				return &iterator{v: reflect.ValueOf(vals)}, nil
			}

			vals = append(vals, e.Value)
		}
	}

	target := reflect.ValueOf(in)
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}

	switch target.Kind() {
	case reflect.Slice:
		if f.all {
			f.stop = target.Len()
		}
		return &iterator{v: target.Slice(f.start, f.stop)}, nil
	}

	return nil, fmt.Errorf("unexpected range value: %#v", in)
}

type iterator struct {
	i int
	v reflect.Value
}

func (it *iterator) Next() (e *vals.Entry, done bool) {
	defer func() { it.i++ }()
	// fmt.Println(it.i)
	if it.i == it.v.Len() {
		return nil, true
	}
	return &vals.Entry{Index: it.i, Value: it.v.Index(it.i).Interface()}, false
}

func (it *iterator) Done() {}

func (it *iterator) ValueForIndex(i int) (v interface{}, err error) {
	return it.v.Index(i).Interface(), nil
}

type fBinaryOp struct {
	left  filter
	op    tokenType
	right filter
}

func (f fBinaryOp) apply(in value) (out value, err error) {
	left, err := f.left.apply(in)
	if err != nil {
		return nil, err
	}
	left, lk := normalizeValue(left)

	right, err := f.right.apply(in)
	if err != nil {
		return nil, err
	}
	right, rk := normalizeValue(right)

	switch f.op {
	case tStar:
		if lk == reflect.Float64 && rk == reflect.Float64 {
			return left.(float64) * right.(float64), nil
		}
	case tPlus:
		if lk == reflect.Float64 && rk == reflect.Float64 {
			return left.(float64) + right.(float64), nil
		}
	}

	fmt.Printf("binary operations are not finished cannot %v %#v %s %v %#v\n", lk, left, f.op, rk, right)
	return nil, fmt.Errorf("binary operations are not finished cannot %#v %s %#v", left, f.op, right)
}

func normalizeValue(in value) (out value, rk reflect.Kind) {
	if nl, ok := in.(fNumericLiteral); ok {
		return float64(nl), reflect.Float64
	} else if sl, ok := in.(fStringLiteral); ok {
		return string(sl), reflect.String
	}

	rk = reflect.TypeOf(in).Kind()
	switch rk {
	case reflect.Int:
		return float64(in.(int)), reflect.Float64
	case reflect.Float64:
		return in, rk
	}

	return in, rk
}

type fSlice []filter

func (fs fSlice) apply(in value) (out value, err error) {
	vals := make([]interface{}, len(fs))
	for i, f := range fs {
		if vals[i], err = f.apply(in); err != nil {
			return nil, err
		}
	}
	return vals, nil
}
