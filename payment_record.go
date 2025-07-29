package paymentrecord

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/suifengpiao14/paymentrecord/repository"
)

type PayOrderService struct {
	config     Config
	repository repository.PayOrderRepository
}

func NewPayOrderService(config Config, repository repository.PayOrderRepository) PayOrderService {
	return PayOrderService{
		config:     config,
		repository: repository,
	}

}

type PayOrderCreateIn struct {
	PayId       string `json:"payId"`
	OrderId     string `json:"orderId"`
	PayAgent    string `json:"payAgent"`   // 支付机构 weixin:微信 alipay:支付宝
	OrderAmount int    `json:"orderPrice"` // 订单金额，单位分
	PayAmount   int    `json:"payAmount"`  // 实际支付金额，单位分
	PayParam    string `json:"payParam"`
	UserId      string `json:"userId"`
	ClientIp    string `json:"clientIp"`
}

type PayOrder struct {
	PayId       string                   `json:"payId"`
	OrderId     string                   `json:"orderId"`
	OrderAmount int                      `json:"orderPrice"`
	PayAmount   int                      `json:"paidPrice"`
	PayUrl      string                   `json:"payUrl"`
	State       repository.PayOrderState `json:"state"`
	Expire      int                      `json:"timeOut"`
	CreatedAt   string                   `json:"date"`
	PayAgent    string                   `json:"payingAgent"`
}

func (m *PayOrder) GetStateFSM() (stateMachine *PayOrderStateMachine) {
	stateMachine = NewPayOrderStateMachine(m.State)
	return stateMachine
}

type PayOrders []PayOrder

type Config struct {
	Key       string `json:"key"`
	NotifyUrl string `json:"notifyUrl"`
	ReturnUrl string `json:"returnUrl"`
	PayQf     int    `json:"payQf"`
	PayUrl    string `json:"payUrl"`
}

// Create 创建订单
func (s PayOrderService) Create(in PayOrderCreateIn) (out *PayOrder, err error) {
	err = in.validate()
	if err != nil {
		return nil, err
	}
	payRecords, err := s.repository.GetByOrderId(in.OrderId)
	if err != nil {
		return nil, err
	}
	orderAmount := payRecords.GetOrderAmount()
	if orderAmount != 0 && orderAmount != in.OrderAmount {
		err = errors.Errorf("订单已开始支付，不许修改金额，已支付的支付单记录订单金额为:%d,当前订单金额为:%d", orderAmount, in.OrderAmount)
		return nil, err
	}

	paidRecords := payRecords.FilterByStatePaid()
	paidAmount := paidRecords.TotalAmount()
	if paidAmount >= in.OrderAmount {
		err = errors.New("订单已支付完成")
		return nil, err
	}
	pendingAmount := payRecords.FilterByStatePending().TotalAmount()
	paidPendingAmount := paidAmount + pendingAmount
	if paidPendingAmount >= in.OrderAmount { // 如果有支付中的订单，则不允许创建新的
		err = errors.New("支付单金额已足够支付订单，请完成支付中的支付单")
		return nil, err
	}
	maxAmount := in.OrderAmount - paidPendingAmount
	if maxAmount < in.PayAmount { // 支付金额总和大于订单金额，不允许创建
		err = errors.Errorf(
			"金额有误(订单金额-%d,已支付金额-%d,待支付金额-%d,当前支付单最大金额-%d),收到支付金额-%d,订单ID-%s",
			in.OrderAmount,
			paidAmount,
			pendingAmount,
			maxAmount,
			in.OrderAmount,
			in.OrderId,
		)
		return nil, err
	}

	createdAt := time.Now().Format(time.DateTime)
	payOrderIn := repository.PayOrderCreateIn{
		PayId:       in.PayId,
		OrderId:     in.OrderId,
		OrderAmount: in.OrderAmount,
		PayAmount:   in.PayAmount,
		PayAgent:    in.PayAgent,
		State:       string(repository.PayOrderModel_state_pending),
		UserId:      in.UserId,
		ClientIp:    in.ClientIp,
		PayParam:    in.PayParam,
	}
	err = s.repository.Create(payOrderIn)
	if err != nil {
		return nil, err
	}

	out = &PayOrder{
		PayId:       payOrderIn.PayId,
		OrderId:     payOrderIn.OrderId,
		OrderAmount: payOrderIn.OrderAmount,
		PayAmount:   payOrderIn.PayAmount,
		PayUrl:      payOrderIn.PayUrl,
		State:       repository.PayOrderState(payOrderIn.State),
		Expire:      payOrderIn.Expire,
		CreatedAt:   createdAt,
		PayAgent:    payOrderIn.PayAgent,
	}
	return out, nil
}

// validate 验证请求参数
func (req *PayOrderCreateIn) validate() error {
	if req.PayId == "" {
		return errors.New("payId不能为空")
	}
	payAgents := []string{repository.PayingAgent_Alipay, repository.PayingAgent_Wechat}
	// 验证type
	if !slices.Contains(payAgents, req.PayAgent) {
		err := errors.Errorf("请传入支付方式=>%s", strings.Join(payAgents, ","))
		return err
	}
	// 验证price
	if req.OrderAmount <= 0 {
		return errors.New("订单金额必须大于0")
	}
	return nil
}

// PayIdGenerator 生成订单ID（格式：YYYYMMDDHHMMSS + 4位随机数）
func PayIdGenerator() string {
	// 设置随机数种子
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 生成时间部分（格式：YYYYMMDDHHMMSS）
	timePart := time.Now().Format("20060102150405")

	// 生成4位随机数（1-9之间的数字）
	randPart := fmt.Sprintf("%d%d%d%d",
		rd.Intn(9)+1,
		rd.Intn(9)+1,
		rd.Intn(9)+1,
		rd.Intn(9)+1)

	// 组合订单ID
	return timePart + randPart
}

// GetOrderPayInfo 获取订单支付信息
func (s PayOrderService) GetOrderPayInfo(orderId string) (payOrders PayOrders, err error) {
	r := s.repository
	models, err := r.GetByOrderId(orderId)
	if err != nil {
		return nil, err
	}
	effectRecords := models.FilterByStateEffect()
	for _, v := range effectRecords {
		payOrder := PayOrder{
			PayId:       v.PayId,
			OrderId:     v.OrderId,
			OrderAmount: v.OrderAmount,
			PayAmount:   v.PayAmount,
			PayUrl:      v.PayUrl,
			State:       repository.PayOrderState(v.State),
			Expire:      v.Expire,
		}
		payOrders = append(payOrders, payOrder)
	}
	return payOrders, nil
}

// Pay 支付订单 返回订单是否已经支付完成（同一个订单下所有已支付的单总额等于订单金额）
func (s PayOrderService) Pay(payId string) (isOrderPayFinished bool, err error) {
	r := s.repository
	model, err := r.GetByPayIdMust(payId)
	if err != nil {
		return false, err
	}
	stateFSM := NewPayOrderStateMachine(repository.PayOrderState(model.State))
	err = stateFSM.CanPay()
	if err != nil {
		return false, err
	}
	if model.StateIsPaid() { // 已支付，直接返回
		return true, nil
	}

	err = r.Pay(model.PayId, repository.PayOrderModel_state_paid.String(), model.State)
	if err != nil {
		return false, err
	}
	//查看是否订单已经支付完成
	payRecords, err := r.GetByOrderId(model.OrderId)
	if err != nil {
		return false, err
	}
	isOrderPayFinished = payRecords.IsOrderPayFinished()
	return isOrderPayFinished, nil
}

func (s PayOrderService) IsPaid(orderId string) (ok bool, err error) {
	records, err := s.repository.GetByOrderId(orderId)
	if err != nil {
		return false, err
	}
	payFinished := records.IsOrderPayFinished()
	return payFinished, nil
}

// GetOrderRestPayRecordAmount 获取订单剩余可创建待支付单的金额
func (s PayOrderService) GetOrderRestPayRecordAmount(orderId string) (restPayRecordAmount int, err error) {
	records, err := s.repository.GetByOrderId(orderId)
	if err != nil {
		return 0, err
	}
	effectRecords := records.FilterByStateEffect()
	restPayRecordAmount = effectRecords.GetOrderAmount() - effectRecords.TotalAmount()
	return restPayRecordAmount, nil
}

// CloseByOrderId 关闭订单支付，当订单关闭时，关闭订单对应的支付单
func (s PayOrderService) CloseByOrderId(orderId string) (err error) {
	records, err := s.repository.GetByOrderId(orderId)
	if err != nil {
		return err
	}

	r := s.repository
	closeBatchIn := make([]repository.CloseIn, 0)
	for _, payRecord := range records {
		if payRecord.StateIsClosed() { // 已关闭的不需要再关闭
			continue
		}
		closeIn := repository.CloseIn{PayId: payRecord.PayId, NewState: string(repository.PayOrderModel_state_closed), OldState: payRecord.State}
		stateMachine := NewPayOrderStateMachine(repository.PayOrderState(payRecord.State))
		err = stateMachine.CanClose()
		if err != nil {
			err = errors.WithMessagef(err, "支付单payId:%s", payRecord.PayId)
			return err
		}
		closeBatchIn = append(closeBatchIn, closeIn)
	}
	err = r.CloseBatch(closeBatchIn...)
	if err != nil {
		return err
	}
	return nil
}

func (s PayOrderService) GetByPayId(payId string) (payOrder *PayOrder, err error) {
	r := s.repository
	model, err := r.GetByPayIdMust(payId)
	if err != nil {
		return nil, err
	}
	out := &PayOrder{
		PayId:       model.PayId,
		OrderId:     model.OrderId,
		OrderAmount: model.OrderAmount,
		PayAmount:   model.PayAmount,
		PayUrl:      model.PayUrl,
		State:       repository.PayOrderState(model.State),
		Expire:      model.Expire,
		CreatedAt:   model.CreatedAt,
		PayAgent:    model.PayAgent,
	}
	return out, nil
}

func (s PayOrderService) CloseByPayId(payId string) (err error) {
	record, err := s.GetByPayId(payId)
	if err != nil {
		return err
	}
	stateFSM := record.GetStateFSM()
	err = stateFSM.CanClose()
	if err != nil {
		return err
	}

	r := s.repository
	err = r.CloseByPayId(record.PayId, repository.PayOrderModel_state_paid.String(), record.State.String())
	if err != nil {
		return err
	}
	return nil
}
func (s PayOrderService) ExpireByPayId(payId string, reason string) (err error) {
	record, err := s.GetByPayId(payId)
	if err != nil {
		return err
	}
	stateFSM := record.GetStateFSM()
	err = stateFSM.CanExpire()
	if err != nil {
		return err
	}

	r := s.repository
	err = r.ExpireByPayId(record.PayId, repository.PayOrderModel_state_expired.String(), record.State.String(), reason)
	if err != nil {
		return err
	}
	return nil
}

func (s PayOrderService) FailByPayId(payId string, reason string) (err error) {
	record, err := s.GetByPayId(payId)
	if err != nil {
		return err
	}
	stateFSM := record.GetStateFSM()
	err = stateFSM.CanFail()
	if err != nil {
		return err
	}

	r := s.repository
	err = r.Failed(record.PayId, repository.PayOrderModel_state_failed.String(), record.State.String(), reason)
	if err != nil {
		return err
	}
	return nil
}
