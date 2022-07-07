package xmlrpc

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"time"
)

// Encoder implementations are responsible for handling encoding of XML-RPC requests to the proper wire format.
type Encoder interface {
	Encode(w io.Writer, methodName string, args interface{}) error
}

// StdEncoder is the default implementation of Encoder interface.
type StdEncoder struct{}

func (e *StdEncoder) Encode(w io.Writer, methodName string, args interface{}) error {
	_, _ = fmt.Fprintf(w, "<methodCall><methodName>%s</methodName>", methodName)

	if args != nil {
		if err := e.encodeArgs(w, args); err != nil {
			return fmt.Errorf("cannot encoded provided method arguments: %w", err)
		}
	}

	_, _ = fmt.Fprint(w, "</methodCall>")

	return nil
}

func (e *StdEncoder) encodeArgs(w io.Writer, args interface{}) error {

	// Allows reading both pointer and value-structs
	elem := reflect.Indirect(reflect.ValueOf(args))
	numFields := elem.NumField()
	if numFields > 0 {

		hasExportedFields := false
		for fN := 0; fN < numFields; fN++ {

			field := elem.Field(fN)
			if !field.CanInterface() {
				continue
			}

			// If this is first exported field - print out <params> tag
			if !hasExportedFields {
				hasExportedFields = true
				_, _ = fmt.Fprint(w, "<params>")
			}

			_, _ = fmt.Fprint(w, "<param>")
			if err := e.encodeValue(w, field.Interface()); err != nil {
				return fmt.Errorf("cannot encode argument '%s': %w", elem.Type().Field(fN).Name, err)
			}
			_, _ = fmt.Fprint(w, "</param>")
		}

		// Only write closing </params> tag if at least one exported field is found
		if hasExportedFields {
			_, _ = fmt.Fprint(w, "</params>")
		}
	}

	return nil
}

// encodeValue will encode input into the XML-RPC compatible format.
// If provided value is a pointer, value of pointer will be used, unless pointer is nil.
// In that case a <nil/> value is returned.
//
// See more: https://en.wikipedia.org/wiki/XML-RPC#Data_types
func (e *StdEncoder) encodeValue(w io.Writer, value interface{}) error {

	valueOf := reflect.ValueOf(value)
	kind := valueOf.Kind()

	// Handling pointers by following them.
	if kind == reflect.Ptr {
		if valueOf.IsNil() {
			_, _ = fmt.Fprint(w, "<value><nil/></value>")
			return nil
		}
		return e.encodeValue(w, valueOf.Elem().Interface())
	}

	_, _ = fmt.Fprint(w, "<value>")
	switch kind {

	case reflect.Bool:
		if err := e.encodeBoolean(w, value.(bool)); err != nil {
			return fmt.Errorf("cannot encode boolean value: %w", err)
		}

	case reflect.Int:
		if err := e.encodeInteger(w, value.(int)); err != nil {
			return fmt.Errorf("cannot encode integer value: %w", err)
		}

	case reflect.Float64:
		if err := e.encodeDouble(w, value.(float64)); err != nil {
			return fmt.Errorf("cannot encode double value: %w", err)
		}

	case reflect.String:
		if err := e.encodeString(w, value.(string)); err != nil {
			return fmt.Errorf("cannot encode string value: %w", err)
		}

	case reflect.Array, reflect.Slice:

		if e.isByteArray(value) {
			if err := e.encodeBase64(w, value.([]byte)); err != nil {
				return fmt.Errorf("cannot encode byte-array value: %w", err)
			}
		} else {
			if err := e.encodeArray(w, value); err != nil {
				return fmt.Errorf("cannot encode array value: %w", err)
			}
		}

	case reflect.Struct:
		if reflect.TypeOf(value).String() != "time.Time" {
			if err := e.encodeStruct(w, value); err != nil {
				return fmt.Errorf("cannot encode struct value: %w", err)
			}
		} else {
			if err := e.encodeTime(w, value.(time.Time)); err != nil {
				return fmt.Errorf("cannot encode time.Time value: %w", err)
			}
		}
	}

	_, _ = fmt.Fprint(w, "</value>")
	return nil
}

func (e *StdEncoder) isByteArray(val interface{}) bool {

	_, ok := val.([]byte)
	return ok
}

func (e *StdEncoder) encodeInteger(w io.Writer, val int) error {

	_, err := fmt.Fprintf(w, "<int>%d</int>", val)
	return err
}

func (e *StdEncoder) encodeDouble(w io.Writer, val float64) error {

	_, err := fmt.Fprintf(w, "<double>%f</double>", val)
	return err
}

func (e *StdEncoder) encodeBoolean(w io.Writer, val bool) error {

	v := 0
	if val {
		v = 1
	}
	_, err := fmt.Fprintf(w, "<boolean>%d</boolean>", v)

	return err
}

func (e *StdEncoder) encodeString(w io.Writer, val string) error {

	_, _ = fmt.Fprint(w, "<string>")
	if err := xml.EscapeText(w, []byte(val)); err != nil {
		return fmt.Errorf("failed to escape string: %w", err)
	}
	_, _ = fmt.Fprint(w, "</string>")

	return nil
}

func (e *StdEncoder) encodeArray(w io.Writer, val interface{}) error {

	_, _ = fmt.Fprint(w, "<array><data>")
	for i := 0; i < reflect.ValueOf(val).Len(); i++ {
		if err := e.encodeValue(w, reflect.ValueOf(val).Index(i).Interface()); err != nil {
			return fmt.Errorf("cannot encode array element at index %d: %w", i, err)
		}
	}

	_, _ = fmt.Fprint(w, "</data></array>")

	return nil
}

func (e *StdEncoder) encodeStruct(w io.Writer, val interface{}) error {

	_, _ = fmt.Fprint(w, "<struct>")
	for i := 0; i < reflect.TypeOf(val).NumField(); i++ {

		field := reflect.ValueOf(val).Field(i)
		// Skip over unexported fields
		if !field.CanInterface() {
			continue
		}

		fieldType := reflect.TypeOf(val).Field(i)
		fieldName := fieldType.Name
		if fieldType.Tag.Get("xml") != "" {
			fieldName = fieldType.Tag.Get("xml")
		}
		_, _ = fmt.Fprintf(w, "<member><name>%s</name>", fieldName)

		if err := e.encodeValue(w, field.Interface()); err != nil {
			return fmt.Errorf("cannot encode value of struct field '%s': %w", fieldName, err)
		}
		_, _ = fmt.Fprint(w, "</member>")
	}
	_, _ = fmt.Fprint(w, "</struct>")

	return nil
}

func (e *StdEncoder) encodeBase64(w io.Writer, val []byte) error {

	_, err := fmt.Fprintf(w, "<base64>%s</base64>", base64.StdEncoding.EncodeToString(val))
	return err
}

func (e *StdEncoder) encodeTime(w io.Writer, val time.Time) error {

	_, err := fmt.Fprintf(w, "<dateTime.iso8601>%s</dateTime.iso8601>", val.Format(time.RFC3339))
	return err
}
