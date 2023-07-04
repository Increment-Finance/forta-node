package config

import (
	"path"
)

// Docker container names
var (
	DockerSupervisorImage = "forta-network/forta-node:latest"
	DockerUpdaterImage    = "forta-network/forta-node:latest"
	UseDockerImages       = "local"

	DockerSupervisorManagedContainers = 6
	DockerUpdaterContainerName        = "localhost"
	DockerSupervisorContainerName     = "localhost"
	DockerNatsContainerName           = "localhost"
	DockerIpfsContainerName           = "localhost"
	DockerScannerContainerName        = "localhost"
	DockerInspectorContainerName      = "localhost"
	DockerJSONRPCProxyContainerName   = "localhost"
	DockerPublicAPIProxyContainerName = "localhost"
	DockerJWTProviderContainerName    = "localhost"
	DockerStorageContainerName        = "localhost"

	DockerNetworkName = DockerScannerContainerName

	DefaultContainerFortaDirPath      = "/.forta"
	DefaultContainerConfigPath        = path.Join(DefaultContainerFortaDirPath, DefaultConfigFileName)
	DefaultContainerWrappedConfigPath = path.Join(DefaultContainerFortaDirPath, DefaultWrappedConfigFileName)
	DefaultContainerKeyDirPath        = path.Join(DefaultContainerFortaDirPath, DefaultKeysDirName)
)
