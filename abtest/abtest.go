package abtest

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/phoenix-rec/abtest-sdk-go/abtest/consts"
	logger "github.com/phoenix-rec/abtest-sdk-go/abtest/log"
)

var client ConfigReader
var _json = jsoniter.ConfigCompatibleWithStandardLibrary

// Open the A/B client which holds all the A/B configs in memory
// and incrementally synchronizes data from server at intervals specified in conf.
func Open(projectId int64, hostport string, interval int) (err error) {
	if projectId == 0 {
		err = ErrClientSettingErr
		return
	}

	if hostport == "" {
		hostport = consts.DefaultAbConfigHost
	}

	if interval <= 0 {
		interval = consts.DefaultIntervalInSecond
	}
	logger.InitDefaultLogger()
	client = new(ABClient)
	client.Open(client, hostport, interval, projectId)
	return
}

func Close() {
	if client == nil {
		return
	}

	client.Close()
	return
}

func GetConfig(id string) (config map[string]interface{}) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return make(map[string]interface{})
	}

	config, err := client.GetConfig(id)
	if err != nil {
		client.TrackError("GetConfig", id, "", "", err)
	}

	return
}

func GetExperiments(id string) (experiments map[string]map[string]interface{}) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return make(map[string]map[string]interface{})
	}

	experiments, err := client.GetExperiments(id)
	if err != nil {
		client.TrackError("GetExperiments", id, "", "", err)
	}

	return
}

func GetExperiment(id, expName string) (exp map[string]interface{}) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return make(map[string]interface{})
	}

	exp, err := client.GetExperiment(id, expName)
	if err != nil {
		client.TrackError("GetExperiment", id, expName, "", err)
	}

	return
}

func GetStrategyNamesByExpName(expName string) (strategies []string, err error) {
	if client == nil {
		err = ErrClientUninitialized
		return
	}
	return client.GetStrategyNamesByExpName(expName)
}

func GetStrategyName(id, expName string) (strategy string, err error) {
	if client == nil {
		err = ErrClientUninitialized
		return
	}
	return client.GetStrategyName(id, expName)
}

func GetBool(id, expName, keyName string, defaultValue bool) (val bool) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return defaultValue
	}

	if err := client.GetKey(id, expName, keyName, &val); err != nil {
		client.TrackError("GetBool", id, expName, keyName, err)
		val = defaultValue
	}

	return
}

func GetString(id, expName, keyName, defaultValue string) (val string) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return defaultValue
	}

	if err := client.GetKey(id, expName, keyName, &val); err != nil {
		client.TrackError("GetString", id, expName, keyName, err)
		val = defaultValue
	}

	return
}

func GetInt64(id, expName, keyName string, defaultValue int64) (val int64) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return defaultValue
	}

	if err := client.GetKey(id, expName, keyName, &val); err != nil {
		client.TrackError("GetInt64", id, expName, keyName, err)
		val = defaultValue
	}

	return
}

func GetFloat64(id, expName, keyName string, defaultValue float64) (val float64) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return float64(defaultValue)
	}

	if err := client.GetKey(id, expName, keyName, &val); err != nil {
		client.TrackError("GetFloat64", id, expName, keyName, err)
		val = float64(defaultValue)
	}

	return
}

func GetStringSlice(id, expName, keyName string, defaultValue []string) (val []string) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return defaultValue
	}

	if err := client.GetKey(id, expName, keyName, &val); err != nil {
		client.TrackError("GetStringSlice", id, expName, keyName, err)
		val = defaultValue
	}

	return
}

func GetInt64Slice(id, expName, keyName string, defaultValue []int64) (val []int64) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return defaultValue
	}

	if err := client.GetKey(id, expName, keyName, &val); err != nil {
		client.TrackError("GetInt64Slice", id, expName, keyName, err)
		val = defaultValue
	}

	return
}

func GetMap(id, expName, keyName string, defaultValue map[string]interface{}) (val map[string]interface{}) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return defaultValue
	}

	if err := client.GetKey(id, expName, keyName, &val); err != nil {
		client.TrackError("GetMap", id, expName, keyName, err)
		val = defaultValue
	}

	return
}

func GetRawConfigs(expName string) (data map[string][]byte, err error) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return nil, ErrClientUninitialized
	}

	data, err = client.GetRawConfigs(expName)
	if err != nil {
		client.TrackErrorNew("GetRawConfigs", "", expName, "", err)
		return
	}
	return
}

func GetRawConfig(id, expName string) (data []byte, err error) {
	if client == nil {
		logger.Error(ErrClientUninitialized)
		return nil, ErrClientUninitialized
	}

	data, err = client.GetRawConfig(id, expName)
	if err != nil {
		client.TrackErrorNew("GetRawConfigs", "", "", expName, err)
		return
	}
	return
}
