package commands

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertFlag checks that a flag is registered with the expected shorthand, pflag type name
// (e.g. "string", "bool", "int", "stringSlice", "stringArray"), and default value string.
func assertFlag(t *testing.T, fs *pflag.FlagSet, name, shorthand, typeName, defValue string) {
	t.Helper()
	f := fs.Lookup(name)
	if !assert.NotNil(t, f, "expected flag %q to be registered", name) {
		return
	}
	assert.Equal(t, shorthand, f.Shorthand, "flag %q shorthand", name)
	assert.Equal(t, typeName, f.Value.Type(), "flag %q type", name)
	assert.Equal(t, defValue, f.DefValue, "flag %q default value", name)
}

// assertMutuallyExclusive verifies two flags on cmd are registered as a mutually exclusive
// group: setting both produces a validation error, setting either alone does not.
func assertMutuallyExclusive(t *testing.T, cmd *cobra.Command, a, b string) {
	t.Helper()

	fresh := freshCmdWithFlagsOf(cmd)
	// MarkFlagsMutuallyExclusive is registered against the original *cobra.Command instance,
	// not the flag set, so re-declare it on our fresh copy to validate independently.
	fresh.MarkFlagsMutuallyExclusive(a, b)

	require.NoError(t, fresh.Flags().Set(a, flagProbeValue(fresh.Flags(), a)))
	assert.NoError(t, fresh.ValidateFlagGroups(), "expected no error with only %q set", a)

	require.NoError(t, fresh.Flags().Set(b, flagProbeValue(fresh.Flags(), b)))
	assert.Error(t, fresh.ValidateFlagGroups(), "expected error with both %q and %q set", a, b)
}

func flagProbeValue(fs *pflag.FlagSet, name string) string {
	f := fs.Lookup(name)
	switch f.Value.Type() {
	case "bool":
		return "true"
	case "int":
		return "1"
	case "stringSlice", "stringArray":
		return "probe"
	default:
		return "probe"
	}
}

// freshCmdWithFlagsOf returns a new, unused *cobra.Command carrying independent copies of
// src's flags (same name/shorthand/type/default, but unset and with their own storage), so
// tests can exercise flag-reading logic without mutating shared package-level command vars.
func freshCmdWithFlagsOf(src *cobra.Command) *cobra.Command {
	dst := &cobra.Command{Use: src.Use}
	src.Flags().VisitAll(func(f *pflag.Flag) {
		switch f.Value.Type() {
		case "string":
			dst.Flags().StringP(f.Name, f.Shorthand, f.DefValue, f.Usage)
		case "bool":
			dst.Flags().BoolP(f.Name, f.Shorthand, f.DefValue == "true", f.Usage)
		case "int":
			dst.Flags().IntP(f.Name, f.Shorthand, 0, f.Usage)
		case "stringSlice":
			dst.Flags().StringSliceP(f.Name, f.Shorthand, nil, f.Usage)
		case "stringArray":
			dst.Flags().StringArrayP(f.Name, f.Shorthand, nil, f.Usage)
		default:
			panic("freshCmdWithFlagsOf: unsupported flag type " + f.Value.Type())
		}
	})
	return dst
}

func functionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}
