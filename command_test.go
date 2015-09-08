package cobra

import (
	"reflect"
	"testing"
)

func init() {
	cmdHiddenFlags.Flags().BoolVarP(&flagbh, "boolh", "", false, "")
	cmdHiddenFlags.Flags().MarkHidden("boolh")

	cmdHiddenFlags.PersistentFlags().BoolVarP(&flagbph, "boolph", "", false, "")
	cmdHiddenFlags.PersistentFlags().MarkHidden("boolph")
}

// test to ensure hidden flags run as intended; if the the hidden flag fails to
// run, the output will be incorrect
func TestHiddenFlagExecutes(t *testing.T) {
	cmdHiddenFlags.execute([]string{"--boolh"})
	if outs != "hidden" {
		t.Errorf("Hidden flag failed to run!")
	}
}

// test to ensure hidden flags do not show up in usage/help text; if a flag is
// found by Lookup() it will be visible in usage/help text
func TestHiddenFlagsAreHidden(t *testing.T) {

	if cmdHiddenFlags.LocalFlags().Lookup("boolh") != nil {
		t.Errorf("unexpected flag 'boolh'")
	}

	if cmdHiddenFlags.InheritedFlags().Lookup("boolph") != nil {
		t.Errorf("unexpected flag 'boolph'")
	}
}

func TestStripFlags(t *testing.T) {
	tests := []struct {
		input  []string
		output []string
	}{
		{
			[]string{"foo", "bar"},
			[]string{"foo", "bar"},
		},
		{
			[]string{"foo", "--bar", "-b"},
			[]string{"foo"},
		},
		{
			[]string{"-b", "foo", "--bar", "bar"},
			[]string{},
		},
		{
			[]string{"-i10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-i=10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"--int=100", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-ib", "echo", "-bfoo", "baz"},
			[]string{"echo", "baz"},
		},
		{
			[]string{"-i=baz", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--int=baz", "-bbar", "-i", "foo", "blah"},
			[]string{"blah"},
		},
		{
			[]string{"--cat", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"-c", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--persist", "bar"},
			[]string{"bar"},
		},
		{
			[]string{"-p", "bar"},
			[]string{"bar"},
		},
	}

	cmdPrint := &Command{
		Use:   "print [string to print]",
		Short: "Print anything to the screen",
		Long:  `an utterly useless command for testing.`,
		Run: func(cmd *Command, args []string) {
			tp = args
		},
	}

	var flagi int
	var flagstr string
	var flagbool bool
	cmdPrint.PersistentFlags().BoolVarP(&flagbool, "persist", "p", false, "help for persistent one")
	cmdPrint.Flags().IntVarP(&flagi, "int", "i", 345, "help message for flag int")
	cmdPrint.Flags().StringVarP(&flagstr, "bar", "b", "bar", "help message for flag string")
	cmdPrint.Flags().BoolVarP(&flagbool, "cat", "c", false, "help message for flag bool")

	for _, test := range tests {
		output := stripFlags(test.input, cmdPrint)
		if !reflect.DeepEqual(test.output, output) {
			t.Errorf("expected: %v, got: %v", test.output, output)
		}
	}
}
