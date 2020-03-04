package echo

import (
	"github.com/labstack/echo"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
	"gopkg.in/yaml.v2"
)

//routeMapValue stores tags corresponding to routeMapKey
type routeToTagMapValue struct {
	OperationName string            `yaml:"operationName"`
	Metatags      map[string]string `yaml:"tags"`
}

//customTags stores tags corresponding to different types
type customTags struct {
	JwtClaims  []string          `yaml:"jwtClaims"`
	StaticTags map[string]string `yaml:"staticTags"`
}

// tracerConfig can be used to define config while creating
// tracer object
type tracerConfig struct {
	Cluster               string                        `yaml:"cluster"`
	Shard                 string                        `yaml:"shard"`
	Application           string                        `yaml:"application"`
	Service               string                        `yaml:"service"`
	Source                string                        `yaml:"source"`
	CustomApplicationTags customTags                    `yaml:"customApplicationTags"`
	RateSampler           uint64                        `yaml:"rateSampler"`
	DurationSampler       int64                         `yaml:"durationSampler"`
	RouteToTagsMap        map[string]routeToTagMapValue `yaml:"routesRegistration"`
}

// Config stores the middleware config
type Config struct {
	DirectCfg  *senders.DirectConfiguration
	ProxyCfg   *senders.ProxyConfiguration
	CfgFile    string
	RoutesFile string
	EchoWeb    *echo.Echo
}

//ReadConfigFromyaml reads the config file
//Constructs the tracer config from it.
//Populates the GlobalTracerConfig.
func readConfigFromYaml(configFilePath string) (*tracerConfig, error) {
	// Open our yamlFile
	configFile, err := readFromFile(configFilePath)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, err
	}

	// we unmarshal our byteArray which contains our
	// yamlFile's content into 'GlobalTracerConfig'
	yaml.Unmarshal([]byte(configFile), &globalTracerConfig)
	return globalTracerConfig, nil
}
