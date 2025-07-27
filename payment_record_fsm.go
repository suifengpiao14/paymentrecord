package paymentrecord

import (
	"context"
	"encoding/json"

	"github.com/looplab/fsm"
)

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

type PayOrderStateMachine struct {
	ExtraAttrs []any
	State      PayOrderState
	fsm        *fsm.FSM
}

type StateMachineError struct {
	ExtraAttrs      []any         `json:"extraAttrs"`
	Message         string        `json:"message"`
	CurrentState    PayOrderState `json:"currentState"`
	AvailableEvents []string      `json:"availableEvents"`
}

func (e StateMachineError) Error() string {
	b, _ := json.Marshal(e)
	s := string(b)
	return s

}

func (matchine *PayOrderStateMachine) CanPay() (err error) {
	if !matchine.fsm.Can(Event_Pay) {
		err = matchine.makeError("当前状态不可支付")
		return err
	}
	return nil
}
func (matchine *PayOrderStateMachine) CanFail() (err error) {
	if !matchine.fsm.Can(Event_Fail) {
		err = matchine.makeError("当前状态不可支付")
		return err
	}
	return nil
}
func (matchine *PayOrderStateMachine) makeError(message string) (err error) {
	return StateMachineError{
		ExtraAttrs:      matchine.ExtraAttrs,
		Message:         message,
		CurrentState:    matchine.State,
		AvailableEvents: matchine.fsm.AvailableTransitions(),
	}
}
func (matchine *PayOrderStateMachine) CanExpire() (err error) {
	if !matchine.fsm.Can(Event_Expire) {
		err = matchine.makeError("当前状态不可过期")
		return err
	}
	return nil
}

func (matchine *PayOrderStateMachine) CanClose() (err error) {
	if !matchine.fsm.Can(Event_Close) {
		err = matchine.makeError("当前状态不可关闭")
		return err
	}
	return nil
}

func NewPayOrderStateMachine(state PayOrderState, extraAttrs ...any) *PayOrderStateMachine {
	stateMachine := &PayOrderStateMachine{State: state, ExtraAttrs: extraAttrs}
	stateMachine.InitFSM()
	return stateMachine
}

const (
	Event_Pay    = "pay"
	Event_Expire = "expire"
	Event_Fail   = "fail"
	Event_Close  = "close"
)

func (o *PayOrderStateMachine) InitFSM() {
	o.fsm = fsm.NewFSM(
		o.State.String(),
		fsm.Events{
			{Name: Event_Pay, Src: []string{PayOrderModel_state_pending.String()}, Dst: PayOrderModel_state_paid.String()},
			{Name: Event_Expire, Src: []string{PayOrderModel_state_pending.String()}, Dst: PayOrderModel_state_expired.String()},
			{Name: Event_Fail, Src: []string{PayOrderModel_state_pending.String()}, Dst: PayOrderModel_state_failed.String()},
			{Name: Event_Close, Src: []string{PayOrderModel_state_pending.String(), PayOrderModel_state_paid.String()}, Dst: PayOrderModel_state_closed.String()},
		},
		fsm.Callbacks{
			"enter_state": func(ctx context.Context, e *fsm.Event) {
				// 这里更新实际状态
				o.State = PayOrderState(e.Dst)
			},
		},
	)
}
