package abtest

import (
	"fmt"
	"github.com/phoenix-rec/abtest-sdk-go/abtest/log"
	"testing"
)

var testClient *ABClient

func init() {
	log.InitDefaultLogger()
	testClient = new(ABClient)
	hostport := "http://phoenix-rec-recsrv.srv.test.ixiaochuan.cn"
	//hostport := "http://phoenix-api.ixiaochuan.cn"
	testClient.Open(testClient, hostport, 10, 16)
}

func TestServerUnavailable1(t *testing.T) {
	strategyName, err := testClient.GetStrategyNamesByExpName("recommend_0703")
	if err != nil {
		fmt.Printf("get strategy name error: %v\n", err)
	}
	fmt.Printf("strategy name : %v\n", strategyName)
	expInfo, _ := testClient.GetStrategyName("3", "recommend_0703")
	fmt.Printf("exp info : %v\n", expInfo)

}

//func TestServerUnavailable2(t *testing.T) {
//	// projectId 项目id,
//	// hostport  默认实验config域名
//	// interval  实验config更新间隔，默认为10s
//	abtest.Open(16, "", 10)
//
//	boolVal := abtest.GetBool("123", "exp_name", "key_name", false)
//	//对照组false  实验组true
//	if boolVal {
//
//	} else {
//
//	}
//
//	intVal := abtest.GetInt64("123", "exp_name", "key_name", 1)
//	if intVal == 1 {
//
//	} else if intVal == 2 {
//
//	} else if intVal == 3 {
//
//	}
//}
