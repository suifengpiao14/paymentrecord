package paymentrecord

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/cast"
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
	OrderId     string `json:"orderId"`
	PayAgent    string `json:"payAgent"`   // 支付机构 weixin:微信 alipay:支付宝
	OrderAmount int    `json:"orderPrice"` // 订单金额，单位分
	PayAmount   int    `json:"payAmount"`  // 实际支付金额，单位分
	PayParam    string `json:"payParam"`
	UserId      string `json:"userId"`
	ClientIp    string `json:"clientIp"`
}

const (
	PayingAgent_Wechat = "weixin"
	PayingAgent_Alipay = "alipay"
)

type PayOrder struct {
	PayId       string        `json:"payId"`
	OrderId     string        `json:"orderId"`
	OrderAmount int           `json:"orderPrice"`
	PayAmount   int           `json:"paidPrice"`
	PayUrl      string        `json:"payUrl"`
	State       PayOrderState `json:"state"`
	Expire      int           `json:"timeOut"`
	CreatedAt   string        `json:"date"`
	PayAgent    string        `json:"payingAgent"`
}

func (m *PayOrder) GetStateFSM() (stateMachine *PayOrderStateMachine) {
	stateMachine = NewPayOrderStateMachine(m.State)
	return stateMachine
}

type PayOrders []PayOrder

// PaidMoney 支付单中已支付金额总和，使用时要确保全部为一个订单的支付单，否则无意义
func (orders PayOrders) PaidMoney() (paidMoney int) {
	if len(orders) == 0 {
		return 0
	}
	firstOrderId := orders[0].OrderId
	for _, order := range orders {
		if order.OrderId != firstOrderId { // 确保只统计同一个支付单的金额
			err := errors.New("PayOrders.PaidMoney 方法只能用于同一个订单的支付单")
			panic(err)
		}
		if order.State == PayOrderModel_state_paid {
			paidMoney += cast.ToInt(order.PayAmount)
		}
	}
	return paidMoney
}

func (orders PayOrders) IsPayFinished() (payfinished bool) {
	if len(orders) == 0 {
		return true
	}
	orderPrice := orders[0].OrderAmount
	payfinished = orderPrice >= orders.PaidMoney() // 所有支付单支付金额总和大于等于订单金额即为支付完成
	return payfinished
}

func (orders PayOrders) CanClose() (err error) {
	for _, order := range orders {
		stateMachine := NewPayOrderStateMachine(order.State)
		err = stateMachine.CanClose()
		if err != nil {
			return err
		}

	}
	return nil
}

type Config struct {
	Key       string `json:"key"`
	NotifyUrl string `json:"notifyUrl"`
	ReturnUrl string `json:"returnUrl"`
	PayQf     int    `json:"payQf"`
	PayUrl    string `json:"payUrl"`
}

// Create 创建订单
func (s PayOrderService) Create(in PayOrderCreateIn) (out *PayOrder, err error) {
	err = in.Validate()
	if err != nil {
		return nil, err
	}
	payId := PayNOGenerator()
	createdAt := time.Now().Format(time.DateTime)
	payOrderIn := repository.PayOrderCreateIn{
		PayId:       payId,
		OrderId:     in.OrderId,
		OrderAmount: in.OrderAmount,
		PayAmount:   in.PayAmount,
		PayAgent:    in.PayAgent,
		State:       string(PayOrderModel_state_pending),
		UserId:      in.UserId,
		ClientIp:    in.ClientIp,
		PayParam:    in.PayParam,
		CreatedAt:   createdAt,
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
		State:       PayOrderState(payOrderIn.State),
		Expire:      payOrderIn.Expire,
		CreatedAt:   payOrderIn.CreatedAt,
		PayAgent:    payOrderIn.PayAgent,
	}
	return out, nil
}

// Validate 验证请求参数
func (req *PayOrderCreateIn) Validate() error {

	payAgents := []string{PayingAgent_Alipay, PayingAgent_Wechat}
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

// PayNOGenerator 生成订单ID（格式：YYYYMMDDHHMMSS + 4位随机数）
func PayNOGenerator() string {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 生成时间部分（格式：YYYYMMDDHHMMSS）
	timePart := time.Now().Format("20060102150405")

	// 生成4位随机数（1-9之间的数字）
	randPart := fmt.Sprintf("%d%d%d%d",
		rand.Intn(9)+1,
		rand.Intn(9)+1,
		rand.Intn(9)+1,
		rand.Intn(9)+1)

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
	for _, v := range models {
		payOrder := PayOrder{
			PayId:       v.PayId,
			OrderId:     v.OrderId,
			OrderAmount: v.OrderAmount,
			PayAmount:   v.PayAmount,
			PayUrl:      v.PayUrl,
			State:       PayOrderState(v.State),
			Expire:      v.Expire,
		}
		payOrders = append(payOrders, payOrder)
	}
	return payOrders, nil
}

// Pay 支付订单
func (s PayOrderService) Pay(payId string) (err error) {
	r := s.repository
	model, err := r.GetByPayIdMust(payId)
	if err != nil {
		return err
	}
	stateFSM := NewPayOrderStateMachine(PayOrderState(model.State))
	err = stateFSM.CanPay()
	if err != nil {
		return err
	}
	err = r.Pay(model.PayId, PayOrderModel_state_paid.String(), model.State)
	if err != nil {
		return err
	}
	return nil
}

func (s PayOrderService) IsPaid(orderId string) (ok bool, err error) {
	records, err := s.GetOrderPayInfo(orderId)
	if err != nil {
		return false, err
	}
	payFinished := records.IsPayFinished()
	return payFinished, nil
}

// CloseByOrderId 关闭订单支付，当订单关闭时，关闭订单对应的支付单
func (s PayOrderService) CloseByOrderId(orderId string) (err error) {
	records, err := s.GetOrderPayInfo(orderId)
	if err != nil {
		return err
	}

	r := s.repository
	closeBatchIn := make([]repository.CloseIn, 0)
	for _, v := range records {
		closeIn := repository.CloseIn{PayId: v.PayId, NewState: string(PayOrderModel_state_closed), OldState: v.State.String()}
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
		State:       PayOrderState(model.State),
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
	err = stateFSM.CanPay()
	if err != nil {
		return err
	}

	r := s.repository
	err = r.CloseByPayId(record.PayId, PayOrderModel_state_paid.String(), record.State.String())
	if err != nil {
		return err
	}
	return nil
}
func (s PayOrderService) ExpiredByPayId(payId string) (err error) {
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
	err = r.CloseByPayId(record.PayId, PayOrderModel_state_expired.String(), record.State.String())
	if err != nil {
		return err
	}
	return nil
}

func (s PayOrderService) Failed(payId string) (err error) {
	record, err := s.GetByPayId(payId)
	if err != nil {
		return err
	}
	stateFSM := record.GetStateFSM()
	err = stateFSM.CanPay()
	if err != nil {
		return err
	}

	r := s.repository
	err = r.Failed(record.PayId, PayOrderModel_state_failed.String(), record.State.String())
	if err != nil {
		return err
	}
	return nil
}
