package pmx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v4"
)

var ErrInvalidRef = errors.New("invalid ref")

type Executor interface {
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
}

func IsZero(entity interface{}) bool {
	if entity == nil {
		return true
	}
	v := reflect.ValueOf(entity)
	if reflect.TypeOf(entity).Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.IsZero()
}

func Insert(ctx context.Context, e Executor, entity interface{}) error {
	t := reflect.TypeOf(entity)
	v := reflect.ValueOf(entity)

	if t.Kind() != reflect.Ptr {
		return ErrInvalidRef
	}

	t = t.Elem()
	v = v.Elem()

	if t.Kind() != reflect.Struct {
		return ErrInvalidRef
	}

	buf := bytes.NewBufferString(
		fmt.Sprintf(
			"insert into %s ",
			t.Field(0).Tag.Get("table"),
		),
	)

	columns := []string{}
	marks := []string{}
	args := []interface{}{}

	for i := 0; i < t.NumField(); i++ {
		column := t.Field(i).Tag.Get("db")
		if t.Field(i).Tag.Get("generated") == "always" {
			continue
		}
		if !v.Field(i).CanInterface() {
			continue
		}
		if len(column) == 0 {
			continue
		}
		columns = append(columns, column)
		args = append(args, v.Field(i).Interface())
	}

	for i := 1; i <= len(columns); i++ {
		marks = append(marks, fmt.Sprintf("$%d", i))
	}

	buf.WriteString(
		fmt.Sprintf(
			"(%s) values (%s) returning *",
			strings.Join(columns, ", "),
			strings.Join(marks, ", "),
		),
	)

	rows, err := e.Query(ctx, buf.String(), args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scan(rows, entity)
}

func Update(ctx context.Context, e Executor, entity interface{}, fields []string) error {
	t := reflect.TypeOf(entity)
	v := reflect.ValueOf(entity)

	if t.Kind() != reflect.Ptr {
		return ErrInvalidRef
	}

	t = t.Elem()
	v = v.Elem()

	if t.Kind() != reflect.Struct {
		return ErrInvalidRef
	}

	buf := bytes.NewBufferString(
		fmt.Sprintf(
			"update %s set ",
			t.Field(0).Tag.Get("table"),
		),
	)

	columns := []string{}
	statements := []string{}
	args := []interface{}{}
	allowed := map[string]bool{}

	for _, field := range fields {
		allowed[field] = true
	}

	for i := 0; i < t.NumField(); i++ {
		column := t.Field(i).Tag.Get("db")
		if t.Field(i).Tag.Get("generated") == "always" {
			continue
		}
		if !v.Field(i).CanInterface() {
			continue
		}
		if len(column) == 0 {
			continue
		}
		if i == 0 {
			continue
		}
		if !allowed[t.Field(i).Name] {
			continue
		}
		columns = append(columns, column)
		args = append(args, v.Field(i).Interface())
	}

	for i, column := range columns {
		statements = append(statements,
			fmt.Sprintf(
				"%s = $%d",
				column, i+1,
			),
		)
	}

	buf.WriteString(strings.Join(statements, ", "))
	args = append(args, v.Field(0).Interface())

	buf.WriteString(
		fmt.Sprintf(
			" where %s = $%d returning *",
			t.Field(0).Tag.Get("db"),
			len(args),
		),
	)

	rows, err := e.Query(ctx, buf.String(), args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scan(rows, entity)
}

func Select(ctx context.Context, e Executor, dest interface{}, sql string, args ...interface{}) error {
	rows, err := e.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scan(rows, dest)
}

func scan(rows pgx.Rows, dest interface{}) error {
	t := reflect.TypeOf(dest)
	if t.Kind() != reflect.Ptr {
		return ErrInvalidRef
	}

	t = t.Elem()
	v := reflect.ValueOf(dest)

	switch t.Kind() {
	case reflect.Slice:
		return scanSlice(rows, t, v)
	case reflect.Struct:
		return scanStruct(rows, t, v)
	default:
		return ErrInvalidRef
	}
}

func scanSlice(rows pgx.Rows, t reflect.Type, v reflect.Value) error {
	t = t.Elem()
	if t.Kind() != reflect.Ptr {
		return ErrInvalidRef
	}

	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return ErrInvalidRef
	}

	for rows.Next() {
		ptr, err := scanFields(rows, t)
		if err != nil {
			return err
		}
		sv := v.Elem()
		sv.Set(reflect.Append(sv, ptr))
	}

	return nil
}

func scanStruct(rows pgx.Rows, t reflect.Type, v reflect.Value) error {
	if !rows.Next() {
		return nil
	}

	ptr, err := scanFields(rows, t)
	if err != nil {
		return err
	}

	v.Elem().Set(ptr.Elem())
	return nil
}

func scanFields(rows pgx.Rows, t reflect.Type) (reflect.Value, error) {
	fields := []interface{}{}
	ptr := reflect.New(t)
	v := ptr.Elem()

	for _, fd := range rows.FieldDescriptions() {
		var field interface{}
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).Tag.Get("db") != string(fd.Name) {
				continue
			}
			field = v.Field(i).Addr().Interface()
		}
		fields = append(fields, field)
	}

	for i := range fields {
		if len(rows.RawValues()[i]) == 0 {
			fields[i] = new(interface{})
		}
	}

	err := rows.Scan(fields...)
	if err != nil {
		return reflect.ValueOf(nil), err
	}

	return ptr, nil
}
