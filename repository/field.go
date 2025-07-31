package repository

import (
	"github.com/suifengpiao14/commonlanguage"
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

func NewId(id int) *sqlbuilder.Field {
	return commonlanguage.NewId(id).SetMaximum(sqlbuilder.Int_maximum_bigint).SetTag(sqlbuilder.Tag_autoIncrement)
}

func NewOrderId(orderId string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(orderId, "orderId", "订单号", 64)
}

func NewOrderAmount(orderAmount int) *sqlbuilder.Field {
	return sqlbuilder.NewIntField(orderAmount, "orderAmount", "订单总金额，单位分", sqlbuilder.Int_maximum_bigint).SetTag(sqlbuilder.Tag_unsigned)
}
func NewPayAmount(payAmount int) *sqlbuilder.Field {
	return sqlbuilder.NewIntField(payAmount, "paymentAmount", "支付金额，单位分", sqlbuilder.Int_maximum_bigint).SetTag(sqlbuilder.Tag_unsigned)
}

const (
	PayingAgent_Wechat = "weixin"
	PayingAgent_Alipay = "alipay"
)

func NewPayAgent(payAgent string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(payAgent, "payAgent", "支付类型", 0).AppendEnum(
		sqlbuilder.Enum{
			Key:   PayingAgent_Wechat,
			Title: "微信",
		},
		sqlbuilder.Enum{
			Key:   PayingAgent_Alipay,
			Title: "支付宝",
		},
	)
}

type PayOrderState string

func (s PayOrderState) String() string {
	return string(s)
}

const (
	PayOrderModel_state_pending PayOrderState = "pending" //未支付
	PayOrderModel_state_paid    PayOrderState = "paid"    //已支付
	PayOrderModel_state_expired PayOrderState = "expired" //已过期（可选扩展）
	PayOrderModel_state_failed  PayOrderState = "failed"  //支付失败（可选扩展）
	PayOrderModel_state_closed  PayOrderState = "closed"  //已关闭（可选扩展）
	PayOrderModel_state_unknown PayOrderState = "unknown"
)

func NewState(state string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(state, "state", "支付状态", 0).AppendEnum(
		sqlbuilder.Enum{
			Key:   PayOrderModel_state_pending.String(),
			Title: "未支付",
		},
		sqlbuilder.Enum{
			Key:   PayOrderModel_state_paid.String(),
			Title: "已支付",
		},
		sqlbuilder.Enum{
			Key:   PayOrderModel_state_expired.String(),
			Title: "已过期",
		},
		sqlbuilder.Enum{
			Key:   PayOrderModel_state_failed.String(),
			Title: "支付失败",
		},
		sqlbuilder.Enum{
			Key:   PayOrderModel_state_closed.String(),
			Title: "已关闭",
		},
		sqlbuilder.Enum{
			Key:   PayOrderModel_state_unknown.String(),
			Title: "未知状态",
		},
	)
}

func NewPayId(payId string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(payId, "payNo", "支付流水号", 64)
}

func NewUserId(userId string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(userId, "userId", "用户ID", 64)
}

func NewPayAt(payAt string) *sqlbuilder.Field {
	f := commonlanguage.NewTime(payAt).SetName("payAt").SetTitle("支付时间")
	return f
}

func NewClientIp(clientIp string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(clientIp, "clientIp", "客户端IP地址", 20)
}

func NewClosedAt(closedAt string) *sqlbuilder.Field {
	f := commonlanguage.NewTime(closedAt).SetName("closedAt").SetTitle("关单时间")
	return f
}
func NewExpiredAt(expiredAt string) *sqlbuilder.Field {
	f := commonlanguage.NewTime(expiredAt).SetName("expiredAt").SetTitle("过期时间")
	return f
}
func NewFailedAt(failedAt string) *sqlbuilder.Field {
	f := commonlanguage.NewTime(failedAt).SetName("failedAt").SetTitle("失败时间")
	return f
}

func NewExpire(expire int) *sqlbuilder.Field {
	return sqlbuilder.NewIntField(expire, "expire", "超时时间，单分钟", 0).SetTag(sqlbuilder.Tag_unsigned)
}

func NewPayUrl(payUrl string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(payUrl, "payUrl", "支付链接地址", 0)
}

func NewReturnUrl(returnUrl string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(returnUrl, "returnUrl", "支付完成后的回调地址", 0)
}
func NewNotifyUrl(notifyUrl string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(notifyUrl, "notifyUrl", "支付完成后的通知地址", 0)
}
func NewPayParam(payParam string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(payParam, "payParam", "支付参数，json格式", 0)
}
func NewRemark(remark string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(remark, "remark", "支付备注(如失败原因)", 0)
}

func NewRecipientId(recipientId string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(recipientId, "recipientId", "收款人ID", 64)
}

func NewRecipientName(recipientName string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(recipientName, "recipientName", "收款人名称", 64)
}

func NewPayerId(payerId string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(payerId, "payerId", "付款人ID", 64)
}

func NewPayerName(payerName string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(payerName, "payerName", "付款人名称", 64)
}

var NewCreatedAt = commonlanguage.NewCreatedAt
var NewUpdatedAt = commonlanguage.NewUpdatedAt

var NewDeletedAt = commonlanguage.NewDeletedAt

const (
	Pay_order_type_wechat = 1
	Pay_order_type_alipay = 2
)
