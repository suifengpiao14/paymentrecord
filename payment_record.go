package paymentrecord

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/paymentrecord/repository"
	"github.com/suifengpiao14/sqlbuilder"
)

type PayRecordService struct {
	orderRepository  repository.PayOrderRepository
	recordRepository repository.PayRecordRepository
}

func NewPayRecordService(payRecordRepository repository.PayRecordRepository) (payRecordService *PayRecordService) {
	payRecordService = &PayRecordService{
		recordRepository: payRecordRepository,
	}
	return payRecordService
}

type PayRecordCreateIn struct {
	PayId            string `json:"payId"`
	OrderId          string `json:"orderId"`
	PayAgent         string `json:"payAgent"`   // 支付机构 weixin:微信 alipay:支付宝
	OrderAmount      int    `json:"orderPrice"` // 订单金额，单位分
	PayAmount        int    `json:"payAmount"`  // 实际支付金额，单位分
	PayParam         string `json:"payParam"`
	UserId           string `json:"userId"`
	ClientIp         string `json:"clientIp"`
	RecipientAccount string `json:"recipientAccount"`
	RecipientName    string `json:"recipientName"`
	PaymentAccount   string `json:"paymentAccount"`
	PaymentName      string `json:"paymentName"`
	PayUrl           string `json:"payUrl"`
	NotifyUrl        string `json:"notifyUrl"`
	ReturnUrl        string `json:"returnUrl"`
	Remark           string `json:"remark"`
}

type PayRecord struct {
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

type PayRecords []PayRecord

type Config struct {
	Key       string `json:"key"`
	NotifyUrl string `json:"notifyUrl"`
	ReturnUrl string `json:"returnUrl"`
	PayUrl    string `json:"payUrl"`
}

// Create 创建订单
func (s PayRecordService) Create(in PayRecordCreateIn) (out *PayRecord, err error) {
	err = in.validate()
	if err != nil {
		return nil, err
	}
	payRecords, err := s.recordRepository.GetByOrderId(in.OrderId)
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
	payOrderIn := repository.PayRecordCreateIn{
		PayId:            in.PayId,
		OrderId:          in.OrderId,
		OrderAmount:      in.OrderAmount,
		PayAmount:        in.PayAmount,
		PayAgent:         in.PayAgent,
		State:            string(repository.PayOrderModel_state_pending),
		UserId:           in.UserId,
		ClientIp:         in.ClientIp,
		PayParam:         in.PayParam,
		PayUrl:           in.PayUrl,
		Expire:           0,
		ReturnUrl:        in.ReturnUrl,
		NotifyUrl:        in.NotifyUrl,
		Remark:           in.Remark,
		RecipientAccount: in.RecipientAccount,
		RecipientName:    in.RecipientName,
		PaymentAccount:   in.PaymentAccount,
		PaymentName:      in.PaymentName,
	}
	err = s.recordRepository.Create(payOrderIn)
	if err != nil {
		return nil, err
	}

	out = &PayRecord{
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
func (req *PayRecordCreateIn) validate() error {
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
func (s PayRecordService) GetOrderPayInfo(orderId string) (payOrders PayRecords, err error) {
	r := s.recordRepository
	models, err := r.GetByOrderId(orderId)
	if err != nil {
		return nil, err
	}
	effectRecords := models.FilterByStateEffect()
	for _, v := range effectRecords {
		payOrder := PayRecord{
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
func (s PayRecordService) Pay(payId string) (isOrderPayFinished bool, err error) {
	r := s.recordRepository
	model, err := r.GetByPayIdMust(payId)
	if err != nil {
		return false, err
	}
	exFs := sqlbuilder.Fields{
		repository.NewPaidAt(time.Now().Format(time.DateTime)),
	}
	err = s.recordRepository.GetStateMachine().Transform(repository.Action_pay_record_Pay, model.State, model.PayId, exFs...)
	if err != nil {
		return false, err
	}
	//查看是否订单已经支付完成
	payRecords, err := r.GetByOrderId(model.OrderId)
	if err != nil {
		return false, err
	}
	isOrderPayFinished = payRecords.IsOrderPayFinished()

	if isOrderPayFinished { // 如果订单已经支付完成，则改变pay_order 状态为 已支付
		err = s.orderRepository.GetStateMachine().TransformByIdentity(repository.Action_pay_order_Pay, model.OrderId)
		if err != nil {
			return isOrderPayFinished, err
		}
	}
	return isOrderPayFinished, nil
}

func (s PayRecordService) IsPaid(orderId string) (ok bool, err error) {
	records, err := s.recordRepository.GetByOrderId(orderId)
	if err != nil {
		return false, err
	}
	payFinished := records.IsOrderPayFinished()
	return payFinished, nil
}

// GetOrderRestPayRecordAmount 获取订单剩余可创建待支付单的金额
func (s PayRecordService) GetOrderRestPayRecordAmount(orderId string) (restPayRecordAmount int, err error) {
	records, err := s.recordRepository.GetByOrderId(orderId)
	if err != nil {
		return 0, err
	}
	effectRecords := records.FilterByStateEffect()
	restPayRecordAmount = effectRecords.GetOrderAmount() - effectRecords.TotalAmount()
	return restPayRecordAmount, nil
}

func (s PayRecordService) Get(payId string) (payOrder *PayRecord, err error) {
	r := s.recordRepository
	model, err := r.GetByPayIdMust(payId)
	if err != nil {
		return nil, err
	}
	out := &PayRecord{
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

func (s PayRecordService) Close(payId string) (err error) {
	fs := sqlbuilder.Fields{
		repository.NewClosedAt(time.Now().Format(time.DateTime)),
	}
	err = s.recordRepository.GetStateMachine().TransformByIdentity(repository.Action_pay_record_Close, payId, fs...)
	if err != nil {
		return err
	}
	return nil
}
func (s PayRecordService) Expire(payId string, reason string) (err error) {
	fs := sqlbuilder.Fields{
		repository.NewExpiredAt(time.Now().Format(time.DateTime)),
		repository.NewRemark(reason),
	}
	err = s.recordRepository.GetStateMachine().TransformByIdentity(repository.Action_pay_record_Expire, payId, fs...)
	if err != nil {
		return err
	}
	return nil
}

func (s PayRecordService) Fail(payId string, reason string) (err error) {
	fs := sqlbuilder.Fields{
		repository.NewFailedAt(time.Now().Format(time.DateTime)),
		repository.NewRemark(reason),
	}
	err = s.recordRepository.GetStateMachine().TransformByIdentity(repository.Action_pay_record_Expire, payId, fs...)
	if err != nil {
		return err
	}
	return nil
}
