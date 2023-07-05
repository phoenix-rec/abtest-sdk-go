package abtest

import (
	"fmt"
	"github.com/phoenix-rec/abtest-sdk-go/abtest/log"
	abtest "github.com/phoenix-rec/abtest-sdk-go/abtest/proto"
	"testing"
)

var testClient *ABClient

func init() {
	log.InitDefaultLogger()
	testClient = new(ABClient)
	hostport := "http://phoenix-rec-recsrv.srv.test.ixiaochuan.cn"
	//hostport := "http://phoenix-api.icocofun.com"

	testClient.Open(testClient, hostport, 10, 16)
}

func TestServerUnavailable1(t *testing.T) {
	strategyName, err := testClient.GetStrategyNamesByExpName("recommend_0703")
	if err != nil {
		fmt.Printf("get strategy name error: %v\n", err)
	}
	fmt.Printf("strategy name : %v\n", strategyName)
	expInfo, _ := testClient.GetRawConfigs("recommend_0703")
	fmt.Printf("exp info : %v\n", expInfo)

}

func TestServerUnavailable2(t *testing.T) {
	var opts []abtest.Option
	//opts = append(opts, abtest.WithHostport("http://phoenix-rec-recsrv.srv.test.ixiaochuan.cn")) //默认为phoenix系统实验config服务域名, 如有私有化部署域名，可配置
	//opts = append(opts, abtest.WithInterval(10)) //默认实验config更新间隔，默认为10s
	err := Open(16, opts...)
	if err != nil {
		fmt.Errorf("open fail, err: %v", err)
		return
	}

	boolVal := GetBool("123", "exp_name", "key_name", false)
	//对照组false  实验组true
	if boolVal {

	} else {

	}

	intVal := GetInt64("123", "exp_name", "key_name", 1)
	if intVal == 1 {

	} else if intVal == 2 {

	} else if intVal == 3 {

	}
}
