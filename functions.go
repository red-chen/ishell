package ishell

import (
	"os"
)

func exitFunc(c *Context) {
	c.Stop()
}

func helpFunc(c *Context) {
	c.Println(c.HelpText())
}

func clearFunc(c *Context) {
	err := c.ClearScreen()
	if err != nil {
		c.Err(err)
	}
}

func addDefaultFuncs(s *Shell) {
	s.AddCmd(&Cmd{
		Name: "exit",
		Help: "Exit the program",
		Func: exitFunc,
	})
	s.AddCmd(&Cmd{
		Name: "quit",
		Help: "Quit the program",
		Func: exitFunc,
	})
	s.AddCmd(&Cmd{
		Name: "help",
		Help: "Display help",
		Func: helpFunc,
	})
	s.AddCmd(&Cmd{
		Name: "clear",
		Help: "Clear the screen",
		Func: clearFunc,
	})
	s.Interrupt(interruptFunc)
}

func interruptFunc(c *Context, count int, line string) {
	if count >= 2 {
		c.Println("Interrupted")
		os.Exit(1)
	}
	c.Println("Input Ctrl-c once more to exit")
}
