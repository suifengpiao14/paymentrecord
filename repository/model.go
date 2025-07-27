package repository

import (
	"time"

	"github.com/suifengpiao14/sqlbuilder"
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

type PayModelOrderModel struct {
	Id          int64  `gorm:"column:Fid" json:"id"`
	PayId       string `gorm:"column:Fpay_id" json:"payId"`
	OrderId     string `gorm:"column:Forder_id" json:"orderId"`
	OrderAmount int    `gorm:"column:Forder_amount" json:"orderAmount"`
	PayAmount   int    `gorm:"column:Fpay_amount" json:"payAmount"`
	PayAgent    string `gorm:"column:Fpay_agent" json:"payAgent"`
	State       string `gorm:"column:Fstate" json:"state"`
	UserId      int64  `gorm:"column:Fuser_id" json:"userId"`
	ClientIp    string `gorm:"column:Fclient_ip" json:"clientIp"`
	PayUrl      string `gorm:"column:Fpay_url" json:"payUrl"`
	Expire      int    `gorm:"column:Fexpire" json:"expire"`
	ReturnUrl   string `gorm:"column:Freturn_url" json:"returnUrl"`
	NotifyUrl   string `gorm:"column:Fnotify_url" json:"notifyUrl"`
	PayParam    string `gorm:"column:Fpay_param" json:"payParams"`
	PayAt       string `gorm:"column:Fpay_at" json:"payAt"`
	ClosedAt    string `gorm:"column:Fclosed_at" json:"closedAt"`
	CreatedAt   string `gorm:"column:Fcreated_at" json:"createdAt"`
}

type PayOrderModels []PayModelOrderModel

func (ms PayOrderModels) TotalAmount() int {
	var total = 0
	for _, m := range ms {
		total += m.PayAmount
	}
	return total
}

var table_pay_order = sqlbuilder.NewTableConfig("pay_record").AddColumns(
	sqlbuilder.NewColumn("Fid", sqlbuilder.GetField(NewId)),
	sqlbuilder.NewColumn("Fpay_id", sqlbuilder.GetField(NewPayId)),
	sqlbuilder.NewColumn("Forder_id", sqlbuilder.GetField(NewOrderId)),
	sqlbuilder.NewColumn("Ftotal_amount", sqlbuilder.GetField(NewTotalAmount)),
	sqlbuilder.NewColumn("Fpay_amount", sqlbuilder.GetField(NewPayAmount)),
	sqlbuilder.NewColumn("Fpay_agent", sqlbuilder.GetField(NewPayAgent)),
	sqlbuilder.NewColumn("Fstate", sqlbuilder.GetField(NewState)),
	sqlbuilder.NewColumn("Fuser_id", sqlbuilder.GetField(NewUserId)),
	sqlbuilder.NewColumn("Fclient_ip", sqlbuilder.GetField(NewClientIp)),
	sqlbuilder.NewColumn("Fpay_url", sqlbuilder.GetField(NewPayUrl)),
	sqlbuilder.NewColumn("Freturn_url", sqlbuilder.GetField(NewReturnUrl)),
	sqlbuilder.NewColumn("Fnotify_url", sqlbuilder.GetField(NewNotifyUrl)),
	sqlbuilder.NewColumn("Fpay_param", sqlbuilder.GetField(NewPayParam)),
	sqlbuilder.NewColumn("Fpay_at", sqlbuilder.GetField(NewPayAt)),
	sqlbuilder.NewColumn("Fclosed_at", sqlbuilder.GetField(NewClosedAt)),
	sqlbuilder.NewColumn("Fcreated_at", sqlbuilder.GetField(NewCreatedAt)),
).AddIndexs(
	sqlbuilder.Index{
		IsPrimary: true,
		ColumnNames: func(tableColumns sqlbuilder.ColumnConfigs) (columnNames []string) {
			return []string{"Fid"}
		},
	},
	sqlbuilder.Index{
		Unique: true,
		ColumnNames: func(tableColumns sqlbuilder.ColumnConfigs) (columnNames []string) {
			return []string{"Fpay_id"}
		},
	},
	sqlbuilder.Index{
		ColumnNames: func(tableColumns sqlbuilder.ColumnConfigs) (columnNames []string) {
			return []string{"order_id"}
		},
	},
)

type PayOrderRepository struct {
	repository sqlbuilder.Repository[PayModelOrderModel]
}

func NewPayOrderRepository() PayOrderRepository {
	return PayOrderRepository{
		repository: sqlbuilder.NewRepository[PayModelOrderModel](table_pay_order),
	}
}

/*
PayId       string `gorm:"column:Fpay_id" json:"payId"`
OrderId     int64  `gorm:"column:Forder_id" json:"orderId"`
TotalAmount int    `gorm:"column:Ftotal_amount" json:"totalAmount"`
PayAmount   int    `gorm:"column:Fpay_amount" json:"payAmount"`
PayAgent    string `gorm:"column:Fpay_agent" json:"payAgent"`
State       string `gorm:"column:Fstate" json:"state"`
UserId      int64  `gorm:"column:Fuser_id" json:"userId"`
ClientIp    string `gorm:"column:Fclient_ip" json:"clientIp"`
PayUrl      string `gorm:"column:Fpay_url" json:"payUrl"`
ReturnUrl   string `gorm:"column:Freturn_url" json:"returnUrl"`
NotifyUrl   string `gorm:"column:Fnotify_url" json:"notifyUrl"`
PayParam    string `gorm:"column:Fpay_param" json:"payParams"`
*/
type PayOrderCreateIn struct {
	PayId       string `json:"payId"`
	OrderId     string `json:"orderId"`
	OrderAmount int    `json:"totalAmount"`
	PayAmount   int    `json:"payAmount"`
	PayAgent    string `json:"payAgent"`
	State       string `json:"state"`
	UserId      string `json:"userId"`
	ClientIp    string `json:"clientIp"`
	PayUrl      string `json:"payUrl"`
	ReturnUrl   string `json:"returnUrl"`
	NotifyUrl   string `json:"notifyUrl"`
	PayParam    string `json:"payParams"`
	CreatedAt   string `json:"createdAt"`
	Expire      int    `json:"expire"`
}

func (in PayOrderCreateIn) Fields() sqlbuilder.Fields {
	return sqlbuilder.Fields{
		NewPayId(in.PayId).SetRequired(true),
		NewOrderId(in.OrderId).SetRequired(true),
		NewTotalAmount(in.OrderAmount).SetRequired(true),
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
	}
}

func (po PayOrderRepository) Create(in PayOrderCreateIn) (err error) {
	err = po.repository.Insert(in.Fields())
	if err != nil {
		return err
	}
	return nil
}

func (po PayOrderRepository) GetByPayId(payId string) (model PayModelOrderModel, exists bool, err error) {
	fs := sqlbuilder.Fields{
		NewPayId(payId).AppendWhereFn(sqlbuilder.ValueFnForward),
	}
	model, exists, err = po.repository.First(fs)
	if err != nil {
		return model, exists, err
	}
	return model, exists, nil
}

func (po PayOrderRepository) GetByPayIdMust(payId string) (model PayModelOrderModel, err error) {
	model, exists, err := po.GetByPayId(payId)
	if !exists {
		err = sqlbuilder.ERROR_NOT_FOUND
		return model, err
	}
	return model, nil
}

func (po PayOrderRepository) GetByOrderId(orderId string) (models PayOrderModels, err error) {
	fs := sqlbuilder.Fields{
		NewOrderId(orderId).AppendWhereFn(sqlbuilder.ValueFnForward),
	}
	models, err = po.repository.All(fs)
	if err != nil {
		return models, err
	}
	return models, nil
}

type ChangeStatusIn struct {
	PayId       string
	NewState    string
	OldState    string
	ExtraFields sqlbuilder.Fields
}

func (in ChangeStatusIn) Fields() sqlbuilder.Fields {
	fs := sqlbuilder.Fields{
		NewPayId(in.PayId).SetRequired(true).ShieldUpdate(true).AppendWhereFn(sqlbuilder.ValueFnForward),
		NewState(in.NewState).Apply(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			//查询条件值使用旧状态值
			f.WhereFns.ResetSetValueFn(func(inputValue any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) {
				return in.OldState, nil
			})
		}),
	}
	fs.Add(in.ExtraFields...)
	return fs
}

func (po PayOrderRepository) ChangeStatus(in ChangeStatusIn) (err error) {
	fs := in.Fields()
	err = po.repository.Update(fs)
	if err != nil {
		return err
	}
	return nil
}

func (po PayOrderRepository) Pay(payId string, paidState string, oldState string) (err error) {
	in := ChangeStatusIn{
		PayId:       payId,
		NewState:    paidState,
		OldState:    oldState,
		ExtraFields: sqlbuilder.Fields{NewPayAt(time.Now().Format(time.DateTime))},
	}
	err = po.ChangeStatus(in)
	if err != nil {
		return err
	}
	return nil
}

func (po PayOrderRepository) CloseByPayId(payId string, closeState string, oldState string) (err error) {
	in := ChangeStatusIn{
		PayId:       payId,
		NewState:    closeState,
		OldState:    oldState,
		ExtraFields: sqlbuilder.Fields{NewClosedAt(time.Now().Format(time.DateTime))},
	}
	err = po.ChangeStatus(in)
	return err
}

type CloseIn = ChangeStatusIn

// CloseBatch 批量关闭订单支付状态，如果存在多个支付流水号，则全部关闭。当订单关闭时，使用事务批量关闭。
func (po PayOrderRepository) CloseBatch(closeInArr ...CloseIn) (err error) {

	extraFields := sqlbuilder.Fields{NewClosedAt(time.Now().Format(time.DateTime))}
	for i := range closeInArr {
		closeInArr[i].ExtraFields = extraFields
	}

	po.repository.Transaction(func(txRepository sqlbuilder.Repository[PayModelOrderModel]) (err error) {
		for _, closeIn := range closeInArr {
			fs := closeIn.Fields()
			err = txRepository.Update(fs)
			if err != nil {
				return err
			}
		}
		return nil

	})

	return nil
}

func (po PayOrderRepository) Failed(payId string, failedState string, oldState string) (err error) {
	in := ChangeStatusIn{PayId: payId, NewState: failedState, OldState: oldState}
	err = po.ChangeStatus(in)
	return err
}
