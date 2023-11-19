package main

import (
	server "com.github/salpreh/devserver/pkg/servers"
	"com.github/salpreh/devserver/pkg/servers/contracts"
	collectionutils "com.github/salpreh/devserver/pkg/utils"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

func CreateContractCmd() *cobra.Command {
	contractTypeFlag := v3Type
	contractCmd := &cobra.Command{
		Use:   "contract [contractPath] [outputPath]",
		Short: "Read openapi contract",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			config := contracts.LoadContractMockConfig(args[0])
			importedConfig, err := server.NewImportedMockConfig(config)
			if err != nil {
				log.Panicf("Something went wrong importing config: %v", err)
			}
			if err := collectionutils.ExportToFileAsJson(args[1], importedConfig); err != nil {
				log.Panicf("Unable to export config to file %s: %v", args[1], err)
			}
		},
	}

	contractCmd.Flags().VarP(&contractTypeFlag, "contractType", "t", fmt.Sprintf("Contract type. Allowed: %s, %s", v2Type, v3Type))

	return contractCmd
}

type contractType string

const (
	v2Type contractType = "v2"
	v3Type contractType = "v3"
)

func (t *contractType) String() string {
	return string(*t)
}

func (t *contractType) Set(v string) error {
	newType := contractType(v)
	switch newType {
	case v2Type, v3Type:
		*t = newType
		return nil
	default:
		return errors.New(fmt.Sprintf("Unknown contract type: %s", v))
	}
}

func (t *contractType) Type() string {
	return "contractType"
}
