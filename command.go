package ishell

import (
	"bytes"
	"fmt"
	"sort"
	"text/tabwriter"
	"text/template"

	flag "github.com/spf13/pflag"
)

// Cmd is a shell command handler.
type Cmd struct {
	// Command name.
	Name string
	// Command name aliases.
	Aliases []string
	// Function to execute for the command.
	Func func(c *Context)
	// One liner help message for the command.
	Help string
	// More descriptive help message for the command.
	LongHelp string

	// Completer is custom autocomplete for command.
	// It takes in command arguments and returns
	// autocomplete options.
	// By default all commands get autocomplete of
	// subcommands.
	// A non-nil Completer overrides the default behaviour.
	Completer func(args []string) []string

	// subcommands.
	children map[string]*Cmd

	flags *flag.FlagSet
	// flagErrorBuf contains all error messages from pflag.
	flagErrorBuf *bytes.Buffer

	Root *Shell
}

func (c *Cmd) ShellName() string {
	return c.Root.Name
}

// AddCmd adds cmd as a subcommand.
func (c *Cmd) AddCmd(cmd *Cmd) {
	if c.children == nil {
		c.children = make(map[string]*Cmd)
	}
	c.children[cmd.Name] = cmd
}

// DeleteCmd deletes cmd from subcommands.
func (c *Cmd) DeleteCmd(name string) {
	delete(c.children, name)
}

// Children returns the subcommands of c.
func (c *Cmd) Children() []*Cmd {
	var cmds []*Cmd
	for _, cmd := range c.children {
		cmds = append(cmds, cmd)
	}
	sort.Sort(cmdSorter(cmds))
	return cmds
}

func (c *Cmd) hasSubcommand() bool {
	if len(c.children) > 1 {
		return true
	}
	if _, ok := c.children["help"]; !ok {
		return len(c.children) > 0
	}
	return false
}

// HelpText returns the computed help of the command and its subcommands.
func (c Cmd) HelpText() string {
	var b bytes.Buffer

	//	p := func(s ...interface{}) {
	//		fmt.Fprintln(&b)
	//		if len(s) > 0 {
	//			fmt.Fprintln(&b, s...)
	//		}
	//	}
	//	if c.LongHelp != "" {
	//		p(c.LongHelp)
	//	} else if c.Help != "" {
	//		p(c.Help)
	//	} else if c.Name != "" {
	//		p(c.Name, "has no help")
	//	}
	//	if c.hasSubcommand() {
	//		p("Commands:")
	//		w := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	//		for _, child := range c.Children() {
	//			fmt.Fprintf(w, "\t%s\t\t\t%s\n", child.Name, child.Help)
	//		}
	//		w.Flush()
	//		p()
	//	}

	var tmpl string

	// shell
	if c.Name == c.Root.Name {
		tmpl = `{{.HelpHelpText}}

Usage:
    {{.ShellName}} [options]
    
`
	} else {
		tmpl = `{{with (or .LongHelp .Help)}}{{.}}{{end}}

Usage:
    {{.ShellName}} {{.Name}} [options]

Flags:
{{ .Flags.FlagUsages }}`

	}

	c.template(tmpl, &b)

	if c.hasSubcommand() {
		fmt.Fprintln(&b, "Commands:")
		w := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
		for _, child := range c.Children() {
			fmt.Fprintf(w, "\t%s\t\t\t%s\n", child.Name, child.Help)
		}
		w.Flush()
	}
	return b.String()
}

func (c *Cmd) HelpHelpText() string {
	cmdNames := []string{"help"}
	helpCmd, _ := c.FindCmd(cmdNames)
	if helpCmd == nil {
		panic("Cano not found help command")
	}
	return helpCmd.Help
}

func (c *Cmd) template(tmpl string, buf *bytes.Buffer) {
	t := template.New("top")
	template.Must(t.Parse(tmpl))
	t.Execute(buf, &c)
}

// findChildCmd returns the subcommand with matching name or alias.
func (c *Cmd) findChildCmd(name string) *Cmd {
	// find perfect matches first
	if cmd, ok := c.children[name]; ok {
		return cmd
	}

	// find alias matching the name
	for _, cmd := range c.children {
		for _, alias := range cmd.Aliases {
			if alias == name {
				return cmd
			}
		}
	}

	return nil
}

// FindCmd finds the matching Cmd for args.
// It returns the Cmd and the remaining args.
func (c Cmd) FindCmd(args []string) (*Cmd, []string) {
	var cmd *Cmd
	for i, arg := range args {
		if cmd1 := c.findChildCmd(arg); cmd1 != nil {
			cmd = cmd1
			c = *cmd
			continue
		}
		return cmd, args[i:]
	}
	return cmd, nil
}

type cmdSorter []*Cmd

func (c cmdSorter) Len() int           { return len(c) }
func (c cmdSorter) Less(i, j int) bool { return c[i].Name < c[j].Name }
func (c cmdSorter) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

// TODO from cobra

// Flags returns the complete FlagSet that applies
// to this command (local and persistent declared here and by all parents).
func (c *Cmd) Flags() *flag.FlagSet {
	if c.flags == nil {
		c.flags = flag.NewFlagSet(c.Name, flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.flags.SetOutput(c.flagErrorBuf)
	}

	return c.flags
}

func (c *Cmd) ParseFlags(arguments []string) error {
	if c.flags != nil {
		return c.flags.Parse(arguments)
	}
	return nil

}
