package cli

import (
	"fmt"

	"github.com/pivotal-cf/servicescli/command"
)

func init() {
	command.CommandBuilders = append(command.CommandBuilders, InstanceCommandBuilder)
}

func InstanceCommandBuilder() ([]command.Command, error) {

	return []command.Command{
		{
			Command:          []string{"instance", "health"},
			ShortDescription: "short-binding",
			LongDescription:  "long-binding",
			Executor:         &InstanceCommand{},
		},
	}, nil
}

type InstanceCommand struct {
}

func (i *InstanceCommand) Execute(args []string) error {
	fmt.Println("instancing...")
	return nil
}
