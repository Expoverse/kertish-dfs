package flags

import (
	"fmt"
	"path"
)

const local = "local:"

type Command struct {
	filename    string
	args        []string
	headAddress string
	command     execution
}

func NewCommand(args []string) *Command {
	_, filename := path.Split(args[0])

	mrArgs := make([]string, 0)
	if 1 < len(args) {
		mrArgs = args[1:]
	}

	return &Command{
		filename:    filename,
		args:        mrArgs,
		headAddress: "localhost:4000",
	}
}

func (c *Command) printUsageHeader() {
	fmt.Println("2020-dfs usage: ")
	fmt.Println()
}

func (c *Command) printUsage() {
	c.printUsageHeader()
	fmt.Printf("   %s [options] command [arguments] parameters\n", c.filename)
	fmt.Println()
	fmt.Println("options:")
	fmt.Println("  --head-address   Points the end point of head node to work with. Default: localhost:4000")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  mkdir   Create folders.")
	fmt.Println("  ls      List files and folders.")
	fmt.Println("  cp      Copy file or folder.")
	fmt.Println("  mv      Move file or folder.")
	fmt.Println("  rm      Remove files and/or folders.")
	fmt.Println()
}

func (c *Command) Parse() bool {
	if len(c.args) == 0 {
		c.printUsage()
		return false
	}

	for i := 0; i < len(c.args); i++ {
		arg := c.args[i]

		switch arg {
		case "--head-address":
			if i+1 == len(c.args) {
				fmt.Println("--head-address requires value")
				fmt.Println()
				c.printUsage()
				return false
			}

			i++
			c.headAddress = c.args[i]
			continue
		}

		switch arg {
		case "mkdir", "ls", "cp", "mv", "rm":
			mrArgs := make([]string, 0)
			if i+1 < len(c.args) {
				mrArgs = c.args[i+1:]
			}

			var err error
			c.command, err = newExecution([]string{c.headAddress}, arg, mrArgs)
			if err != nil {
				fmt.Println(err.Error())
				fmt.Println()
				c.printUsage()
				return false
			}

			err = c.command.Parse()
			if err != nil {
				fmt.Println(err.Error())
				fmt.Println()
				c.printUsageHeader()
				c.command.PrintUsage()
				return false
			}

			return true
		}
	}

	c.printUsage()
	return false
}

func (c *Command) Execute() error {
	return c.command.Execute()
}
