package ruleeval

import (
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// RuleEnv holds date/time components for hide-at-times expressions.
// D is short weekday (Mon, Tue, ..., Sun), H is hour 0-23, M is minute 0-59.
type RuleEnv struct {
	D string
	H int
	M int
}

var shortDay = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

// NewRuleEnv returns an env for the given time.
func NewRuleEnv(t time.Time) RuleEnv {
	return RuleEnv{
		D: shortDay[t.Weekday()],
		H: t.Hour(),
		M: t.Minute(),
	}
}

// Compile compiles a hide-at-times expression. The expression can use D, H, M.
func Compile(expression string) (*vm.Program, error) {
	return expr.Compile(expression, expr.Env(RuleEnv{}))
}

// Run evaluates the program with the given env and returns a boolean result.
func Run(program *vm.Program, env RuleEnv) (bool, error) {
	out, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}
	b, _ := out.(bool)
	return b, nil
}

// Eval compiles and runs the expression with the given env. Returns (result, nil) or (false, error).
func Eval(expression string, env RuleEnv) (bool, error) {
	program, err := Compile(expression)
	if err != nil {
		return false, err
	}
	return Run(program, env)
}
