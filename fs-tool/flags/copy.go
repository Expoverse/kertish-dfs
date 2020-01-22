package flags

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/freakmaxi/kertish-dfs/fs-tool/common"
	"github.com/freakmaxi/kertish-dfs/fs-tool/dfs"
	"github.com/freakmaxi/kertish-dfs/fs-tool/errors"
)

type copyCommand struct {
	headAddresses []string
	args          []string

	overwrite bool
	readRange *common.ReadRange
	source    string
	target    string
}

func NewCopy(headAddresses []string, args []string) execution {
	return &copyCommand{
		headAddresses: headAddresses,
		args:          args,
	}
}

func (c *copyCommand) Parse() error {
	for len(c.args) > 0 {
		arg := c.args[0]
		switch arg {
		case "-f":
			c.args = c.args[1:]
			c.overwrite = true
			continue
		case "-r":
			c.args = c.args[1:]
			if len(c.args) == 0 {
				return fmt.Errorf("range argument needs value")
			}
			readRange, err := common.NewReadRange(c.args[0])
			if err != nil {
				return err
			}
			c.args = c.args[1:]
			c.readRange = readRange
			continue
		case "-h":
			return errors.ErrShowUsage
		default:
			if strings.Index(arg, "-") == 0 {
				return fmt.Errorf("unsupported argument for cp command")
			}
		}
		break
	}

	if len(c.args) != 2 {
		return fmt.Errorf("cp command needs source and target parameters")
	}

	c.source = c.args[0]
	c.target = c.args[1]

	return nil
}

func (c *copyCommand) PrintUsage() {
	fmt.Println("  cp          Copy file or folder.")
	fmt.Println("              Ex: cp [arguments] [source] [target]          # Copy in dfs")
	fmt.Println("              Ex: cp [arguments] local:[source] [target]    # Copy from local to dfs")
	fmt.Println("              Ex: cp [arguments] [source] local:[target]    # Copy from dfs to local")
	fmt.Println()
	fmt.Println("arguments:")
	fmt.Println("  -f          overwrites the existent file / folder")
	fmt.Println("  -r value    copies only defined range of the file.")
	fmt.Println("              Ex: cp -r [byteBegins]->[byteEnds] [source] local:[target]")
	fmt.Println()
	fmt.Println("              WARNING: range works only from dfs to local copy operations")
	fmt.Println()
}

func (c *copyCommand) Execute() error {
	if strings.Index(c.source, local) == 0 {
		if err := c.localToRemote(); err != nil {
			return err
		}
		return nil
	}

	if strings.Index(c.target, local) == 0 {
		if err := c.remoteToLocal(); err != nil {
			return err
		}
		return nil
	}

	anim := common.NewAnimation("processing...")
	anim.Start()

	if err := dfs.Change(c.headAddresses, c.source, c.target, c.overwrite, true); err != nil {
		anim.Cancel()
		return err
	}
	anim.Stop()
	return nil
}

func (c *copyCommand) remoteToLocal() error {
	c.target = c.target[len(local):]

	info, err := os.Stat(c.target)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local file can't open")
	}

	if info != nil && info.IsDir() {
		_, sourceFileName := path.Split(c.source)
		c.target = path.Join(c.target, sourceFileName)

		info, err = os.Stat(c.target)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("local file can't open")
		}
	}

	if info != nil {
		if info.IsDir() {
			return fmt.Errorf("target %s is a path", c.target)
		}

		if !c.overwrite {
			fmt.Printf("File %s is already exists\n", c.target)
			fmt.Print("Do you want to overwrite? (y/N) ")

			reader := bufio.NewReader(os.Stdin)
			char, _, err := reader.ReadRune()
			if err != nil {
				return err
			}

			switch char {
			case 'Y', 'y':
			default:
				return nil
			}
		}
	}

	anim := common.NewAnimation("processing...")
	anim.Start()

	if err := dfs.Pull(c.headAddresses, c.source, c.target, c.readRange); err != nil {
		anim.Cancel()
		return err
	}
	anim.Stop()
	return nil
}

func (c *copyCommand) localToRemote() error {
	if strings.Index(c.target, local) == 0 {
		return fmt.Errorf("please use O/S native commands to copy/move files/folders between local locations")
	}

	c.source = c.source[len(local):]
	if len(c.source) == 0 {
		return fmt.Errorf("please specify the source")
	}

	anim := common.NewAnimation("processing...")
	anim.Start()

	if err := dfs.Put(c.headAddresses, c.source, c.target, c.overwrite); err != nil {
		anim.Cancel()
		return fmt.Errorf(err.Error())
	}
	anim.Stop()
	return nil
}