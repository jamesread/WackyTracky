package ruleeval

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuleEnv(t *testing.T) {
	// Saturday
	env := NewRuleEnv(time.Date(2025, 2, 22, 14, 30, 0, 0, time.UTC))
	assert.Equal(t, "Sat", env.D)
	assert.Equal(t, 14, env.H)
	assert.Equal(t, 30, env.M)
}

func TestEval(t *testing.T) {
	env := RuleEnv{D: "Sat", H: 14, M: 30}

	result, err := Eval("D == \"Sat\" || D == \"Sun\"", env)
	require.NoError(t, err)
	assert.True(t, result)

	env.D = "Mon"
	result, err = Eval("D == \"Sat\" || D == \"Sun\"", env)
	require.NoError(t, err)
	assert.False(t, result)
}

func TestCompileError(t *testing.T) {
	_, err := Compile("D == Sat") // unquoted Sat is invalid in expr
	assert.Error(t, err)
}

func TestRuleTest_CompilesAndResult(t *testing.T) {
	program, err := Compile("D == \"Sat\"")
	require.NoError(t, err)
	env := RuleEnv{D: "Sat", H: 0, M: 0}
	result, err := Run(program, env)
	require.NoError(t, err)
	assert.True(t, result)
}
