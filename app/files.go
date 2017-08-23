package app

import (
	"encoding/json"

	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/os/config/cloudinit/config"
	"github.com/rancher/os/config/cloudinit/system"
)

func ProcessCloudInit(mdClient metadata.Client) error {
	self, err := mdClient.GetSelfService()
	if err != nil {
		return err
	}
	jsonBytes, err := json.Marshal(self.Metadata["cloud-init"])
	if err != nil {
		return err
	}

	cloudConfig, err := config.NewCloudConfig(string(jsonBytes))
	if err != nil {
		return err
	}

	for _, file := range cloudConfig.WriteFiles {
		_, err := system.WriteFile(&system.File{file}, "/")
		if err != nil {
			return err
		}
	}

	return nil
}
