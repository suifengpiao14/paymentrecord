package paymentrecord

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/looplab/fsm"
	"github.com/suifengpiao14/paymentrecord/repository"
)

type PayOrderStateMachine struct {
	ExtraAttrs []any
	State      repository.PayOrderState
	fsm        *fsm.FSM
}

type StateMachineError struct {
	ExtraAttrs      []any                    `json:"extraAttrs"`
	Message         string                   `json:"message"`
	CurrentState    repository.PayOrderState `json:"currentState"`
	AvailableEvents []string                 `json:"availableEvents"`
}

func (e StateMachineError) Error() string {
	b, _ := json.Marshal(e)
	s := string(b)
	return s

}

func (matchine *PayOrderStateMachine) CanPay() (err error) {
	if !matchine.fsm.Can(Event_Pay) {
		err = matchine.makeError(Event_Pay)
		return err
	}
	return nil
}
func (matchine *PayOrderStateMachine) CanFail() (err error) {
	if !matchine.fsm.Can(Event_Fail) {
		err = matchine.makeError(Event_Fail)
		return err
	}
	return nil
}
func (matchine *PayOrderStateMachine) makeError(event string) (err error) {
	message := fmt.Sprintf("当前状态(%s)不可触发/执行%s事件/动作，可用事件/动作：%s", matchine.State, event, strings.Join(matchine.fsm.AvailableTransitions(), ","))
	stateErr := StateMachineError{

		ExtraAttrs:      matchine.ExtraAttrs,
		Message:         message,
		CurrentState:    matchine.State,
		AvailableEvents: matchine.fsm.AvailableTransitions(),
	}
	if stateErr.ExtraAttrs == nil {
		stateErr.ExtraAttrs = make([]any, 0)
	}
	return stateErr
}
func (matchine *PayOrderStateMachine) CanExpire() (err error) {
	if !matchine.fsm.Can(Event_Expire) {
		err = matchine.makeError(Event_Expire)
		return err
	}
	return nil
}

func (matchine *PayOrderStateMachine) CanClose() (err error) {
	if !matchine.fsm.Can(Event_Close) {
		err = matchine.makeError(Event_Close)
		return err
	}
	return nil
}

func NewPayOrderStateMachine(state repository.PayOrderState, extraAttrs ...any) *PayOrderStateMachine {
	stateMachine := &PayOrderStateMachine{State: state, ExtraAttrs: extraAttrs}
	stateMachine.InitFSM()
	return stateMachine
}

const (
	Event_Pay    = "actionPay"
	Event_Expire = "actionExpire"
	Event_Fail   = "actionFail"
	Event_Close  = "actionClose"
)

func (o *PayOrderStateMachine) InitFSM() {
	o.fsm = fsm.NewFSM(
		o.State.String(),
		fsm.Events{
			{
				Name: Event_Pay,
				Src: []string{
					repository.PayOrderModel_state_pending.String(),
					repository.PayOrderModel_state_failed.String(), // 支付失败可以继续支付（比如钱包金额不够、选中的优惠券过期等，充值后再支付，增加支付失败状态可以记录原因）
					repository.PayOrderModel_state_paid.String(),   // 支持幂等
				},
				Dst: repository.PayOrderModel_state_paid.String(),
			},
			{
				Name: Event_Expire, // 过期时需要先同步查询，看是否已经支付（比如消息异常导致未同步到数据）

				Src: []string{
					repository.PayOrderModel_state_pending.String(),
					repository.PayOrderModel_state_expired.String(), // 支持幂等
				},
				Dst: repository.PayOrderModel_state_expired.String(),
			},
			{
				Name: Event_Fail,
				Src: []string{
					repository.PayOrderModel_state_pending.String(),
					repository.PayOrderModel_state_failed.String(), // 支持幂等
				},
				Dst: repository.PayOrderModel_state_failed.String(),
			},
			{
				Name: Event_Close,
				Src: []string{
					repository.PayOrderModel_state_pending.String(),
					repository.PayOrderModel_state_closed.String(), // 支持幂等
				},
				Dst: repository.PayOrderModel_state_closed.String(),
			},
		},
		fsm.Callbacks{
			"enter_state": func(ctx context.Context, e *fsm.Event) {
				// 这里更新实际状态
				o.State = repository.PayOrderState(e.Dst)
			},
		},
	)
}
