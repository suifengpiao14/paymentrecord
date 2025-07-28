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
	UserName:     "hjx",
	Password:     "123456",
	Host:         "129.204.136.31",
	Port:         3306,
	DatabaseName: "test",
}, nil)

func TestPayOrderService(t *testing.T) {
	cfg := paymentrecord.Config{}
	handler := sqlbuilder.NewGormHandler(mysqlDB)
	repo := repository.NewPayOrderRepository(handler)
	payOrderService := paymentrecord.NewPayOrderService(cfg, repo)
	crateIn := paymentrecord.PayOrderCreateIn{
		OrderId:     "o1545",
		PayAgent:    paymentrecord.PayingAgent_Wechat,
		OrderAmount: 5000,
		PayAmount:   3000,
		PayParam:    "",
		UserId:      "test_user_154",
		ClientIp:    "127.0.0.1",
	}
	out, err := payOrderService.Create(crateIn)
	require.NoError(t, err)
	fmt.Println(out)
}
