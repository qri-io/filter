package filter

import (
	"fmt"
	"reflect"
	"strings"
)

// Apply executes a filter string against a given source, returning a filtered result
func Apply(filterStr string, source interface{}) (val interface{}, err error) {
	// fmt.Printf("parse %s\n", filterStr)
	r := strings.NewReader(filterStr)
	p := parser{s: newScanner(r)}
	filters, err := p.read()
	if err != nil {
		return nil, err
	}

	val = source
	for _, f := range filters {
		val, err = f.apply(val)
		if err != nil {
			return val, err
		}
	}

	return val, err
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
	target := reflect.ValueOf(in)
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}

	switch target.Kind() {
	case reflect.Struct:
		return f.selectStructField(target)
	case reflect.Map:
		val := target.MapIndex(reflect.ValueOf(string(f)))
		return val.Interface(), nil
	}

	return nil, fmt.Errorf("field selection by key not finished")
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

type fRangeSelector struct {
	start string
	stop  string
}

func (f fRangeSelector) isSelector() {}

func (f fRangeSelector) apply(in value) (out value, err error) {
	return nil, fmt.Errorf("range selection not finished")
}

type fBinaryOp struct {
	left  filter
	op    token
	right filter
}

func (f fBinaryOp) apply(in value) (out value, err error) {
	return nil, fmt.Errorf("binary operations are not finished")
}
