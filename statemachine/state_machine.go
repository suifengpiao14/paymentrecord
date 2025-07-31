package statemachine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/looplab/fsm"
	"github.com/pkg/errors"
	"github.com/suifengpiao14/sqlbuilder"
)

type StateMachineError struct {
	WhereData       any      `json:"identity"`
	ExtraUpdateData any      `json:"extraAttrs"`
	Message         string   `json:"message"`
	CurrentState    string   `json:"currentState"`
	AvailableEvents []string `json:"availableEvents"`
}

func (e StateMachineError) Error() string {
	b, _ := json.Marshal(e)
	s := string(b)
	return s

}

type StateMachine struct {
	fsmEvents       fsm.Events
	callbacks       map[string]fsm.Callback
	stateRepository StateRepository
}

// Action 定义了状态转换的动作，包括动作名称、源状态和目标状态 ，这里重新定义Action，不直接使用fsm.Events，是应为日常大家对触发某个动作转换状态更习惯，也更清晰。ActionName 也会自然想到方法名
type Action struct {
	ActionName string
	SrcStates  []string
	DstState   string
}

type Actions []Action

func (actions Actions) fsmEvents() fsm.Events {
	fsmEvents := make(fsm.Events, len(actions))
	for _, action := range actions {
		fsmEvents = append(fsmEvents, fsm.EventDesc{
			Src: action.SrcStates,
			Dst: action.DstState,
		})
	}
	return fsmEvents
}

func NewStateMachine(actions Actions, stateRepository StateRepository) *StateMachine {
	stateMachine := &StateMachine{
		stateRepository: stateRepository,
		fsmEvents:       actions.fsmEvents(),
		callbacks:       make(map[string]fsm.Callback),
	}
	stateMachine.callbacks["enter_state"] = func(ctx context.Context, e *fsm.Event) {
		// 这里更新实际状态
		if len(e.Args) < 1 {
			e.Err = fmt.Errorf("where condition required")
			return
		}
		arg0 := e.Args[0]
		if arg0 == nil {
			e.Err = fmt.Errorf("where condition required")
			return
		}
		var WhereFields sqlbuilder.Fields
		if fs, ok := arg0.(sqlbuilder.Fields); ok {
			WhereFields = fs
		} else {
			e.Err = fmt.Errorf("WhereFields must be sqlbuilder.Fields")
			return
		}

		var extraFields sqlbuilder.Fields

		if len(e.Args) > 1 {
			arg1 := e.Args[1]
			if arg1 != nil {
				if fs, ok := arg1.(sqlbuilder.Fields); ok {
					extraFields = fs
				} else {
					e.Err = fmt.Errorf("extraFields must be sqlbuilder.Fields")
					return
				}
			}
		}
		in := ChangeStatusIn{
			NewState:    e.Dst,
			OldState:    e.Src,
			WhereFields: WhereFields,
			ExtraFields: extraFields,
		}
		e.Err = stateRepository.ChangeState(in)
	}

	return stateMachine
}

func (matchine StateMachine) makeError(action string, currentState string, whereFields sqlbuilder.Fields, availableTransitions []string, extraFields sqlbuilder.Fields) (err error) {
	message := fmt.Sprintf("当前状态(%s)不可触发/执行%s事件/动作，可用事件/动作：%s", currentState, action, strings.Join(availableTransitions, ","))
	whereData, whereDataErr := whereFields.Data()
	ExtraUpdateData, updateDataErr := extraFields.Data()
	stateErr := StateMachineError{
		WhereData:       whereData,
		ExtraUpdateData: ExtraUpdateData,
		Message:         message,
		CurrentState:    currentState,
		AvailableEvents: availableTransitions,
	}
	if stateErr.ExtraUpdateData == nil {
		stateErr.ExtraUpdateData = make([]any, 0)
	}
	err = stateErr
	if whereDataErr != nil {
		err = errors.WithMessagef(err, "whereFields.Data() error:%s", whereDataErr.Error())
	}
	if updateDataErr != nil {
		err = errors.WithMessagef(err, "extraFields.Data() error:%s", updateDataErr.Error())
	}
	return err
}

func (stateMachine StateMachine) Can(currentState string, action string) bool {
	fsmInstace := stateMachine.getFSM(currentState)
	canDo := fsmInstace.Can(action)
	fsmInstace.Transition()
	return canDo
}

func (stateMachine StateMachine) getFSM(currentState string) *fsm.FSM {
	return fsm.NewFSM(currentState, stateMachine.fsmEvents, stateMachine.callbacks)
}

func (stateMachine StateMachine) Transaction(action string, currentState string, whereFields sqlbuilder.Fields, extraFields ...*sqlbuilder.Field) (err error) {
	fsmInstace := stateMachine.getFSM(currentState)
	updateDataFields := sqlbuilder.Fields(extraFields)
	ok := fsmInstace.Can(action)
	if !ok {
		err = stateMachine.makeError(action, currentState, whereFields, fsmInstace.AvailableTransitions(), updateDataFields)
		return err
	}
	ctx := context.Background()
	err = fsmInstace.Event(ctx, action, whereFields, updateDataFields)
	if err != nil {
		return err
	}
	return nil
}
