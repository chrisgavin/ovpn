package vpn

import (
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Connections map[string]*Connection `yaml:"connections"`
}

func LoadConfiguration() (*Configuration, error) {
	configurationPath, err := xdg.SearchConfigFile("ovpn/configuration.yaml")
	if err != nil {
		return nil, err
	}

	configurationBytes, err := ioutil.ReadFile(configurationPath)
	if err != nil {
		return nil, err
	}

	configuration := &Configuration{}
	err = yaml.Unmarshal(configurationBytes, configuration)
	if err != nil {
		return nil, err
	}

	for connectionName, connection := range configuration.Connections {
		connection.Name = connectionName
		if connection.WorkingDirectory == "" {
			connection.WorkingDirectory = filepath.Dir(connection.ConfigurationFiles[0])
		}
	}

	return configuration, nil
}

func (configuration *Configuration) Connection(name string) *Connection {
	return configuration.Connections[name]
}

func (configuration *Configuration) ConnectionsList() []*Connection {
	connectionsList := []*Connection{}
	for _, connection := range configuration.Connections {
		connectionsList = append(connectionsList, connection)
	}
	sort.SliceStable(connectionsList, func(i, j int) bool {
		return connectionsList[i].Name < connectionsList[j].Name
	})
	return connectionsList
}
