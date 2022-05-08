package options

import (
	"github.com/spf13/cobra"
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

func SetOptions(cmd *cobra.Command, flags *flag.FlagSet, optionStore any, config []OptionConfig) {
	cmd.Long = cmd.Short
	cmd.Flags().SortFlags = false
	cmd.InheritedFlags().SortFlags = false
	flags.SortFlags = false
	for _, c := range config {
		field := reflect.ValueOf(optionStore).Elem().FieldByName(c.Target)
		switch c.DefaultValue.(type) {
		case string:
			fieldPtr := (*string)(unsafe.Pointer(field.UnsafeAddr()))
			if c.Alias != "" {
				flags.StringVarP(fieldPtr, c.Name, c.Alias, c.DefaultValue.(string), c.Description)
			} else {
				flags.StringVar(fieldPtr, c.Name, c.DefaultValue.(string), c.Description)
			}
		case int:
			fieldPtr := (*int)(unsafe.Pointer(field.UnsafeAddr()))
			if c.Alias != "" {
				flags.IntVarP(fieldPtr, c.Name, c.Alias, c.DefaultValue.(int), c.Description)
			} else {
				flags.IntVar(fieldPtr, c.Name, c.DefaultValue.(int), c.Description)
			}
		case int64:
			fieldPtr := (*int64)(unsafe.Pointer(field.UnsafeAddr()))
			if c.Alias != "" {
				flags.Int64VarP(fieldPtr, c.Name, c.Alias, c.DefaultValue.(int64), c.Description)
			} else {
				flags.Int64Var(fieldPtr, c.Name, c.DefaultValue.(int64), c.Description)
			}
		case bool:
			fieldPtr := (*bool)(unsafe.Pointer(field.UnsafeAddr()))
			if c.Alias != "" {
				flags.BoolVarP(fieldPtr, c.Name, c.Alias, c.DefaultValue.(bool), c.Description)
			} else {
				flags.BoolVar(fieldPtr, c.Name, c.DefaultValue.(bool), c.Description)
			}
		}
		if c.Hidden {
			_ = flags.MarkHidden(c.Name)
		}
		if c.Required {
			_ = cmd.MarkFlagRequired(c.Name)
		}
	}
}
