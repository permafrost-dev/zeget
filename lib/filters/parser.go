package filters

import (
	"regexp"
	"strings"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseDefinitions(definitions string) []*Filter {
	defs := strings.Split(definitions, ";")

	filters := make([]*Filter, 0)

	for _, def := range defs {
		filter := p.ParseDefinition(def)
		if filter != nil {
			filters = append(filters, filter)
		}
	}

	return filters
}

// ParseDefinition parses a string definition of a filter, like "always(abc,def)" or "never(.deb)"
// and returns a Filter struct with the appropriate values set.
func (p *Parser) ParseDefinition(definition string) *Filter {
	re := regexp.MustCompile(`([a-zA-Z]+)\((.*)\)`)
	matches := re.FindStringSubmatch(definition)

	if len(matches) < 3 {
		return nil
	}

	name := matches[1]
	argStr := matches[2]
	args := strings.Split(argStr, ",")

	if FilterMap[name] == nil {
		return nil
	}

	return &Filter{
		Name:       FilterMap[name].Name,
		Handler:    FilterMap[name].Handler,
		Action:     FilterMap[name].Action,
		Args:       args,
		Definition: definition,
	}
}
