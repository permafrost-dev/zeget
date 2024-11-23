package filters

import (
	"context"
	"fmt"
	"log"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

func ParseCELFilter(expression string) (*cel.Program, error) {
	makeFetch := func(ctx any) cel.EnvOption {
		fn := func(arg ref.Val) ref.Val {
			return types.NewErr("stateful context not bound")
		}
		if ctx != nil {
			fn = func(resource ref.Val) ref.Val {
				return types.DefaultTypeAdapter.NativeToValue(
					ctx.(context.Context).Value(string(resource.(types.String))),
				)
			}
		}
		return cel.Function("fetch",
			cel.Overload("fetch_string",
				[]*cel.Type{cel.StringType}, cel.StringType,
				cel.UnaryBinding(fn),
			),
		)
	}

	// var adapter types.Adapter

	// // Define the slice of strings
	// elements := []string{"apple", "banana", "cherry"}

	env, err := cel.NewEnv(
		cel.Variable("name", cel.StringType),
		cel.Variable("group", cel.StringType),
		cel.Variable("i", cel.StringType),
		cel.Variable("you", cel.StringType),
		cel.Variable("assets", cel.ListType(cel.StringType)),
		// Function to generate a greeting from one person to another: i.greet(you)
		cel.Function("greet",
			cel.MemberOverload("string_greet_string", []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				cel.BinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					return types.String(fmt.Sprintf("Hello %s! Nice to meet you, I'm %s.\n", rhs, lhs))
				}),
			),
		),
		makeFetch(nil),
	)

	if err != nil {
		log.Fatalf("environment creation error: %s", err)
	}

	ast, issues := env.Compile(`assets.filter(asset, asset.startsWith("a")).map(a, a.size())`)
	if issues != nil && issues.Err() != nil {
		log.Fatalf("type-check error: %s", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		log.Fatalf("program construction error: %s", err)
	}

	return &prg, nil
}

func CELTest() {
	prg, err := ParseCELFilter(`assets.some(asset, asset.startsWith("a"))`)
	if err != nil {
		log.Fatalf("error parsing CEL filter: %s", err)
	}

	// Evaluate the program in a context.
	out, details, err := (*prg).Eval(map[string]interface{}{
		"name":   "/groups/abc",
		"group":  "abc",
		"i":      "CEL",
		"assets": []string{"abc", "def", "ghi", "aaadd"},
		"you":    func() ref.Val { return types.String("world") },
	})

	if err != nil {
		log.Fatalf("error evaluating program: %s", err)
	}

	log.Printf("result: %v, details: %v", out, details)
}
