package esxi

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
)

func ParseTemplate(text string, data any) (string, error) {
	// Check if we need to parse text
	re := regexp.MustCompile(`{{.*}}`)
	matches := re.FindAllString(text, -1)
	if len(matches) == 0 {
		return text, nil
	}

	funcMap := template.FuncMap{
		"upper":        strings.ToUpper,
		"lower":        strings.ToLower,
		"trim":         strings.TrimSpace,
		"len":          func(s string) int { return len(s) },
		"substr":       func(s string, start, length int) string { return s[start : start+length] },
		"replace":      strings.Replace,
		"printf":       fmt.Sprintf,
		"add":          func(n, add int) int { return n + add },
		"formatAsDate": func(t time.Time, layout string) string { return t.Format(layout) },
		"now":          time.Now,
		"base64encode": func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) },
		"base64gzip": func(s string) (string, error) {
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)
			if _, err := zw.Write([]byte(s)); err != nil {
				return "", err
			}
			if err := zw.Close(); err != nil {
				return "", err
			}
			return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
		},
		"parsedTemplateOutput": func(parsedOutput string) string { return parsedOutput },
	}

	// Check if parsedTemplate is at the end of the template text
	re = regexp.MustCompile(`{{\s*parsedTemplateOutput\s*.*}}`)
	matches = re.FindAllString(text, -1)

	if len(matches) > 1 {
		return text, fmt.Errorf("parsedTemplateOutput should be present only once")
	}

	if len(matches) > 0 {
		// Get text after the last match
		lastMatch := matches[len(matches)-1]
		lastMatchIndex := strings.LastIndex(text, lastMatch)
		remainingText := strings.TrimSpace(text[lastMatchIndex+len(lastMatch):])
		if remainingText != "" {
			return text, fmt.Errorf("parsedTemplateOutput should be at the bottom of the template")
		}
	}

	parsedTextOutputTpl := ""

	if len(matches) == 1 {
		parsedTextOutputTpl = re.FindString(text)
		text = re.ReplaceAllString(text, "")
	}

	tmpl, err := template.New("text-functions").Funcs(funcMap).Parse(text)
	if err != nil {
		return text, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return text, err
	}

	output := buf.String()

	if parsedTextOutputTpl == "" {
		return output, nil
	}

	tmpl, err = template.New("text-final-functions").Funcs(funcMap).Parse(parsedTextOutputTpl)
	if err != nil {
		return text, err
	}

	var parsedBuf bytes.Buffer
	err = tmpl.Execute(&parsedBuf, output)
	if err != nil {
		return text, err
	}

	output = parsedBuf.String()

	return output, nil
}

func CloseFile(file *os.File) {
	if e := file.Close(); e != nil {
		logging.V(logLevel).Info(e)
	}
}

func RemoveFile(file *os.File) {
	if e := os.Remove(file.Name()); e != nil {
		logging.V(logLevel).Info(e)
	}
}

// Contains checks if an item is present in a collection
func Contains[T comparable](collection []T, value T) bool {
	for _, item := range collection {
		if item == value {
			return true
		}
	}
	return false
}

// ContainsValue checks if an item property value is present in a collection
func ContainsValue[V comparable, T any](collection []T, selector func(T) V, value V) bool {
	for _, item := range collection {
		if selector(item) == value {
			return true
		}
	}
	return false
}

func structToMap(dataStruct interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	value := reflect.ValueOf(dataStruct)
	typeOfStruct := value.Type()

	if typeOfStruct.Kind() == reflect.Ptr {
		value = value.Elem()
		typeOfStruct = typeOfStruct.Elem()
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := typeOfStruct.Field(i)

		// Convert the first letter of the field name to lowercase
		const lowercaseBits = 32
		key := string(fieldType.Name[0]+lowercaseBits) + fieldType.Name[1:]

		switch field.Kind() {
		case reflect.Struct:
			result[key] = structToMap(field.Interface())
		case reflect.Array, reflect.Slice:
			if field.Len() > 0 {
				slice := make([]interface{}, field.Len())
				for j := 0; j < field.Len(); j++ {
					slice[j] = structToMap(field.Index(j).Interface())
				}
				result[key] = slice
			}
		case reflect.Invalid:
			// Handle reflect.Invalid case
			result[key] = nil
		case reflect.Bool:
			result[key] = field.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			result[key] = field.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			result[key] = field.Uint()
		case reflect.Float32, reflect.Float64:
			result[key] = field.Float()
		case reflect.Complex64, reflect.Complex128:
			result[key] = field.Complex()
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.String, reflect.UnsafePointer:
			// Handle other common cases
			result[key] = field.Interface()
		default:
			// Handle the rest of the cases (unlikely to occur in practice)
			result[key] = field.Interface()
		}
	}

	return result
}
