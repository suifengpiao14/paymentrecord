package statemachine

import "github.com/suifengpiao14/sqlbuilder"

func NewState(state string, enums sqlbuilder.Enums) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(state, "state", "状态", 0).AppendEnum(enums...)
}

func NewIdentity(identity string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(identity, "identity", "对象唯一标识", 64)
}
