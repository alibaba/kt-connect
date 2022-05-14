package options

import (
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"reflect"
	"unsafe"
)

type OptionConfig struct {
	Target string
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
		name := util.UnCapitalize(c.Target)
		field := reflect.ValueOf(optionStore).Elem().FieldByName(c.Target)
		switch c.DefaultValue.(type) {
		case string:
			fieldPtr := (*string)(unsafe.Pointer(field.UnsafeAddr()))
			if c.Alias != "" {
				flags.StringVarP(fieldPtr, name, c.Alias, c.DefaultValue.(string), c.Description)
			} else {
				flags.StringVar(fieldPtr, name, c.DefaultValue.(string), c.Description)
			}
		case int:
			fieldPtr := (*int)(unsafe.Pointer(field.UnsafeAddr()))
			if c.Alias != "" {
				flags.IntVarP(fieldPtr, name, c.Alias, c.DefaultValue.(int), c.Description)
			} else {
				flags.IntVar(fieldPtr, name, c.DefaultValue.(int), c.Description)
			}
		case bool:
			fieldPtr := (*bool)(unsafe.Pointer(field.UnsafeAddr()))
			if c.Alias != "" {
				flags.BoolVarP(fieldPtr, name, c.Alias, c.DefaultValue.(bool), c.Description)
			} else {
				flags.BoolVar(fieldPtr, name, c.DefaultValue.(bool), c.Description)
			}
		}
		if c.Hidden {
			_ = flags.MarkHidden(name)
		}
		if c.Required {
			_ = cmd.MarkFlagRequired(name)
		}
	}
}