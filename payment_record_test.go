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
var mysqlDB = sqlbuilder.DB2Gorm(sqlbuilder.MakeDBHandler(sqlbuilder.DBConfig{
	UserName:     "root",
	Password:     "123456",
	Host:         "10.0.11.125",
	Port:         3306,
	DatabaseName: "test",
}), nil)

var handler = sqlbuilder.NewGormHandler(mysqlDB)
var payOrderService = paymentrecord.NewPayRecordService(handler)

func init() {
	sqlbuilder.CreateTableIfNotExists = true
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
	payId := "202508011738119472"
	in := paymentrecord.PayIn{
		PayId: payId,
	}
	isOrderPaid, err := payOrderService.Pay(in)
	require.NoError(t, err)
	fmt.Println(isOrderPaid)
}

func TestIsPaid(t *testing.T) {
	isPaid, err := payOrderService.IsPaid(orderId)
	require.NoError(t, err)
	fmt.Println(isPaid)
}

func TestCloseByPayId(t *testing.T) {
	payId := "202508011738066655"
	in := paymentrecord.CloseIn{
		PayId:  payId,
		Reason: "测试关闭",
	}
	err := payOrderService.Close(in)
	require.NoError(t, err)
}

func TestCloseByOrderId(t *testing.T) {
	in := paymentrecord.CloseByOrderIdIn{
		OrderId: orderId,
		Reason:  "测试关闭",
	}
	err := payOrderService.CloseByOrderId(in)
	require.NoError(t, err)
}

func TestFailByPayId(t *testing.T) {
	payId := "202507291725134384"
	in := paymentrecord.FailIn{
		PayId:  payId,
		Reason: "测试失败",
	}
	err := payOrderService.Fail(in)
	require.NoError(t, err)
}
func TestExpireByPayId(t *testing.T) {
	payId := "202507291740221449"
	in := paymentrecord.ExpireIn{
		PayId:  payId,
		Reason: "很长时间m没有支付，过期了",
	}
	err := payOrderService.Expire(in)
	require.NoError(t, err)
}

func TestGetOrderPayInfo(t *testing.T) {
	payRecords, err := payOrderService.GetOrderPayInfo(orderId)
	require.NoError(t, err)
	fmt.Println(payRecords)
}

func TestGetByPayId(t *testing.T) {
	payId := "202508011738119472"
	payRecord, err := payOrderService.Get(payId)
	require.NoError(t, err)
	fmt.Println(payRecord)
}

func TestGetOrderRestPayRecordAmount(t *testing.T) {
	restPayAmount, err := payOrderService.GetOrderRestPayRecordAmount(orderId)
	require.NoError(t, err)
	fmt.Println(restPayAmount)
}
