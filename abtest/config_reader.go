package abtest

type ConfigReader interface {
	Open(c ConfigReader, hostport string, interval int, projectId int64) ConfigReader
	Update() error
	Close()

	GetConfig(id string) (map[string]interface{}, error)
	GetRawConfigs(expName string) (map[string][]byte, error)
	GetRawConfig(id, expName string) ([]byte, error)
	GetExperiments(id string) (map[string]map[string]interface{}, error)
	GetExperiment(id, expName string) (map[string]interface{}, error)
	GetKey(id, expName, keyName string, result interface{}) error
	GetStrategyName(id, expName string) (string, error)
	GetStrategyNamesByExpName(expName string) (strategies []string, err error)
	TrackError(f, id, expName, keyName string, err error)
	TrackErrorNew(f, id, expName, keyName string, err error)
}
