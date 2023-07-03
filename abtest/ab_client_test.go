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
	testClient.Open(testClient, hostport, 10, 16)
}

func TestServerUnavailable(t *testing.T) {
	strategyName, err := testClient.GetStrategyNamesByExpName("recommend_0703")
	if err != nil {
		fmt.Printf("get strategy name error: %v\n", err)
	}
	fmt.Printf("strategy name : %v\n", strategyName)
	expInfo, _ := testClient.GetStrategyName("3", "recommend_0703")
	fmt.Printf("exp info : %v\n", expInfo)

}
