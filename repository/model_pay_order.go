package repository

import (
	"time"

	"github.com/suifengpiao14/sqlbuilder"
	"gitlab.huishoubao.com/gopackage/statemachine"
)

/*
CREATE TABLE `t_payment_order` (
  `Fid` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `Forder_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '订单Id',
  `Forder_amount` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '订单的总金额',
  `Fstate` varchar(15) unsigned NOT NULL DEFAULT '1' COMMENT '支付状态 pending-未支付 paid-已支付,closed-已关闭',
  `Fuser_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
  `Fpaid_at` datetime NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '支付成功时间',
  `Fclosed_at`datetime NOT NULL DEFAULT '' COMMENT '关闭时间',
  `Fcreated_at` datetime NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '发起支付时间',
  PRIMARY KEY (`Fid`),
  KEY `key_order` (`Forder_id`),
  KEY `key_user` (`Fuser_id`),
  KEY `key_state` (`Fstate`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='支付单表';
*/

var table_pay_order = sqlbuilder.NewTableConfig("pay_order").AddColumns(
	sqlbuilder.NewColumn("Fid", sqlbuilder.GetField(NewId)),
	sqlbuilder.NewColumn("Forder_id", sqlbuilder.GetField(NewOrderId)),
	sqlbuilder.NewColumn("Forder_amount", sqlbuilder.GetField(NewOrderAmount)),
	sqlbuilder.NewColumn("Fstate", sqlbuilder.GetField(NewState)),
	sqlbuilder.NewColumn("Fuser_id", sqlbuilder.GetField(NewUserId)),
	sqlbuilder.NewColumn("Fremark", sqlbuilder.GetField(NewRemark)),
	sqlbuilder.NewColumn("Fexpire", sqlbuilder.GetField(NewExpire)),
	sqlbuilder.NewColumn("Fcreated_at", sqlbuilder.GetField(NewCreatedAt)),
	sqlbuilder.NewColumn("Fpaid_at", sqlbuilder.GetField(NewPaidAt)),
	sqlbuilder.NewColumn("Fclosed_at", sqlbuilder.GetField(NewClosedAt)),
).AddIndexs(
	sqlbuilder.Index{
		IsPrimary: true,
		ColumnNames: func(tableColumns sqlbuilder.ColumnConfigs) (columnNames []string) {

			return []string{tableColumns.GetByFieldNameMust(sqlbuilder.GetFieldName(NewId)).DbName}
		},
	},
	sqlbuilder.Index{
		Unique: true,
		ColumnNames: func(tableColumns sqlbuilder.ColumnConfigs) (columnNames []string) {
			return []string{tableColumns.GetByFieldNameMust(sqlbuilder.GetFieldName(NewPayId)).DbName}
		},
	},
	sqlbuilder.Index{
		ColumnNames: func(tableColumns sqlbuilder.ColumnConfigs) (columnNames []string) {
			return []string{tableColumns.GetByFieldNameMust(sqlbuilder.GetFieldName(NewOrderId)).DbName}
		},
	},
).WithComment("收款单表")

type PayOrderModel struct {
	Id          int64  `gorm:"column:Fid" json:"id"`
	OrderId     string `gorm:"column:Forder_id" json:"orderId"`
	OrderAmount int    `gorm:"column:Forder_amount" json:"orderAmount"`
	State       string `gorm:"column:Fstate" json:"state"`
	UserId      string `gorm:"column:Fuser_id" json:"userId"`
	Remark      string `gorm:"column:Fremark" json:"remark"`
	Expire      string `gorm:"column:Fexpire" json:"expire"`
	CreatedAt   string `gorm:"column:Fcreated_at" json:"createdAt"`
	PaidAt      string `gorm:"column:Fpaid_at" json:"paidAt"`
	ClosedAt    string `gorm:"column:Fclosed_at" json:"closedAt"`
}

type PayOrderModels []PayOrderModel

type PayOrderRepository struct {
	stateMachine statemachine.StateMachine
	repository   sqlbuilder.Repository[PayOrderModel]
}

func NewPayOrderRepository(handler sqlbuilder.Handler) (repository PayOrderRepository) {
	tableConfig := table_pay_order.WithHandler(handler)
	stateMachine := repository.makeStateMachine(tableConfig)
	repository = PayOrderRepository{
		stateMachine: *stateMachine,
		repository:   sqlbuilder.NewRepository[PayOrderModel](tableConfig),
	}
	return repository
}

func (repo PayOrderRepository) GetStateMachine() statemachine.StateMachine {
	return repo.stateMachine
}

func (repo PayOrderRepository) makeStateMachine(tableConfig sqlbuilder.TableConfig) (stateMachine *statemachine.StateMachine) {
	fieldNameOrderId := sqlbuilder.GetFieldName(NewOrderId)
	colIdentity := tableConfig.Columns.GetByFieldNameMust(fieldNameOrderId)
	fieldNameState := sqlbuilder.GetFieldName(NewState)
	colState := tableConfig.Columns.GetByFieldNameMust(fieldNameState)
	stateRepository := statemachine.NewStateRepository(
		tableConfig,
		statemachine.StateModelDbColumnRefer{
			Identity: colIdentity.DbName,
			State:    colState.DbName,
		},
	)
	stateMachine = newPayOrderStateMachine(stateRepository)
	return stateMachine
}

func newPayOrderStateMachine(stateRepository statemachine.StateRepository) *statemachine.StateMachine {
	var actions = statemachine.Actions{
		{
			ActionName: Action_pay_order_Pay,
			SrcStates: []string{
				PayOrderModel_state_pending.String(),
				PayOrderModel_state_paid.String(), // 支持幂等
			},
			DstState: PayOrderModel_state_paid.String(),
		},
		{
			ActionName: Action_pay_order_Close,
			SrcStates: []string{
				PayOrderModel_state_pending.String(),
				PayOrderModel_state_closed.String(), // 支持幂等
			},
			DstState: PayOrderModel_state_closed.String(),
		},
	}
	stateMachine := statemachine.NewStateMachine(actions, stateRepository)
	return stateMachine
}

const (
	Action_pay_order_Pay   = "actionPay"
	Action_pay_order_Close = "actionClose"
)

type PayOrderSetIn struct {
	OrderId     string `json:"orderId"`
	OrderAmount int    `json:"orderAmount"`
	UserId      string `json:"userId"`
	Remark      string `json:"remark"`
}

func (in PayOrderSetIn) Fields() sqlbuilder.Fields {
	return sqlbuilder.Fields{
		NewOrderId(in.OrderId).SetRequired(true).AppendWhereFn(sqlbuilder.ValueFnForward),
		NewOrderAmount(in.OrderAmount).SetRequired(true).SetMinimum(1),
		NewUserId(in.UserId),
		NewRemark(in.Remark),
		NewCreatedAt(time.Now().Format(time.DateTime)),
		NewState(PayOrderModel_state_pending.String()),
	}
}

func (repo PayOrderRepository) Set(in PayOrderSetIn) (err error) {
	_, _, _, err = repo.repository.Set(in.Fields(), func(p *sqlbuilder.SetParam) {
		p.WithPolicy(sqlbuilder.SetPolicy_only_Insert)
	})
	if err != nil {
		return err
	}
	return nil
}
