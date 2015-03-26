// Copyright © 2013 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//Package cobra is a commander providing a simple interface to create powerful modern CLI interfaces.
//In addition to providing an interface, Cobra simultaneously provides a controller to organize your application code.
package cobra

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
)

// Command is just that, a command for your application.
// eg.  'go run' ... 'run' is the command. Cobra requires
// you to define the usage and description as part of your command
// definition to ensure usability.
type Command struct {
	// Name is the command name, usually the executable's name.
	name string
	// The one-line usage message.
	Use string
	// An array of aliases that can be used instead of the first word in Use.
	Aliases []string
	// The short description shown in the 'help' output.
	Short string
	// The long message shown in the 'help <this-command>' output.
	Long string
	// Examples of how to use the command
	Example string
	// Full set of flags
	flags *flag.FlagSet
	// Set of flags childrens of this command will inherit
	pflags *flag.FlagSet
	// Flags that are declared specifically by this command (not inherited).
	lflags *flag.FlagSet
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)
	// Commands is the list of commands supported by this program.
	commands []*Command
	// Parent Command for this command
	parent *Command
	// max lengths of commands' string lengths for use in padding
	commandsMaxUseLen         int
	commandsMaxCommandPathLen int
	commandsMaxNameLen        int

	flagErrorBuf *bytes.Buffer
	cmdErrorBuf  *bytes.Buffer

	args          []string                 // actual args parsed from flags
	output        *io.Writer               // nil means stderr; use Out() method instead
	usageFunc     func(*Command) error     // Usage can be defined by application
	usageTemplate string                   // Can be defined by Application
	helpTemplate  string                   // Can be defined by Application
	helpFunc      func(*Command, []string) // Help can be defined by application
	helpCommand   *Command                 // The help command
	helpFlagVal   bool
}

// os.Args[1:] by default, if desired, can be overridden
// particularly useful when testing.
func (c *Command) SetArgs(a []string) {
	c.args = a
}

func (c *Command) getOut(def io.Writer) io.Writer {
	if c.output != nil {
		return *c.output
	}

	if c.HasParent() {
		return c.parent.Out()
	} else {
		return def
	}
}

func (c *Command) Out() io.Writer {
	return c.getOut(os.Stderr)
}

func (c *Command) getOutOrStdout() io.Writer {
	return c.getOut(os.Stdout)
}

// SetOutput sets the destination for usage and error messages.
// If output is nil, os.Stderr is used.
func (c *Command) SetOutput(output io.Writer) {
	c.output = &output
}

// Usage can be defined by application
func (c *Command) SetUsageFunc(f func(*Command) error) {
	c.usageFunc = f
}

// Can be defined by Application
func (c *Command) SetUsageTemplate(s string) {
	c.usageTemplate = s
}

// Can be defined by Application
func (c *Command) SetHelpFunc(f func(*Command, []string)) {
	c.helpFunc = f
}

func (c *Command) SetHelpCommand(cmd *Command) {
	c.helpCommand = cmd
}

// Can be defined by Application
func (c *Command) SetHelpTemplate(s string) {
	c.helpTemplate = s
}

func (c *Command) UsageFunc() (f func(*Command) error) {
	if c.usageFunc != nil {
		return c.usageFunc
	}

	if c.HasParent() {
		return c.parent.UsageFunc()
	} else {
		return func(c *Command) error {
			err := tmpl(c.Out(), c.UsageTemplate(), c)
			return err
		}
	}
}
func (c *Command) HelpFunc() func(*Command, []string) {
	if c.helpFunc != nil {
		return c.helpFunc
	}

	if c.HasParent() {
		return c.parent.HelpFunc()
	} else {
		return func(c *Command, args []string) {
			if len(args) == 0 {
				// Help called without any topic, calling on root
				c.Root().Help()
				return
			}

			cmd, _, e := c.Root().Find(args)
			if cmd == nil || e != nil {
				c.Printf("Unknown help topic %#q.", args)

				c.Root().Usage()
			} else {
				err := cmd.Help()
				if err != nil {
					c.Println(err)
				}
			}
		}
	}
}

var minUsagePadding int = 25

func (c *Command) UsagePadding() int {
	if c.parent == nil || minUsagePadding > c.parent.commandsMaxUseLen {
		return minUsagePadding
	} else {
		return c.parent.commandsMaxUseLen
	}
}

var minCommandPathPadding int = 11

//
func (c *Command) CommandPathPadding() int {
	if c.parent == nil || minCommandPathPadding > c.parent.commandsMaxCommandPathLen {
		return minCommandPathPadding
	} else {
		return c.parent.commandsMaxCommandPathLen
	}
}

var minNamePadding int = 11

func (c *Command) NamePadding() int {
	if c.parent == nil || minNamePadding > c.parent.commandsMaxNameLen {
		return minNamePadding
	} else {
		return c.parent.commandsMaxNameLen
	}
}

func (c *Command) UsageTemplate() string {
	if c.usageTemplate != "" {
		return c.usageTemplate
	}

	if c.HasParent() {
		return c.parent.UsageTemplate()
	} else {
		return `{{ $cmd := . }}
Usage: {{if .Runnable}}
  {{.UseLine}}{{if .HasFlags}} [flags]{{end}}{{end}}{{if .HasSubCommands}}
  {{ .CommandPath}} [command]{{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}

Examples:
{{ .Example }}
{{end}}{{ if .HasSubCommands}}

Available Commands: {{range .Commands}}{{if .Runnable}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}
{{end}}
{{ if .HasLocalFlags}}Flags:
{{.LocalFlags.FlagUsages}}{{end}}
{{ if .HasInheritedFlags}}Global Flags:
{{.InheritedFlags.FlagUsages}}{{end}}{{if .HasParent}}{{if and (gt .Commands 0) (gt .Parent.Commands 1) }}
Additional help topics: {{if gt .Commands 0 }}{{range .Commands}}{{if not .Runnable}} {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if gt .Parent.Commands 1 }}{{range .Parent.Commands}}{{if .Runnable}}{{if not (eq .Name $cmd.Name) }}{{end}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{end}}
{{end}}{{ if .HasSubCommands }}
Use "{{.Root.Name}} help [command]" for more information about a command.
{{end}}`
	}
}

func (c *Command) HelpTemplate() string {
	if c.helpTemplate != "" {
		return c.helpTemplate
	}

	if c.HasParent() {
		return c.parent.HelpTemplate()
	} else {
		return `{{with or .Long .Short }}{{. | trim}}{{end}}
{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}
`
	}
}

// Really only used when casting a command to a commander
func (c *Command) resetChildrensParents() {
	for _, x := range c.commands {
		x.parent = c
	}
}

// Test if the named flag is a boolean flag.
func isBooleanFlag(name string, f *flag.FlagSet) bool {
	flag := f.Lookup(name)
	if flag == nil {
		return false
	}
	return flag.Value.Type() == "bool"
}

// Test if the named flag is a boolean flag.
func isBooleanShortFlag(name string, f *flag.FlagSet) bool {
	result := false
	f.VisitAll(func(f *flag.Flag) {
		if f.Shorthand == name && f.Value.Type() == "bool" {
			result = true
		}
	})
	return result
}

func stripFlags(args []string, c *Command) []string {
	if len(args) < 1 {
		return args
	}
	c.mergePersistentFlags()

	commands := []string{}

	inQuote := false
	inFlag := false
	for _, y := range args {
		if !inQuote {
			switch {
			case strings.HasPrefix(y, "\""):
				inQuote = true
			case strings.Contains(y, "=\""):
				inQuote = true
			case strings.HasPrefix(y, "--") && !strings.Contains(y, "="):
				// TODO: this isn't quite right, we should really check ahead for 'true' or 'false'
				inFlag = !isBooleanFlag(y[2:], c.Flags())
			case strings.HasPrefix(y, "-") && !strings.Contains(y, "=") && len(y) == 2 && !isBooleanShortFlag(y[1:], c.Flags()):
				inFlag = true
			case inFlag:
				inFlag = false
			case !strings.HasPrefix(y, "-"):
				commands = append(commands, y)
				inFlag = false
			}
		}

		if strings.HasSuffix(y, "\"") && !strings.HasSuffix(y, "\\\"") {
			inQuote = false
		}
	}

	return commands
}

func argsMinusX(args []string, x string) []string {
	newargs := []string{}

	for _, y := range args {
		if x != y {
			newargs = append(newargs, y)
		}
	}
	return newargs
}

// find the target command given the args and command tree
// Meant to be run on the highest node. Only searches down.
func (c *Command) Find(arrs []string) (*Command, []string, error) {
	if c == nil {
		return nil, nil, fmt.Errorf("Called find() on a nil Command")
	}

	if len(arrs) == 0 {
		return c.Root(), arrs, nil
	}

	var innerfind func(*Command, []string) (*Command, []string)

	innerfind = func(c *Command, args []string) (*Command, []string) {
		if len(args) > 0 && c.HasSubCommands() {
			argsWOflags := stripFlags(args, c)
			if len(argsWOflags) > 0 {
				matches := make([]*Command, 0)
				for _, cmd := range c.commands {
					if cmd.Name() == argsWOflags[0] || cmd.HasAlias(argsWOflags[0]) { // exact name or alias match
						return innerfind(cmd, argsMinusX(args, argsWOflags[0]))
					} else if EnablePrefixMatching {
						if strings.HasPrefix(cmd.Name(), argsWOflags[0]) { // prefix match
							matches = append(matches, cmd)
						}
						for _, x := range cmd.Aliases {
							if strings.HasPrefix(x, argsWOflags[0]) {
								matches = append(matches, cmd)
							}
						}
					}
				}

				// only accept a single prefix match - multiple matches would be ambiguous
				if len(matches) == 1 {
					return innerfind(matches[0], argsMinusX(args, argsWOflags[0]))
				}
			}
		}

		return c, args
	}

	commandFound, a := innerfind(c, arrs)

	// if commander returned and the first argument (if it exists) doesn't
	// match the command name, return nil & error
	if commandFound.Name() == c.Name() && len(arrs[0]) > 0 && commandFound.Name() != arrs[0] {
		return nil, a, fmt.Errorf("unknown command %q\nRun 'help' for usage.\n", a[0])
	}

	return commandFound, a, nil
}

func (c *Command) Root() *Command {
	var findRoot func(*Command) *Command

	findRoot = func(x *Command) *Command {
		if x.HasParent() {
			return findRoot(x.parent)
		} else {
			return x
		}
	}

	return findRoot(c)
}

// execute the command determined by args and the command tree
func (c *Command) findAndExecute(args []string) (findErr error, err error) {

	cmd, a, e := c.Find(args)
	if e != nil {
		return e, nil
	}
	return nil, cmd.execute(a)
}

func (c *Command) execute(a []string) (err error) {
	if c == nil {
		return fmt.Errorf("Called Execute() on a nil Command")
	}

	err = c.ParseFlags(a)
	if err == flag.ErrHelp {
		c.Help()
		return nil
	}
	if err != nil {
		// We're writing subcommand usage to root command's error buffer to have it displayed to the user
		r := c.Root()
		if r.cmdErrorBuf == nil {
			r.cmdErrorBuf = new(bytes.Buffer)
		}
		// for writing the usage to the buffer we need to switch the output temporarily
		// since Out() returns root output, you also need to revert that on root
		out := r.Out()
		r.SetOutput(r.cmdErrorBuf)
		c.Usage()
		r.SetOutput(out)
		return err
	} else {
		// If help is called, regardless of other flags, we print that.
		// Print help also if c.Run is nil.
		if c.helpFlagVal || !c.Runnable() {
			c.Help()
			return nil
		}

		c.preRun()
		argWoFlags := c.Flags().Args()
		c.Run(c, argWoFlags)
		return nil
	}
}

func (c *Command) preRun() {
	for _, x := range initializers {
		x()
	}
}

func (c *Command) errorMsgFromParse() string {
	s := c.flagErrorBuf.String()

	x := strings.Split(s, "\n")

	if len(x) > 0 {
		return x[0]
	} else {
		return ""
	}
}

// Call execute to use the args (os.Args[1:] by default)
// and run through the command tree finding appropriate matches
// for commands and then corresponding flags.
func (c *Command) Execute() (err error) {

	// Regardless of what command execute is called on, run on Root only
	if c.HasParent() {
		return c.Root().Execute()
	}

	// initialize help as the last point possible to allow for user
	// overriding
	c.initHelp()

	var args []string

	if len(c.args) == 0 {
		args = os.Args[1:]
	} else {
		args = c.args
	}

	var findErr error
	var rootCommandError error

	if len(args) == 0 {
		// Only the executable is called and the root is runnable, run it
		if c.Runnable() {
			rootCommandError = c.execute([]string(nil))
			err = rootCommandError

		} else {
			c.Help()
		}
	} else {
		findErr, err = c.findAndExecute(args)
	}

	// if there was an error caused by running the root command or an inability to locate the requested command
	// try to reparse the flags and run using the root command.
	if (findErr != nil || rootCommandError != nil) && c.Runnable() {
		// This is pretty much a custom version of the *Command.execute method
		// with a few differences because it's the final command (no fall back)
		e := c.ParseFlags(args)
		if e != nil {
			// Flags parsing had an error.
			// If an error happens here, we have to report it to the user
			c.Println(e.Error())
			// If an error happens search also for subcommand info about that
			if c.cmdErrorBuf != nil && c.cmdErrorBuf.Len() > 0 {
				c.Println(c.cmdErrorBuf.String())
			} else {
				c.Usage()
			}
			err = e
			return
		} else {
			// If help is called, regardless of other flags, we print that
			if c.helpFlagVal {
				c.Help()
				return nil
			}

			argWoFlags := c.Flags().Args()
			if len(argWoFlags) > 0 {
				// If there are arguments (not flags) one of the earlier
				// cases should have caught it.. It means invalid usage
				// print the usage
				c.Usage()
			} else {
				// Only flags left... Call root.Run
				c.preRun()
				c.Run(c, argWoFlags)
				err = nil
			}
		}
	}

	if err != nil {
		if err == flag.ErrHelp {
			c.Help()
		} else {
			c.Println("Error:", err.Error())
			c.Printf("%v: invalid command %#q\n", c.Root().Name(), os.Args[1:])
			c.Printf("Run '%v help' for usage\n", c.Root().Name())
		}
	}

	return
}

func (c *Command) initHelp() {
	if c.helpCommand == nil {
		if !c.HasSubCommands() {
			return
		}

		c.helpCommand = &Command{
			Use:   "help [command]",
			Short: "Help about any command",
			Long: `Help provides help for any command in the application.
    Simply type ` + c.Name() + ` help [path to command] for full details.`,
			Run: c.HelpFunc(),
		}
	}
	c.AddCommand(c.helpCommand)
}

// Used for testing
func (c *Command) ResetCommands() {
	c.commands = nil
	c.helpCommand = nil
	c.cmdErrorBuf = new(bytes.Buffer)
	c.cmdErrorBuf.Reset()
}

//Commands returns a slice of child commands.
func (c *Command) Commands() []*Command {
	return c.commands
}

// AddCommand adds one or more commands to this parent command.
func (c *Command) AddCommand(cmds ...*Command) {
	for i, x := range cmds {
		if cmds[i] == c {
			panic("Command can't be a child of itself")
		}
		cmds[i].parent = c
		// update max lengths
		usageLen := len(x.Use)
		if usageLen > c.commandsMaxUseLen {
			c.commandsMaxUseLen = usageLen
		}
		commandPathLen := len(x.CommandPath())
		if commandPathLen > c.commandsMaxCommandPathLen {
			c.commandsMaxCommandPathLen = commandPathLen
		}
		nameLen := len(x.Name())
		if nameLen > c.commandsMaxNameLen {
			c.commandsMaxNameLen = nameLen
		}
		c.commands = append(c.commands, x)
	}
}

// AddCommand removes one or more commands from a parent command.
func (c *Command) RemoveCommand(cmds ...*Command) {
	commands := []*Command{}
main:
	for _, command := range c.commands {
		for _, cmd := range cmds {
			if command == cmd {
				command.parent = nil
				continue main
			}
		}
		commands = append(commands, command)
	}
	c.commands = commands
	// recompute all lengths
	c.commandsMaxUseLen = 0
	c.commandsMaxCommandPathLen = 0
	c.commandsMaxNameLen = 0
	for _, command := range c.commands {
		usageLen := len(command.Use)
		if usageLen > c.commandsMaxUseLen {
			c.commandsMaxUseLen = usageLen
		}
		commandPathLen := len(command.CommandPath())
		if commandPathLen > c.commandsMaxCommandPathLen {
			c.commandsMaxCommandPathLen = commandPathLen
		}
		nameLen := len(command.Name())
		if nameLen > c.commandsMaxNameLen {
			c.commandsMaxNameLen = nameLen
		}
	}
}

// Convenience method to Print to the defined output
func (c *Command) Print(i ...interface{}) {
	fmt.Fprint(c.Out(), i...)
}

// Convenience method to Println to the defined output
func (c *Command) Println(i ...interface{}) {
	str := fmt.Sprintln(i...)
	c.Print(str)
}

// Convenience method to Printf to the defined output
func (c *Command) Printf(format string, i ...interface{}) {
	str := fmt.Sprintf(format, i...)
	c.Print(str)
}

// Output the usage for the command
// Used when a user provides invalid input
// Can be defined by user by overriding UsageFunc
func (c *Command) Usage() error {
	c.mergePersistentFlags()
	err := c.UsageFunc()(c)
	return err
}

// Output the help for the command
// Used when a user calls help [command]
// by the default HelpFunc in the commander
func (c *Command) Help() error {
	c.mergePersistentFlags()
	err := tmpl(c.getOutOrStdout(), c.HelpTemplate(), c)
	return err
}

func (c *Command) UsageString() string {
	tmpOutput := c.output
	bb := new(bytes.Buffer)
	c.SetOutput(bb)
	c.Usage()
	c.output = tmpOutput
	return bb.String()
}

// CommandPath returns the full path to this command.
func (c *Command) CommandPath() string {
	str := c.Name()
	x := c
	for x.HasParent() {
		str = x.parent.Name() + " " + str
		x = x.parent
	}
	return str
}

//The full usage for a given command (including parents)
func (c *Command) UseLine() string {
	str := ""
	if c.HasParent() {
		str = c.parent.CommandPath() + " "
	}
	return str + c.Use
}

// For use in determining which flags have been assigned to which commands
// and which persist
func (c *Command) DebugFlags() {
	c.Println("DebugFlags called on", c.Name())
	var debugflags func(*Command)

	debugflags = func(x *Command) {
		if x.HasFlags() || x.HasPersistentFlags() {
			c.Println(x.Name())
		}
		if x.HasFlags() {
			x.flags.VisitAll(func(f *flag.Flag) {
				if x.HasPersistentFlags() {
					if x.persistentFlag(f.Name) == nil {
						c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [L]")
					} else {
						c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [LP]")
					}
				} else {
					c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [L]")
				}
			})
		}
		if x.HasPersistentFlags() {
			x.pflags.VisitAll(func(f *flag.Flag) {
				if x.HasFlags() {
					if x.flags.Lookup(f.Name) == nil {
						c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [P]")
					}
				} else {
					c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [P]")
				}
			})
		}
		c.Println(x.flagErrorBuf)
		if x.HasSubCommands() {
			for _, y := range x.commands {
				debugflags(y)
			}
		}
	}

	debugflags(c)
}

// Name returns the command's name: the first word in the use line.
func (c *Command) Name() string {
	if c.name != "" {
		return c.name
	}
	name := c.Use
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// Determine if a given string is an alias of the command.
func (c *Command) HasAlias(s string) bool {
	for _, a := range c.Aliases {
		if a == s {
			return true
		}
	}
	return false
}

func (c *Command) NameAndAliases() string {
	return strings.Join(append([]string{c.Name()}, c.Aliases...), ", ")
}

func (c *Command) HasExample() bool {
	return len(c.Example) > 0
}

// Determine if the command is itself runnable
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// Determine if the command has children commands
func (c *Command) HasSubCommands() bool {
	return len(c.commands) > 0
}

// Determine if the command is a child command
func (c *Command) HasParent() bool {
	return c.parent != nil
}

// Get the complete FlagSet that applies to this command (local and persistent declared here and by all parents)
func (c *Command) Flags() *flag.FlagSet {
	if c.flags == nil {
		c.flags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.flags.SetOutput(c.flagErrorBuf)
		c.PersistentFlags().BoolVarP(&c.helpFlagVal, "help", "h", false, "help for "+c.Name())
	}
	return c.flags
}

// Get the local FlagSet specifically set in the current command
func (c *Command) LocalFlags() *flag.FlagSet {
	c.mergePersistentFlags()

	local := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	c.lflags.VisitAll(func(f *flag.Flag) {
		local.AddFlag(f)
	})
	return local
}

// All Flags which were inherited from parents commands
func (c *Command) InheritedFlags() *flag.FlagSet {
	c.mergePersistentFlags()

	inherited := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	local := c.LocalFlags()

	var rmerge func(x *Command)

	rmerge = func(x *Command) {
		if x.HasPersistentFlags() {
			x.PersistentFlags().VisitAll(func(f *flag.Flag) {
				if inherited.Lookup(f.Name) == nil && local.Lookup(f.Name) == nil {
					inherited.AddFlag(f)
				}
			})
		}
		if x.HasParent() {
			rmerge(x.parent)
		}
	}

	if c.HasParent() {
		rmerge(c.parent)
	}

	return inherited
}

// All Flags which were not inherited from parent commands
func (c *Command) NonInheritedFlags() *flag.FlagSet {
	return c.LocalFlags()
}

// Get the Persistent FlagSet specifically set in the current command
func (c *Command) PersistentFlags() *flag.FlagSet {
	if c.pflags == nil {
		c.pflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.pflags.SetOutput(c.flagErrorBuf)
	}
	return c.pflags
}

// For use in testing
func (c *Command) ResetFlags() {
	c.flagErrorBuf = new(bytes.Buffer)
	c.flagErrorBuf.Reset()
	c.flags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	c.flags.SetOutput(c.flagErrorBuf)
	c.pflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	c.pflags.SetOutput(c.flagErrorBuf)
}

// Does the command contain any flags (local plus persistent from the entire structure)
func (c *Command) HasFlags() bool {
	return c.Flags().HasFlags()
}

// Does the command contain persistent flags
func (c *Command) HasPersistentFlags() bool {
	return c.PersistentFlags().HasFlags()
}

// Does the command has flags specifically declared locally
func (c *Command) HasLocalFlags() bool {
	return c.LocalFlags().HasFlags()
}

func (c *Command) HasInheritedFlags() bool {
	return c.InheritedFlags().HasFlags()
}

// Climbs up the command tree looking for matching flag
func (c *Command) Flag(name string) (flag *flag.Flag) {
	flag = c.Flags().Lookup(name)

	if flag == nil {
		flag = c.persistentFlag(name)
	}

	return
}

// recursively find matching persistent flag
func (c *Command) persistentFlag(name string) (flag *flag.Flag) {
	if c.HasPersistentFlags() {
		flag = c.PersistentFlags().Lookup(name)
	}

	if flag == nil && c.HasParent() {
		flag = c.parent.persistentFlag(name)
	}
	return
}

// Parses persistent flag tree & local flags
func (c *Command) ParseFlags(args []string) (err error) {
	c.mergePersistentFlags()
	err = c.Flags().Parse(args)
	return
}

func (c *Command) Parent() *Command {
	return c.parent
}

func (c *Command) mergePersistentFlags() {
	var rmerge func(x *Command)

	// Save the set of local flags
	if c.lflags == nil {
		c.lflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.lflags.SetOutput(c.flagErrorBuf)
		addtolocal := func(f *flag.Flag) {
			c.lflags.AddFlag(f)
		}
		c.Flags().VisitAll(addtolocal)
		c.PersistentFlags().VisitAll(addtolocal)
	}
	rmerge = func(x *Command) {
		if x.HasPersistentFlags() {
			x.PersistentFlags().VisitAll(func(f *flag.Flag) {
				if c.Flags().Lookup(f.Name) == nil {
					c.Flags().AddFlag(f)
				}
			})
		}
		if x.HasParent() {
			rmerge(x.parent)
		}
	}

	rmerge(c)
}
