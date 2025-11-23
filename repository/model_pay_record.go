package repository

import (
	"errors"
	"slices"

	"github.com/spf13/cast"
	"github.com/suifengpiao14/sqlbuilder"
	"gitlab.huishoubao.com/gopackage/statemachine"
)

/*
CREATE TABLE `t_payment_record` (
  `Fid` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
    `Fpay_id` varchar(50) NOT NULL DEFAULT '' COMMENT '支付流水号',
  `Forder_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '订单Id',
  `Ftotal_amount` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '订单的总金额',
  `Fpay_amount` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '支付金额',
  `Fpay_agent` varchar(20)  NOT NULL DEFAULT '' COMMENT '支付机构 weixin:微信 alipay:支付宝',
  `Fstate` varchar(15) unsigned NOT NULL DEFAULT '1' COMMENT '支付状态 pending-未支付 paid-已支付,expired-已过期,failed-支付失败,closed-已关闭',
  `Fuser_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
 `Fclient_ip` varchar(255) NOT NULL DEFAULT '' COMMENT 'ip地址',
 `Fpay_url` varchar(255) NOT NULL DEFAULT '' COMMENT '支付链接',
 `Freturn_url` varchar(255) NOT NULL DEFAULT '' COMMENT '支付完成后的回调地址',
 `Fnotify_url` varchar(255) NOT NULL DEFAULT '' COMMENT '支付完成后的通知地址',
 `Fpay_param` varchar(255) NOT NULL DEFAULT '' COMMENT '支付参数',
  `Fpay_at` datetime NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '支付成功时间',
  `Fclosed_at`datetime NOT NULL DEFAULT '' COMMENT '关闭时间',
  `Fcreated_at` datetime NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '发起支付时间',
  PRIMARY KEY (`Fid`),
  KEY `key_order` (`Forder_id`),
  KEY `key_user` (`Fuser_id`),
  KEY `key_pay_agent` (`Fpay_agent`),
  KEY `key_state` (`Fstate`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='支付记录';
*/

type PayRecordModel struct {
	Id          int64  `gorm:"column:Fid" json:"id"`
	PayId       string `gorm:"column:Fpay_id" json:"payId"`
	OrderId     string `gorm:"column:Forder_id" json:"orderId"`
	OrderAmount int    `gorm:"column:Forder_amount" json:"orderAmount"`
	PayAmount   int    `gorm:"column:Fpay_amount" json:"payAmount"`
	PayAgent    string `gorm:"column:Fpay_agent" json:"payAgent"`
	State       string `gorm:"column:Fstate" json:"state"`
	UserId      string `gorm:"column:Fuser_id" json:"userId"`
	ClientIp    string `gorm:"column:Fclient_ip" json:"clientIp"`
	PayUrl      string `gorm:"column:Fpay_url" json:"payUrl"`
	Expire      int    `gorm:"column:Fexpire" json:"expire"`
	ReturnUrl   string `gorm:"column:Freturn_url" json:"returnUrl"`
	NotifyUrl   string `gorm:"column:Fnotify_url" json:"notifyUrl"`
	Remark      string `gorm:"column:Fremark" json:"remark"` // 支付结果备注信息，如支付失败原因等
	PayParam    string `gorm:"column:Fpay_param" json:"payParams"`
	CreatedAt   string `gorm:"column:Fcreated_at" json:"createdAt"`
	PayAt       string `gorm:"column:Fpaid_at" json:"paidAt"`
	ClosedAt    string `gorm:"column:Fclosed_at" json:"closedAt"`
	ExpiredAt   string `gorm:"column:Fexpired_at" json:"expiredAt"`
	FailedAt    string `gorm:"column:Ffailed_at" json:"failedAt"`
}

type PayRecordModels []PayRecordModel

func (ms PayRecordModels) TotalAmount() int {
	var total = 0
	for _, m := range ms {
		total += m.PayAmount
	}
	return total
}

var EffectStates = []string{PayOrderModel_state_pending.String(), PayOrderModel_state_paid.String()}

func (ms PayRecordModels) FilterByState(state ...string) (paidMs PayRecordModels) {
	for _, m := range ms {
		if slices.Contains(state, m.State) {
			paidMs = append(paidMs, m)
		}
	}
	return paidMs
}

func (ms PayRecordModels) FilterByStatePaid() (paidMs PayRecordModels) {
	return ms.FilterByState(PayOrderModel_state_paid.String())
}
func (ms PayRecordModels) FilterByStateEffect() (paidMs PayRecordModels) {
	return ms.FilterByState(EffectStates...)
}

func (ms PayRecordModels) First() (model *PayRecordModel, exists bool) {
	for _, m := range ms {
		return &m, true
	}
	return nil, false
}

func (ms PayRecordModels) GetOrderAmount() (orderAmount int) {
	effectRecords := ms.FilterByStateEffect()
	first, exists := effectRecords.First()
	if !exists {
		return 0
	}
	orderAmount = first.OrderAmount
	return orderAmount
}

func (ms PayRecordModels) FilterByStatePending() (paidMs PayRecordModels) {
	return ms.FilterByState(PayOrderModel_state_pending.String())
}

func (ms PayRecordModels) PaidMoney() (paidMoney int) {
	if len(ms) == 0 {
		return 0
	}
	firstOrderId := ms[0].OrderId
	for _, order := range ms {
		if order.OrderId != firstOrderId { // 确保只统计同一个支付单的金额
			err := errors.New("PayOrders.PaidMoney 方法只能用于同一个订单的支付单")
			panic(err)
		}
		if order.State == PayOrderModel_state_paid.String() {
			paidMoney += cast.ToInt(order.PayAmount)
		}
	}
	return paidMoney
}

func (ms PayRecordModels) IsOrderPayFinished() (payfinished bool) {
	if len(ms) == 0 {
		return true
	}
	orderAmount := ms.GetOrderAmount()
	payfinished = orderAmount <= ms.PaidMoney() // 所有支付单支付金额总和大于等于订单金额即为支付完成
	return payfinished
}

var table_pay_record = sqlbuilder.NewTableConfig("pay_record").AddColumns(
	sqlbuilder.NewColumn("Fid", sqlbuilder.GetField(NewId)),
	sqlbuilder.NewColumn("Fpay_id", sqlbuilder.GetField(NewPayId)),
	sqlbuilder.NewColumn("Forder_id", sqlbuilder.GetField(NewOrderId)),
	sqlbuilder.NewColumn("Forder_amount", sqlbuilder.GetField(NewOrderAmount)),
	sqlbuilder.NewColumn("Fpay_amount", sqlbuilder.GetField(NewPayAmount)),
	sqlbuilder.NewColumn("Fpay_agent", sqlbuilder.GetField(NewPayAgent)),
	sqlbuilder.NewColumn("Frecipient_account", sqlbuilder.GetField(NewRecipientAccount)),
	sqlbuilder.NewColumn("Frecipient_name", sqlbuilder.GetField(NewRecipientName)),
	sqlbuilder.NewColumn("Fpayment_account", sqlbuilder.GetField(NewPaymentAccount)),
	sqlbuilder.NewColumn("Fpayment_name", sqlbuilder.GetField(NewPaymentName)),
	sqlbuilder.NewColumn("Fstate", sqlbuilder.GetField(NewState)),
	sqlbuilder.NewColumn("Fuser_id", sqlbuilder.GetField(NewUserId)),
	sqlbuilder.NewColumn("Fclient_ip", sqlbuilder.GetField(NewClientIp)),
	sqlbuilder.NewColumn("Fpay_url", sqlbuilder.GetField(NewPayUrl)),
	sqlbuilder.NewColumn("Freturn_url", sqlbuilder.GetField(NewReturnUrl)),
	sqlbuilder.NewColumn("Fnotify_url", sqlbuilder.GetField(NewNotifyUrl)),
	sqlbuilder.NewColumn("Fpay_param", sqlbuilder.GetField(NewPayParam)),
	sqlbuilder.NewColumn("Fremark", sqlbuilder.GetField(NewRemark)),
	sqlbuilder.NewColumn("Fexpire", sqlbuilder.GetField(NewExpire)),
	sqlbuilder.NewColumn("Fcreated_at", sqlbuilder.GetField(NewCreatedAt)),
	sqlbuilder.NewColumn("Fpaid_at", sqlbuilder.GetField(NewPaidAt)),
	sqlbuilder.NewColumn("Fclosed_at", sqlbuilder.GetField(NewClosedAt)),
	sqlbuilder.NewColumn("Fexpired_at", sqlbuilder.GetField(NewExpiredAt)),
	sqlbuilder.NewColumn("Ffailed_at", sqlbuilder.GetField(NewFailedAt)),
).AddIndexs(
	sqlbuilder.Index{
		IsPrimary: true,
		ColumnNames: func(table sqlbuilder.TableConfig) (columnNames []string) {

			return []string{table.GetDBNameByFieldNameMust(sqlbuilder.GetFieldName(NewId))}
		},
	},
	sqlbuilder.Index{
		Unique: true,
		ColumnNames: func(table sqlbuilder.TableConfig) (columnNames []string) {
			return []string{table.GetDBNameByFieldNameMust(sqlbuilder.GetFieldName(NewPayId))}
		},
	},
	sqlbuilder.Index{
		ColumnNames: func(table sqlbuilder.TableConfig) (columnNames []string) {
			return []string{table.GetDBNameByFieldNameMust(sqlbuilder.GetFieldName(NewOrderId))}
		},
	},
).WithComment("收款记录表")

type PayRecordRepository struct {
	stateMachine statemachine.StateMachine
	repository   sqlbuilder.Repository
}

func NewPayRecordRepository(handler sqlbuilder.Handler) (repository PayRecordRepository) {
	tableConfig := table_pay_record.WithHandler(handler)
	stateMachine := repository.makeStateMachine(tableConfig)
	repository = PayRecordRepository{
		stateMachine: *stateMachine,
		repository:   sqlbuilder.NewRepository(tableConfig),
	}
	return repository
}

func (repo PayRecordRepository) GetStateMachine() statemachine.StateMachine {
	return repo.stateMachine
}
func (repo PayRecordRepository) GetAllPayRecordByConditon(whereFs sqlbuilder.Fields) (payRecordModels PayRecordModels, err error) {
	if len(whereFs) == 0 {
		err = errors.New("whereFs 不能为空")
		return nil, err
	}
	err = repo.repository.All(&payRecordModels, whereFs)
	if err != nil {
		return nil, err
	}
	return
}
func (repo PayRecordRepository) GetFirstPayRecordByConditon(whereFs sqlbuilder.Fields) (payRecordModel PayRecordModel, err error) {
	if len(whereFs) == 0 {
		err = errors.New("whereFs 不能为空")
		return payRecordModel, err
	}
	err = repo.repository.FirstMustExists(&payRecordModel, whereFs)
	return
}

func (payRecordRepository PayRecordRepository) makeStateMachine(tableConfig sqlbuilder.TableConfig) (stateMachine *statemachine.StateMachine) {
	fieldNamePayId := sqlbuilder.GetFieldName(NewPayId)
	colIdentity := tableConfig.Columns.GetByFieldNameMust(fieldNamePayId)
	fieldNameState := sqlbuilder.GetFieldName(NewState)
	colState := tableConfig.Columns.GetByFieldNameMust(fieldNameState)
	stateRepository := statemachine.NewStateRepository(
		tableConfig,
		statemachine.StateModelDbColumnRefer{
			Identity: colIdentity.DbName,
			State:    colState.DbName,
		},
	)
	stateMachine = newPayRecordStateMachine(stateRepository)
	return stateMachine
}

func newPayRecordStateMachine(stateRepository statemachine.StateRepository) *statemachine.StateMachine {
	var actions = statemachine.TransformEvents{
		{
			EventName: Action_pay_record_Pay,
			SrcStates: []string{
				PayOrderModel_state_pending.String(),
				PayOrderModel_state_failed.String(), // 支付失败可以继续支付（比如钱包金额不够、选中的优惠券过期等，充值后再支付，增加支付失败状态可以记录原因）
				PayOrderModel_state_paid.String(),   // 支持幂等
			},
			DstState: PayOrderModel_state_paid.String(),
		},
		{
			EventName: Action_pay_record_Expire, // 过期时需要先同步查询，看是否已经支付（比如消息异常导致未同步到数据）

			SrcStates: []string{
				PayOrderModel_state_pending.String(),
				PayOrderModel_state_expired.String(), // 支持幂等
			},
			DstState: PayOrderModel_state_expired.String(),
		},
		{
			EventName: Action_pay_record_Fail,
			SrcStates: []string{
				PayOrderModel_state_pending.String(),
				PayOrderModel_state_failed.String(), // 支持幂等
			},
			DstState: PayOrderModel_state_failed.String(),
		},
		{
			EventName: Action_pay_record_Close,
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
	Action_pay_record_Pay    = "actionPay"
	Action_pay_record_Expire = "actionExpire"
	Action_pay_record_Fail   = "actionFail"
	Action_pay_record_Close  = "actionClose"
)

type PayRecordCreateIn struct {
	PayId            string `json:"payId"`
	OrderId          string `json:"orderId"`
	OrderAmount      int    `json:"totalAmount"`
	PayAmount        int    `json:"payAmount"`
	PayAgent         string `json:"payAgent"`
	State            string `json:"state"`
	UserId           string `json:"userId"`
	ClientIp         string `json:"clientIp"`
	PayUrl           string `json:"payUrl"`
	ReturnUrl        string `json:"returnUrl"`
	NotifyUrl        string `json:"notifyUrl"`
	PayParam         string `json:"payParams"`
	Expire           int    `json:"expire"`
	Remark           string `json:"remark"`
	RecipientAccount string `json:"recipientAccount"`
	RecipientName    string `json:"recipientName"`
	PaymentAccount   string `json:"paymentAccount"`
	PaymentName      string `json:"paymentName"`
}

func (in PayRecordCreateIn) Fields() sqlbuilder.Fields {
	return sqlbuilder.Fields{
		NewPayId(in.PayId).SetRequired(true),
		NewOrderId(in.OrderId).SetRequired(true),
		NewOrderAmount(in.OrderAmount).SetRequired(true),
		NewPayAmount(in.PayAmount).SetRequired(true),
		NewPayAgent(in.PayAgent).SetRequired(true),
		NewState(in.State).SetRequired(true),
		NewUserId(in.UserId).SetRequired(true),
		NewClientIp(in.ClientIp),
		NewPayUrl(in.PayUrl),
		NewReturnUrl(in.ReturnUrl),
		NewNotifyUrl(in.NotifyUrl),
		NewPayParam(in.PayParam),
		NewExpire(in.Expire),
		NewRemark(in.Remark),
		NewRecipientAccount(in.RecipientAccount),
		NewRecipientName(in.RecipientName),
		NewPaymentAccount(in.PaymentAccount),
		NewPaymentName(in.PaymentName),
	}
}

func (repo PayRecordRepository) GetTable() sqlbuilder.TableConfig {
	return repo.repository.GetTable()
}

func (repo PayRecordRepository) Create(in PayRecordCreateIn) (err error) {
	err = repo.repository.Insert(in.Fields())
	if err != nil {
		return err
	}
	return nil
}

func (repo PayRecordRepository) WithTxHandler(txHandler sqlbuilder.Handler) PayRecordRepository {
	repo.repository = repo.repository.WithTxHandler(txHandler)
	return repo
}

func (repo PayRecordRepository) GetByPayId(payId string) (model PayRecordModel, exists bool, err error) {
	fs := sqlbuilder.Fields{
		NewPayId(payId).AppendWhereFn(sqlbuilder.ValueFnForward),
	}
	exists, err = repo.repository.First(&model, fs)
	if err != nil {
		return model, exists, err
	}
	return model, exists, nil
}

func (repo PayRecordRepository) GetByPayIdMust(payId string) (model PayRecordModel, err error) {
	model, exists, err := repo.GetByPayId(payId)
	if err != nil {
		return model, err
	}
	if !exists {
		err = sqlbuilder.ErrNotFound
		return model, err
	}
	return model, nil
}

func (repo PayRecordRepository) GetByOrderId(orderId string) (models PayRecordModels, err error) {
	fs := sqlbuilder.Fields{
		NewOrderId(orderId).AppendWhereFn(sqlbuilder.ValueFnForward),
	}
	err = repo.repository.All(&models, fs)
	if err != nil {
		return models, err
	}
	return models, nil
}
