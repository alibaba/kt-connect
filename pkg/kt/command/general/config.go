package general

import (
	flag "github.com/spf13/pflag"
	"reflect"
	"unsafe"
)

type OptionConfig struct {
	Target string
	Name string
	Alias string
	DefaultValue any
	Description string
	Hidden bool
	Required bool
}

func SetOptions(flags *flag.FlagSet, optionStore any, config []OptionConfig) {
	flags.SortFlags = false
	for _, c := range config {
		field := reflect.ValueOf(optionStore).Elem().FieldByName(c.Target)
		switch c.DefaultValue.(type) {
		case string:
			fieldPtr := (*string)(unsafe.Pointer(field.UnsafeAddr()))
			flags.StringVar(fieldPtr, c.Name, c.DefaultValue.(string), c.Description)
		case int:
			fieldPtr := (*int)(unsafe.Pointer(field.UnsafeAddr()))
			flags.IntVar(fieldPtr, c.Name, c.DefaultValue.(int), c.Description)
		case int64:
			fieldPtr := (*int64)(unsafe.Pointer(field.UnsafeAddr()))
			flags.Int64Var(fieldPtr, c.Name, c.DefaultValue.(int64), c.Description)
		case bool:
			fieldPtr := (*bool)(unsafe.Pointer(field.UnsafeAddr()))
			flags.BoolVar(fieldPtr, c.Name, c.DefaultValue.(bool), c.Description)
		}
	}
}
