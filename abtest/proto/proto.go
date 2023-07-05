package abtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/phoenix-rec/abtest-sdk-go/abtest/consts"
	"github.com/phoenix-rec/abtest-sdk-go/abtest/utils"
	"time"
)

type ExperimentType = int

const (
	RecExperiment ExperimentType = 1
	AllExperiment ExperimentType = 2
)

type ExperimentStatus = int

const (
	Enabled  ExperimentStatus = 1
	Disabled ExperimentStatus = -1
)

type Option func(*Options)

type Options struct {
	Hostport string
	Interval int
}

func WithHostport(s string) Option {
	return func(o *Options) {
		o.Hostport = s
	}
}

func WithInterval(i int) Option {
	return func(o *Options) {
		o.Interval = i
	}
}

type ExperimentInfo struct {
	ExpID          string           `json:"exp_id"`
	Name           string           `json:"name"`
	ExpType        ExperimentType   `json:"exp_type"`
	Ut             int64            `json:"ut"`
	PartitionCount uint64           `json:"partition_count"`
	Status         ExperimentStatus `json:"status"`
	Expire         int64            `json:"expire"`
	Version        int64            `json:"-"`

	// white_id => strategy_name
	WhiteMap map[string]string `json:"white_map,omitempty"`

	// strategy_name => config
	// DefaultStrategyName => default config
	ConfigMap    map[string]map[string]interface{} `json:"config_map"`
	ConfigRawMap map[string]string                 `json:"config_raw_map"`

	// strategy_name => partitions
	PartitionsMap map[string]string `json:"partitions_map"`

	// index => strategy_name
	StrategyNameTable []string `json:"-"`
}

func (i *ExperimentInfo) GetStrategy(id string) (strategyName string, err error) {
	defer func() {
		if err == nil && len(strategyName) == 0 {
			strategyName = consts.DefaultStrategyName
		}
	}()

	if i.WhiteMap == nil {
		err = fmt.Errorf("whiteMap is nil")
		return
	}

	strategyName, ok := i.WhiteMap[id]
	if ok {
		return
	}

	if i.PartitionCount > 0 && int(i.PartitionCount) == len(i.StrategyNameTable) {
		index := utils.HashIndex(i.ExpID, id, i.PartitionCount)
		strategyName = i.StrategyNameTable[index]
	}

	return
}

func (i *ExperimentInfo) GetConfig(id string) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})

	strategyName, err := i.GetStrategy(id)
	if err != nil {
		return
	}

	if i.ConfigMap == nil {
		err = fmt.Errorf("configMap is nil")
		return
	}

	for k, v := range i.ConfigMap[strategyName] {
		result[k] = v
	}

	return
}

func (i *ExperimentInfo) GetDefaultConfig() (result map[string]interface{}, err error) {
	result = make(map[string]interface{})

	if i.ConfigMap == nil {
		err = fmt.Errorf("configMap is nil")
		return
	}

	for k, v := range i.ConfigMap[consts.DefaultStrategyName] {
		result[k] = v
	}

	return
}

func (i *ExperimentInfo) GetRawConfigs() (data map[string][]byte, err error) {
	if i.ConfigRawMap == nil {
		err = fmt.Errorf("configRawMap is nil")
		return
	}

	data = make(map[string][]byte)
	for k, v := range i.ConfigRawMap {
		data[k] = []byte(v)
	}
	return
}

// GetRawConfig get strategy raw data by id
func (i *ExperimentInfo) GetRawConfig(id string) (data []byte, err error) {
	if i.ConfigRawMap == nil {
		err = fmt.Errorf("configRawMap is nil")
		return
	}

	strategyName, err := i.GetStrategy(id)
	if err != nil {
		return
	}
	data = []byte(i.ConfigRawMap[strategyName])

	return
}
func (i *ExperimentInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"exp_id":          i.ExpID,
		"name":            i.Name,
		"exp_type":        i.ExpType,
		"ut":              i.Ut,
		"partition_count": i.PartitionCount,
		"status":          i.Status,
		"expire":          i.Expire,
		"white_map":       i.WhiteMap,
		"config_map":      i.ConfigMap,
		"partitions_map":  i.PartitionsMap,
		"config_raw_map":  i.ConfigRawMap,
	})
}

func (i *ExperimentInfo) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	if err := dec.Decode(&m); err != nil {
		return err
	}

	expID, _ := m["exp_id"].(string)
	i.ExpID = expID

	name, _ := m["name"].(string)
	i.Name = name

	expType, _ := m["exp_type"].(json.Number).Int64()
	i.ExpType = int(expType)

	ut, _ := m["ut"].(json.Number).Int64()
	i.Ut = ut

	status, _ := m["status"].(json.Number).Int64()
	i.Status = int(status)

	expire, _ := m["expire"].(json.Number).Int64()
	i.Expire = int64(expire)

	i.Version = time.Now().UnixNano()

	partitionCount, _ := m["partition_count"].(json.Number).Int64()
	i.PartitionCount = uint64(partitionCount)

	i.WhiteMap = make(map[string]string)
	whiteMap, ok := m["white_map"].(map[string]interface{})
	if ok {
		for k, v := range whiteMap {
			i.WhiteMap[k] = v.(string)
		}
	}

	i.ConfigMap = make(map[string]map[string]interface{})
	configMap, ok := m["config_map"].(map[string]interface{})
	if ok {
		for k, v := range configMap {
			i.ConfigMap[k] = v.(map[string]interface{})
		}
	}

	i.ConfigRawMap = make(map[string]string)
	cfgRawMap := &struct {
		M map[string]string `json:"config_raw_map"`
	}{}
	if err := json.Unmarshal(data, cfgRawMap); err != nil {
		return err
	}
	if cfgRawMap.M != nil {
		i.ConfigRawMap = cfgRawMap.M
	}

	i.PartitionsMap = make(map[string]string)
	i.StrategyNameTable = make([]string, i.PartitionCount)
	partitionsMap, ok := m["partitions_map"].(map[string]interface{})
	if ok {
		pars := new(IntervalList)
		for strategyName, v := range partitionsMap {
			partitions := v.(string)
			i.PartitionsMap[strategyName] = partitions
			pars.Init(partitions, partitionCount)
			for _, index := range pars.Array() {
				i.StrategyNameTable[index] = strategyName
			}
		}
	}

	return nil
}

type ByVersionDesc []*ExperimentInfo

func (s ByVersionDesc) Len() int           { return len(s) }
func (s ByVersionDesc) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByVersionDesc) Less(i, j int) bool { return s[i].Version > s[j].Version }

type DataResp struct {
	Ret     int                `json:"ret"`
	Errcode int                `json:"errcode"`
	Msg     string             `json:"msg"`
	Data    *GetConfigListData `json:"data"`
}

type GetConfigListData struct {
	Time          int64                       `json:"time"`
	ConfigListMap map[int64][]*ExperimentInfo `json:"config_list_map"`
}
