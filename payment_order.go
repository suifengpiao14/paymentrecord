package paymentrecord

import (
	"time"

	"github.com/suifengpiao14/paymentrecord/repository"
	"github.com/suifengpiao14/sqlbuilder"
)

type _PayOrderService struct {
	orderRepository  repository.PayOrderRepository
	recordRepository repository.PayRecordRepository
}

type PayOrderSetIn struct {
	OrderId     string `json:"orderId"`
	OrderAmount int    `json:"orderAmount"`
	UserId      string `json:"userId"`
	Remark      string `json:"remark"`
}

func (s _PayOrderService) Set(in PayOrderSetIn) (err error) {
	payOrderSetIn := repository.PayOrderSetIn{
		OrderId:     in.OrderId,
		OrderAmount: in.OrderAmount,
		UserId:      in.UserId,
		Remark:      in.Remark,
	}
	err = s.orderRepository.Set(payOrderSetIn)
	if err != nil {
		return err
	}
	return nil
}

func (s _PayOrderService) Close(orderId string, reson string) (err error) {
	orderStateMichine := s.orderRepository.GetStateMachine()
	payOrderStateModel, err := orderStateMichine.GetStateByIdentity(orderId)
	if err != nil {
		return err
	}
	// 验证支付单是否可以关闭
	err = orderStateMichine.CanAsErr(payOrderStateModel.State, repository.Action_pay_order_Close)
	if err != nil {
		return err
	}

	//验证支付单下的所有支付记录是否可以关闭
	recordStateModels, err := orderStateMichine.GetAll(sqlbuilder.Fields{
		repository.NewOrderId(orderId),
	})
	if err != nil {
		return err
	}
	for _, recordStateModel := range recordStateModels {
		err = orderStateMichine.CanAsErr(recordStateModel.State, repository.Action_pay_record_Close)
		if err != nil {
			return err
		}
	}
	stateCloseExtraFs := sqlbuilder.Fields{
		repository.NewClosedAt(time.Now().Format(time.DateTime)),
		repository.NewRemark(reson),
	}
	err = orderStateMichine.Transaction(func(txHandler sqlbuilder.Handler) (err error) {
		txOrderStateMachine := orderStateMichine.WithTxHandler(txHandler)
		//关闭订单
		err = txOrderStateMachine.Transform(repository.Action_pay_order_Close, payOrderStateModel.State, payOrderStateModel.Identity, stateCloseExtraFs...)
		if err != nil {
			return err
		}
		txRecordStateMachine := s.recordRepository.GetStateMachine().WithTxHandler(txHandler)
		//关闭订单下的所有支付记录
		for _, recordStateModel := range recordStateModels {
			err = txRecordStateMachine.Transform(repository.Action_pay_record_Close, recordStateModel.State, recordStateModel.Identity, stateCloseExtraFs...)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
