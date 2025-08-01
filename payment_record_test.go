package paymentrecord_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/paymentrecord"
	"github.com/suifengpiao14/paymentrecord/repository"
	"github.com/suifengpiao14/sqlbuilder"
)

// userName string, password string, host string, port int, database string
var mysqlDB = sqlbuilder.GormDBMakeMysql(sqlbuilder.DBConfig{
	UserName:     "test",
	Password:     "test",
	Host:         "127.0.01",
	Port:         3306,
	DatabaseName: "test",
}, nil)

var handler = sqlbuilder.NewGormHandler(mysqlDB)
var repo repository.PayRecordRepository
var payOrderService *paymentrecord.PayRecordService

func init() {
	sqlbuilder.CreateTableIfNotExists = true
	repo = repository.NewPayRecordRepository(handler)
	payOrderService = paymentrecord.NewPayRecordService(repo)
}

var payId = paymentrecord.PayIdGenerator()

// var payId = "202507291414062636"
var orderId = "o1545"

func TestCreateOrder(t *testing.T) {
	crateIn := paymentrecord.PayRecordCreateIn{
		PayId:       payId,
		OrderId:     orderId,
		PayAgent:    repository.PayingAgent_Wechat,
		OrderAmount: 5000,
		PayAmount:   1000,
		PayParam:    "",
		UserId:      "test_user_154",
		ClientIp:    "127.0.0.1",
	}
	err := payOrderService.Create(crateIn)
	require.NoError(t, err)
}

func TestPayOrder(t *testing.T) {
	payId := "202507291503311532"
	isOrderPaid, err := payOrderService.Pay(payId)
	require.NoError(t, err)
	fmt.Println(isOrderPaid)
}

func TestIsPaid(t *testing.T) {
	isPaid, err := payOrderService.IsPaid(orderId)
	require.NoError(t, err)
	fmt.Println(isPaid)
}

func TestCloseByPayId(t *testing.T) {
	payId := "202507291503311532"
	err := payOrderService.Close(payId)
	require.NoError(t, err)
}

func TestFailByPayId(t *testing.T) {
	payId := "202507291725134384"
	err := payOrderService.Fail(payId, "测试失败")
	require.NoError(t, err)
}
func TestExpireByPayId(t *testing.T) {
	payId := "202507291740221449"
	err := payOrderService.Expire(payId, "很长时间m没有支付，过期了")
	require.NoError(t, err)
}

func TestGetOrderPayInfo(t *testing.T) {
	payRecords, err := payOrderService.GetOrderPayInfo(orderId)
	require.NoError(t, err)
	fmt.Println(payRecords)
}

func TestGetByPayId(t *testing.T) {
	payId := "202507291740221449"
	payRecord, err := payOrderService.Get(payId)
	require.NoError(t, err)
	fmt.Println(payRecord)
}

func TestGetOrderRestPayRecordAmount(t *testing.T) {
	restPayAmount, err := payOrderService.GetOrderRestPayRecordAmount(orderId)
	require.NoError(t, err)
	fmt.Println(restPayAmount)
}
