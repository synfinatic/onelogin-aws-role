package utils

/*
 * I find a lot of reasons to print pretty tables for users.  This
 * code attempts to be a general purpose solution for that.
 */

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
)

const TABLE_HEADER_TAG = "header"

type TableStruct interface {
	GetHeader(string) (string, error)
}

// Returns a row and a mapping of struct field name to header names
func TableRow(table TableStruct) (map[string]string, map[string]string) {
	row := map[string]string{}
	tbl := reflect.ValueOf(table)
	fieldCnt := tbl.Type().NumField()
	headers := make(map[string]string, fieldCnt)

	for i := 0; i < fieldCnt; i++ {
		f := reflect.TypeOf(table).Field(i)
		header, err := table.GetHeader(f.Name)
		if err != nil {
			log.Fatal(err)
		}
		fval := tbl.FieldByName(f.Name)
		headers[f.Name] = header
		if !fval.IsValid() {
			log.Error(fval)
			continue // wtf?  this shouldn't happen!
		}
		switch fval.Kind() {
		case reflect.String:
			row[f.Name] = fval.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			row[f.Name] = fmt.Sprintf("%d", fval.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			row[f.Name] = fmt.Sprintf("%d", fval.Uint())
		case reflect.Bool:
			if fval.Bool() {
				row[f.Name] = "true"
			} else {
				row[f.Name] = "false"
			}
		default:
			log.Fatalf("Unsupported type: %s", f.Type.Kind())
			row[f.Name] = ""
		}
	}
	return row, headers
}

// Geneates a table using a list of TableStruct & struct field names in the report
func GenerateTable(tables []TableStruct, fields []string) {
	table := []map[string]string{}
	headers := map[string]string{}
	for _, item := range tables {
		row, h := TableRow(item)
		table = append(table, row)
		headers = h
	}

	generateTable(table, headers, fields)
}

func generateTable(data []map[string]string, fieldMap map[string]string, fields []string) {
	table := [][]string{}
	colWidth := make([]int, len(fields))

	// figure out width of column headers
	for i, field := range fields {
		colWidth[i] = len(fieldMap[field])
	}

	// calc max len of every column & build our row
	for _, r := range data {
		row := make([]string, len(fields))
		for i, field := range fields {
			row[i] = r[field]
			if len(r[field]) > colWidth[i] {
				colWidth[i] = len(r[field])
			}
		}
		table = append(table, row)
	}

	// build our fstring for each row
	fstrings := make([]string, len(fields))
	for i, width := range colWidth {
		fstrings[i] = fmt.Sprintf("%%-%ds", width)
	}
	fstring := strings.Join(fstrings, " | ")
	fstring = fmt.Sprintf("%s\n", fstring)

	// fmt.Sprintf() expects []interface...
	finter := make([]interface{}, len(fields))
	for i, field := range fields {
		finter[i] = fieldMap[field]
	}

	// print the header
	headerLine := fmt.Sprintf(fstring, finter...)
	fmt.Printf("%s%s\n", headerLine, strings.Repeat("=", len(headerLine)-1))

	// print each row
	for _, row := range data {
		values := make([]interface{}, len(fields))
		for i, field := range fields {
			values[i] = row[field]
		}
		fmt.Printf(fstring, values...)
	}
}

func GetHeaderTag(v reflect.Value, fieldName string) (string, error) {
	field, ok := v.Type().FieldByName(fieldName)
	if !ok {
		return "", fmt.Errorf("Invalid field '%s' in %s", fieldName, v.Type().Name())
	}
	tag := string(field.Tag.Get(TABLE_HEADER_TAG))
	return tag, nil
}
