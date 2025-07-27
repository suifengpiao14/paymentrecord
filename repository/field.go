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
	return commonlanguage.NewId(id)
}

func NewOrderId(orderId string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(orderId, "orderId", "订单号", 0)
}

func NewTotalAmount(totalAmount int) *sqlbuilder.Field {
	return sqlbuilder.NewIntField(totalAmount, "totalAmount", "订单总金额，单位分", 0)
}
func NewPayAmount(payAmount int) *sqlbuilder.Field {
	return sqlbuilder.NewIntField(payAmount, "paymentAmount", "支付金额，单位分", 0)
}

func NewPayAgent(type_ string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(type_, "payAgent", "支付类型1-微信，2-支付宝", 0)
}

func NewState(state string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(state, "state", "支付状态", 0)
}

func NewPayId(payId string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(payId, "payNo", "支付流水号", 0)
}

func NewUserId(userId string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(userId, "userId", "用户ID", 0)
}

func NewPayAt(payAt string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(payAt, "paidAt", "支付时间", 0)
}

func NewClientIp(clientIp string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(clientIp, "clientIp", "客户端IP地址", 0)
}

func NewClosedAt(closedAt string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(closedAt, "closedAt", "关单时间", 0)
}

func NewExpire(expire int) *sqlbuilder.Field {
	return sqlbuilder.NewIntField(expire, "expire", "超时时间，单位秒", 0)
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

var NewCreatedAt = commonlanguage.NewCreatedAt

var NewDeletedAt = commonlanguage.NewDeletedAt

const (
	Pay_order_type_wechat = 1
	Pay_order_type_alipay = 2
)
