//INI files specific functions

package cfgp

import (
	"bufio"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/eraclitux/stracer"
)

// parseKeyValue given one line encoded like "key = value" returns corresponding
// []string with kv[0] = "key" and > kv[1] = value".
func parseKeyValue(line string) []string {
	// Check for inline comments.
	if strings.Contains(line, ";") {
		line = strings.Split(line, ";")[0]
	} else if strings.Contains(line, "#") {
		line = strings.Split(line, "#")[0]
	}
	line = strings.Replace(line, " ", "", -1)
	// Does nothing if no "=" sign.
	if strings.Contains(line, "=") {
		return strings.Split(line, "=")

	}
	return nil
}

func getFieldByTagName(structValue reflect.Value, name string) reflect.Value {
	field := reflect.Value{}
	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		if n, ok := nameFromTags(structType.Field(i)); ok {
			if n == name {
				stracer.Traceln("found a key by tag", n, name)
				field = structValue.Field(i)
			}
		}
	}
	return field
}

// putInStruct converts values in conf files to Go types.
// It does not return error for field not found.
func putInStruct(structValue reflect.Value, kv []string) error {
	// FIXME add more types.
	stracer.Traceln("storing pair:", kv)
	f := strings.Title(kv[0])
	fieldValue := structValue.FieldByName(f)
	// Field not found, try to get it by tags.
	if !fieldValue.IsValid() {
		fieldValue = getFieldByTagName(structValue, kv[0])
	}
	err := assignType(fieldValue, kv[1])
	if err != nil {
		return err
	}
	return nil
}

// parseINI opens configuration file specified by path and populate
// passed struct.
// Files must follows INI informal standard:
//
//	https://en.wikipedia.org/wiki/INI_file
//
// FIXME Current implementation stores info about section but
// discards it. Use reflection tags to specify which section to use.
func parseINI(path string, structValue reflect.Value) error {
	conf := make(map[string][][]string)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	sectionExp := regexp.MustCompile(`^(\[).+(\])$`)
	commentExp := regexp.MustCompile(`^(#|;)`)
	// Adds default section "default" in case no one is specified
	section := "default"
	for scanner.Scan() {
		line := scanner.Text()
		stracer.Traceln("raw line to parse:", line)
		if commentExp.MatchString(line) {
			continue
		} else if sectionExp.MatchString(line) {
			// Removes spaces too so "[ section]" is parsed correctly
			section = strings.Trim(line, "[] ")
			continue
		}
		kv := parseKeyValue(line)
		// This even prevents empty line to be added
		if len(kv) > 0 {
			putInStruct(structValue, kv)
			conf[section] = append(conf[section], kv)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	stracer.Traceln("coded map:", conf)
	return nil
}
