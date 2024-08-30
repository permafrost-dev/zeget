package filters

import (
	"strings"

	"github.com/permafrost-dev/zeget/lib/assets"
)

type FilterHandler func(assets.Asset, []string) bool

type FilterAction byte

const (
	FilterActionInclude FilterAction = iota
	FilterActionExclude FilterAction = iota
)

type Filter struct {
	Name       string
	Handler    FilterHandler
	Action     FilterAction
	Args       []string
	Definition string
}

func NewFilter(name string, handler FilterHandler, action FilterAction, args ...string) *Filter {
	return &Filter{
		Name:       name,
		Handler:    handler,
		Action:     action,
		Args:       args,
		Definition: name + "(" + strings.Join(args, ",") + ")",
	}
}

func (f *Filter) Apply(item assets.Asset) bool {
	return f.Handler(item, f.Args)
}

func (f *Filter) WithArgs(args ...string) *Filter {
	f.Args = args
	return f
}
