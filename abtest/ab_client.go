package abtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/phoenix-rec/abtest-sdk-go/abtest/consts"
	logger "github.com/phoenix-rec/abtest-sdk-go/abtest/log"
	abtest "github.com/phoenix-rec/abtest-sdk-go/abtest/proto"
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type ABClient struct {
	abAdapter *http.Client

	typeMask uint

	infoMap atomic.Value // map[int64]map[string]*abtest.ExperimentInfo

	interval               int          // in second
	ticker                 *time.Ticker // init at ABClient.Open
	ut                     int64        // unix nano as the version of local A/B config
	serverUnavailableTicks int
	ticksToSkip            int

	m         sync.Mutex
	closeChan chan bool
	errCount  uint64

	hostport  string
	projectId int64
}

func (c *ABClient) Open(r ConfigReader, hostport string, interval int, projectId int64) ConfigReader {
	if c == nil {
		c = new(ABClient)
	}

	if c.isRunning() {
		return c
	}

	c.closeChan = make(chan bool)
	c.abAdapter = &http.Client{}
	c.projectId = projectId
	c.hostport = hostport
	c.interval = interval

	logger.TraceF("Open hostport: %v, interval: %v", c.hostport, c.interval)

	if c.ticker != nil {
		c.ticker.Stop()
	}
	c.ticker = time.NewTicker(time.Duration(interval) * time.Second)

	c.infoMap.Store(make(map[int64]map[string]*abtest.ExperimentInfo))
	err := r.Update()
	if err != nil {
		logger.ErrorF("ABClient init err: %v", err)
		logger.Error(ErrAllDefault)
	}

	go func() {
		for c.isRunning() {
			c.work(r)
		}
	}()

	return c
}

func (c *ABClient) isRunning() bool {
	return c != nil && c.closeChan != nil
}

// work is not safe for concurrent use.
func (c *ABClient) work(r ConfigReader) {
	defer func() {
		if err := recover(); err != nil {
			logger.ErrorF("worker err: %v", err)
		}
	}()

	select {
	case <-c.closeChan:
		c.ticker.Stop()
		return
	case <-c.ticker.C:
		if c.ticksToSkip > 0 {
			c.ticksToSkip -= 1
			return
		}

		err := r.Update()
		if err != nil {
			c.serverUnavailableTicks = (c.serverUnavailableTicks << 1) + 1
			c.ticksToSkip = c.serverUnavailableTicks

			if c.ut == 0 {
				logger.ErrorF("Update err: %v", err)
				logger.Error(ErrAllDefault)
			} else {
				logger.Warn("Update err: %v", err)
				logger.Warn("A/B server (%s) has been unavailable for %d seconds, local A/B config may be out of date. Retry after %d seconds.", c.hostport, c.serverUnavailableTicks*c.interval, c.ticksToSkip*c.interval)
			}

		} else {
			c.serverUnavailableTicks, c.ticksToSkip = 0, 0
		}
	}

	return
}

func (c *ABClient) Update() (err error) {
	remoteInfoMap, err := c.remoteInfoMap()
	if err != nil {
		return
	}

	if len(remoteInfoMap) == 0 {
		// Already up to date.
		return
	}

	logger.TraceF("%d project experiment(s) to update", len(remoteInfoMap))

	atomic.StoreUint64(&c.errCount, 0)
	c.m.Lock()
	defer c.m.Unlock()

	localInfoMap, _ := c.infoMap.Load().(map[int64]map[string]*abtest.ExperimentInfo)

	for projectId, exp := range localInfoMap {
		if infoMap, ok := remoteInfoMap[projectId]; !ok {
			remoteInfoMap[projectId] = exp
		} else {
			for expName, info := range exp {
				if _, ok1 := infoMap[expName]; !ok1 {
					infoMap[expName] = info
				}
			}
			remoteInfoMap[projectId] = infoMap
		}
	}

	c.infoMap.Store(remoteInfoMap)

	return
}

func (c *ABClient) remoteInfoMap() (projectInfoMap map[int64]map[string]*abtest.ExperimentInfo, err error) {
	projectInfoMap = make(map[int64]map[string]*abtest.ExperimentInfo)

	param := map[string]interface{}{
		"time": c.ut,
	}

	resp, err := c.getConfigList(param)
	if err != nil {
		return
	}
	if resp.Ret != 1 || resp.Data == nil {
		err = fmt.Errorf("unexpected resp: %v", resp)
		return
	}

	for projectId, expList := range resp.Data.ConfigListMap {
		infoMap := make(map[string]*abtest.ExperimentInfo)
		for _, info := range expList {
			infoMap[info.Name] = info
		}
		projectInfoMap[projectId] = infoMap
	}
	c.ut = resp.Data.Time

	return
}

func (c *ABClient) Close() {
	if !c.isRunning() {
		return
	}

	c.closeChan <- true
	c.closeChan = nil

	return
}

func (c *ABClient) GetConfig(id string) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	if !c.isRunning() {
		err = ErrClientStopped
		return
	}

	infoMap, ok := c.infoMap.Load().(map[int64]map[string]*abtest.ExperimentInfo)
	if !ok || infoMap == nil {
		err = ErrClientUninitialized
		return
	}

	expInfoMap, ok := infoMap[c.projectId]
	if !ok {
		err = ErrProjectNotFound
		return
	}

	for expName, info := range expInfoMap {
		if info.Status == abtest.Disabled {
			continue
		}

		expConfig, expErr := c.GetExperiment(id, expName)
		if expErr != nil {
			err = expErr
			continue
		}

		for k, v := range expConfig {
			result[k] = v
		}
	}

	return
}

func (c *ABClient) GetExperiments(id string) (experiments map[string]map[string]interface{}, err error) {
	experiments = make(map[string]map[string]interface{})
	if !c.isRunning() {
		err = ErrClientStopped
		return
	}

	infoMap, ok := c.infoMap.Load().(map[int64]map[string]*abtest.ExperimentInfo)
	if !ok || infoMap == nil {
		err = ErrClientUninitialized
		return
	}

	expInfoMap, ok := infoMap[c.projectId]
	if !ok {
		err = ErrProjectNotFound
		return
	}

	for expName, info := range expInfoMap {
		if info.Status == abtest.Disabled {
			continue
		}
		expConfig, expErr := info.GetConfig(id)
		if expErr != nil {
			err = expErr
			continue
		}

		experiments[expName] = expConfig
	}

	return
}

func (c *ABClient) GetExperiment(id string, expName string) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	if !c.isRunning() {
		err = ErrClientStopped
		return
	}

	infoMap, ok := c.infoMap.Load().(map[int64]map[string]*abtest.ExperimentInfo)
	if !ok || infoMap == nil {
		err = ErrClientUninitialized
		return
	}

	expinfoMap, ok := infoMap[c.projectId]
	if !ok {
		err = ErrProjectNotFound
		return
	}

	info, ok := expinfoMap[expName]
	if !ok {
		err = ErrExperimentNotFound
		return
	}

	if info.Status == abtest.Disabled {
		err = ErrExperimentDisabled
		return
	}

	return info.GetConfig(id)
}

func (c *ABClient) GetKey(id, expName, keyName string, result interface{}) (err error) {
	if !c.isRunning() {
		err = ErrClientStopped
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() != reflect.Ptr {
		err = fmt.Errorf("result argument must be a pointer")
		return
	}

	config, err := c.GetExperiment(id, expName)
	if err != nil {
		return
	}

	val, ok := config[keyName]
	if !ok {
		err = ErrKeyNotFound
		return
	}

	valType := reflect.TypeOf(val)
	if valType == nil {
		return
	}

	if reflect.PtrTo(valType) == reflect.TypeOf(result) {
		resultValue.Elem().Set(reflect.ValueOf(val))
		return
	}

	bs, err := json.Marshal(val)
	if err != nil {
		return
	}

	err = json.Unmarshal(bs, result)
	if err != nil {
		elem := reflect.ValueOf(result).Elem()
		elem.Set(reflect.Zero(elem.Type()))
		return
	}

	return
}

func (c *ABClient) GetRawConfigs(expName string) (result map[string][]byte, err error) {
	if !c.isRunning() {
		err = ErrClientStopped
		return
	}

	infoMap := c.infoMap.Load().(map[int64]map[string]*abtest.ExperimentInfo)

	expInfoMap, ok := infoMap[c.projectId]
	if !ok {
		err = ErrProjectNotFound
		return
	}

	info, ok := expInfoMap[expName]
	if !ok {
		err = ErrExperimentNotFound
		return
	}

	if info.Status == abtest.Disabled {
		err = ErrExperimentDisabled
		return
	}

	return info.GetRawConfigs()
}

func (c *ABClient) GetRawConfig(id, expName string) (data []byte, err error) {
	if !c.isRunning() {
		err = ErrClientStopped
		return
	}

	infoMap := c.infoMap.Load().(map[int64]map[string]*abtest.ExperimentInfo)

	expInfoMap, ok := infoMap[c.projectId]
	if !ok {
		err = ErrProjectNotFound
		return
	}

	info, ok := expInfoMap[expName]
	if !ok {
		err = ErrExperimentNotFound
		return
	}

	if info.Status == abtest.Disabled {
		err = ErrExperimentDisabled
		return
	}

	return info.GetRawConfig(id)
}

func (c *ABClient) GetStrategyNamesByExpName(expName string) (strategies []string, err error) {
	if !c.isRunning() {
		err = ErrClientStopped
		return
	}

	infoMap, ok := c.infoMap.Load().(map[int64]map[string]*abtest.ExperimentInfo)
	if !ok || infoMap == nil {
		err = ErrClientUninitialized
		return
	}

	expinfoMap, ok := infoMap[c.projectId]
	if !ok {
		err = ErrProjectNotFound
		return
	}

	info, ok := expinfoMap[expName]
	if !ok {
		err = ErrExperimentNotFound
		return
	}

	m := make(map[string]bool)
	for _, strategy := range info.StrategyNameTable {
		m[strategy] = true
	}
	for strategy, _ := range m {
		strategies = append(strategies, strategy)
	}
	return
}

func (c *ABClient) GetStrategyName(userId, expName string) (strategyName string, err error) {
	if !c.isRunning() {
		err = ErrClientStopped
		return
	}

	infoMap, ok := c.infoMap.Load().(map[int64]map[string]*abtest.ExperimentInfo)
	if !ok || infoMap == nil {
		err = ErrClientUninitialized
		return
	}

	expinfoMap, ok := infoMap[c.projectId]
	if !ok {
		err = ErrProjectNotFound
		return
	}

	info, ok := expinfoMap[expName]
	if !ok {
		err = ErrExperimentNotFound
		return
	}

	return info.GetStrategy(userId)
}

func (c *ABClient) TrackError(f, id, expName, keyName string, err error) {
	if err == ErrExperimentNotMatch {
		return
	}

	errCount := atomic.AddUint64(&c.errCount, 1)
	if errCount&(errCount-1) == 0 { // is power of two
		logger.Warn("%s err: %v, id: %s, expName: %s, keyName: %s", f, err, id, expName, keyName)
	}

	return
}

func (c *ABClient) TrackErrorNew(f, id, expName, keyName string, err error) {
	if err == ErrExperimentNotMatch {
		return
	}

	errCount := atomic.AddUint64(&c.errCount, 1)
	if errCount&(errCount-1) == 0 { // is power of two
		logger.Warn("%s err: %v, id: %s, expName: %s, keyName: %s, typ: %v", f, err, id, expName, keyName)
	}

	return
}

func (c *ABClient) apiRequest(url string, httpBody []byte) (resp []byte, err error) {
	request, err := http.NewRequest("POST", url, bytes.NewReader(httpBody))
	if err != nil {
		return
	}

	// 发起请求
	httpResp, err := c.abAdapter.Do(request)
	if err != nil {
		return
	}

	defer httpResp.Body.Close()
	resp, err = ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return
	}

	return
}

func (c *ABClient) getConfigList(param map[string]interface{}) (resp *abtest.DataResp, err error) {
	resp = new(abtest.DataResp)
	url := c.hostport + consts.DefaultAbApiPath
	data, err := json.Marshal(param)
	if err != nil {
		return
	}
	respBody, err := c.apiRequest(url, data)
	json.Unmarshal(respBody, &resp)
	return
}
