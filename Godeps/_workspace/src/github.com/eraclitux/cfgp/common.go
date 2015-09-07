// cfgp - go configuration file parser package
// Copyright (c) 2015 Andrea Masi

// Package cfgp is a configuration parser fo Go.
//
// Just define a struct with needed configuration. Values are then taken from multiple source
// in this order of precendece:
//
// 	- command line arguments (which are automagically created and parsed)
// 	- configuration file
//
// Tags
//
// Default is to use lower cased field names in struct to create command line arguments.
// Tags can be used to specify different names, command line help message
// and section in conf file.
//
// Format is:
//	<name>,<help message>,<section in file>
//
// Simplest configuration file
//
// cfgp.Path variable can be set to the path of a configuration file.
// For default it is initialized to the value of evirontment variable:
//
//	CFGP_FILE_PATH
//
// Files ending with:
// 	ini|txt|cfg
// will be parsed as INI informal standard:
//
//	https://en.wikipedia.org/wiki/INI_file
//
// First letter of every key found is upper cased and than,
// a struct field with same name is searched:
//
// 	user -> User
//	portNumber -> PortNumber
//
// If such field name is not found than comparisson is made against
// key specified as first element in tag.
//
// cfgp tries to be modular and easily extendible to support different formats.
//
// This is a work in progress, better packages are out there.
package cfgp

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/eraclitux/stracer"
)

var ErrNeedPointer = errors.New("cfgp: pointer to struct expected")
var ErrFileFormat = errors.New("cfgp: unrecognized file format, only (ini|txt|cfg) supported")
var ErrUnknownFlagType = errors.New("cfgp: unknown flag type")

// Path is the path to configuration file.
// For default is populated with env var CFGP_FILE_PATH.
// This could be left empty if no configuration
// file is needed.
var Path string

func getStructValue(confPtr interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(confPtr)
	if v.Kind() == reflect.Ptr {
		return v.Elem(), nil
	}
	return reflect.Value{}, ErrNeedPointer
}

// myFlag implements Flag.Value.
// TODO is filed needed?
type myFlag struct {
	field      reflect.StructField
	fieldValue reflect.Value
	isBool     bool
}

// String () is used to print default value by PrintDefaults().
func (s *myFlag) String() string {
	// FIXME deal with non string types.
	return s.fieldValue.String()
}

// IsBoolFlag istructs the command-line parser
// to makes -name equivalent to -name=true rather than
// using the next command-line argument.
func (s *myFlag) IsBoolFlag() bool {
	return s.isBool
}

// assignType assing passed arg string to underlying Go type.
func assignType(fieldValue reflect.Value, arg string) error {
	if !fieldValue.CanSet() {
		return ErrUnknownFlagType
	}
	switch fieldValue.Kind() {
	case reflect.Int:
		n, err := strconv.Atoi(arg)
		if err != nil {
			return err
		}
		fieldValue.SetInt(int64(n))
	case reflect.Float64:
		f, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(f)
	case reflect.String:
		fieldValue.SetString(arg)
	case reflect.Bool:
		b, err := strconv.ParseBool(arg)
		if err != nil {
			return err
		}
		fieldValue.SetBool(b)
	default:
		return ErrUnknownFlagType
	}
	return nil
}

// Set converts passed arguments to actual Go types.
func (s *myFlag) Set(arg string) error {
	stracer.Traceln("setting flag", s.field.Name)
	err := assignType(s.fieldValue, arg)
	if err != nil {
		return err
	}
	return nil
}

func helpMessageFromTags(f reflect.StructField) (string, bool) {
	t := f.Tag.Get("cfgp")
	tags := strings.Split(t, ",")
	if len(tags) == 3 {
		return tags[1], true
	}
	return "", false
}

func makeHelpMessage(f reflect.StructField) string {
	var helpM string
	switch f.Type.Kind() {
	case reflect.Int:
		if m, ok := helpMessageFromTags(f); ok {
			helpM = m + ", an int"
		} else {
			helpM = "set an int"
		}
	case reflect.String:
		if m, ok := helpMessageFromTags(f); ok {
			helpM = m + ", a string"
		} else {
			helpM = "set a string"
		}
	case reflect.Bool:
		if m, ok := helpMessageFromTags(f); ok {
			helpM = m + ", a bool"
		} else {
			helpM = "set a bool"
		}
	case reflect.Float64:
		if m, ok := helpMessageFromTags(f); ok {
			helpM = m + ", a float64"
		} else {
			helpM = "set a float64"
		}
	default:
		helpM = "unknown flag kind"
	}
	return helpM
}

func isBool(v reflect.Value) bool {
	if v.Kind() == reflect.Bool {
		return true
	}
	return false
}

func nameFromTags(f reflect.StructField) (string, bool) {
	t := f.Tag.Get("cfgp")
	tags := strings.Split(t, ",")
	// if name position is empty return false es:
	// `cfgp:",help message,"`
	if len(tags) == 3 && tags[0] != "" {
		return tags[0], true
	}
	return "", false
}

// FIXME can we semplify using structType := structValue.Type()?
func createFlag(f reflect.StructField, fieldValue reflect.Value, fs *flag.FlagSet) {
	name := strings.ToLower(f.Name)
	if n, ok := nameFromTags(f); ok {
		name = n
	}
	stracer.Traceln("creating flag:", name)
	fs.Var(&myFlag{f, fieldValue, isBool(fieldValue)}, name, makeHelpMessage(f))
}

// hasTestFlag helps to identify if a test
// is running with flags that can make
// flagSet.Parse() fail.
func hasTestFlag([]string) bool {
	for _, f := range os.Args[1:] {
		if f == `-test.v=true` {
			stracer.Traceln("test flag found")
			return true
		}
	}
	return false
}

// parseFlags parses struct fields, creates command line arguments
// and check if they are passed as arguments.
func parseFlags(s reflect.Value) error {
	flagSet := flag.NewFlagSet("cfgp", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flagSet.PrintDefaults()
	}
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		fieldValue := s.Field(i)
		if fieldValue.CanSet() {
			createFlag(typeOfT.Field(i), fieldValue, flagSet)
		}
	}
	args := os.Args[1:]
	if hasTestFlag(os.Args[1:]) {
		args = []string{}
	}
	err := flagSet.Parse(args)
	if err != nil {
		return err
	}
	return nil
}

// Parse popolate passed struct (via pointer) with configuration from various source.
// It guesses configuration type by file extention and call specific parser.
// (.ini|.txt|.cfg) are evaluated as INI files which is to only format supported for now.
// path can be an empty string to disable file parsing.
func Parse(confPtr interface{}) error {
	structValue, err := getStructValue(confPtr)
	if err != nil {
		return err
	}
	if Path != "" {
		if match, _ := regexp.MatchString(`\.(ini|txt|cfg)$`, Path); match {
			err := parseINI(Path, structValue)
			if err != nil {
				return err
			}
		} else if match, _ := regexp.MatchString(`\.(yaml)$`, Path); match {
			return errors.New("YAML not yet implemented. Want you help?")
		} else {
			return ErrFileFormat
		}
	}
	// Command line arguments override configuration file.
	err = parseFlags(structValue)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	Path = os.Getenv("CFGP_FILE_PATH")
	stracer.Traceln("init file path:", Path)
}
