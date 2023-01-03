package pmx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrInvalidRef = errors.New("invalid ref")

type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
}

type Selector interface {
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
}

type UpdateOptions struct {
	Set []string
	By  []string
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
		if t.Field(i).Tag.Get("generated") == "auto" {
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
			"(%s) values (%s)",
			strings.Join(columns, ", "),
			strings.Join(marks, ", "),
		),
	)

	_, err := e.Exec(ctx, buf.String(), args...)
	if err != nil {
		return err
	}

	return nil
}

func Update(ctx context.Context, e Executor, entity interface{}, options *UpdateOptions) error {
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
	denied := map[string]bool{}

	for _, field := range options.Set {
		allowed[field] = true
	}

	for _, field := range options.By {
		denied[field] = true
	}

	for i := 0; i < t.NumField(); i++ {
		column := t.Field(i).Tag.Get("db")
		if t.Field(i).Tag.Get("generated") == "auto" {
			continue
		}
		if !v.Field(i).CanInterface() {
			continue
		}
		if len(column) == 0 {
			continue
		}
		if !allowed[t.Field(i).Name] {
			continue
		}
		if denied[t.Field(i).Name] {
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

	conditions := []string{}
	for _, field := range options.By {
		sf, ok := t.FieldByName(field)
		column := sf.Tag.Get("db")

		if !ok {
			return fmt.Errorf("struct field not found: %s", field)
		}
		if len(column) == 0 {
			return fmt.Errorf("struct field must be annotated: %s", field)
		}
		if !v.FieldByName(field).CanInterface() {
			return fmt.Errorf("struct field must be exported: %s", field)
		}

		args = append(args, v.FieldByName(field).Interface())
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	buf.WriteString(
		fmt.Sprintf(
			" where %s",
			strings.Join(conditions, " and "),
		),
	)

	_, err := e.Exec(ctx, buf.String(), args...)
	if err != nil {
		return err
	}

	return nil
}

func Select(ctx context.Context, s Selector, dest interface{}, sql string, args ...interface{}) (bool, error) {
	rows, err := s.Query(ctx, sql, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return scan(rows, dest)
}

func scan(rows pgx.Rows, dest interface{}) (bool, error) {
	t := reflect.TypeOf(dest)
	if t.Kind() != reflect.Ptr {
		return false, ErrInvalidRef
	}

	t = t.Elem()
	v := reflect.ValueOf(dest)

	switch t.Kind() {
	case reflect.Slice:
		return scanSlice(rows, t, v)
	case reflect.Struct:
		return scanStruct(rows, t, v)
	default:
		return false, ErrInvalidRef
	}
}

func scanSlice(rows pgx.Rows, t reflect.Type, v reflect.Value) (bool, error) {
	t = t.Elem()
	if t.Kind() != reflect.Ptr {
		return false, ErrInvalidRef
	}

	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return false, ErrInvalidRef
	}

	ok := false

	for rows.Next() {
		ptr, err := scanFields(rows, t)
		if err != nil {
			return false, err
		}
		sv := v.Elem()
		sv.Set(reflect.Append(sv, ptr))
		ok = true
	}

	return ok, nil
}

func scanStruct(rows pgx.Rows, t reflect.Type, v reflect.Value) (bool, error) {
	if !rows.Next() {
		return false, nil
	}

	ptr, err := scanFields(rows, t)
	if err != nil {
		return false, err
	}

	v.Elem().Set(ptr.Elem())
	return true, nil
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
