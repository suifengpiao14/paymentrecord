package statemachine

import (
	"github.com/pkg/errors"

	"github.com/suifengpiao14/sqlbuilder"
)

// StateModel 状态模型，用于管理对象的状态，json tag 和field 定义的Field.Name 必须一致
type StateModel struct {
	Identity string `json:"identity"` // 对象唯一标识，例如订单号、用户ID等,优先取表中单字段的唯一索引字段作为identity，其次取主键字段作为identity，表配置缺失这2中类型索引会报错
	State    string `json:"state"`
}

type StateModels []StateModel

type StateRepository struct {
	StateEnums sqlbuilder.Enums
	repository sqlbuilder.Repository[StateModel]
}

func NewStateService(tableConfig sqlbuilder.TableConfig, stateDBColumnName string, stateEnums sqlbuilder.Enums) StateRepository {
	stateField := sqlbuilder.NewColumn(stateDBColumnName, NewState("", nil))
	tableConfig = tableConfig.AddColumns(stateField)
	stateService := StateRepository{
		StateEnums: stateEnums,
		repository: sqlbuilder.NewRepository[StateModel](tableConfig),
	}
	return stateService
}

func (s StateRepository) ChangeState(in ChangeStatusIn) (err error) {
	fs := in.Fields(s.StateEnums)
	err = s.repository.Update(fs)
	if err != nil {
		return err
	}
	return nil
}

func (s StateRepository) getAll(whereFields ...*sqlbuilder.Field) (stateModels StateModels, err error) {
	fs := sqlbuilder.Fields(whereFields)
	if len(whereFields) == 0 {
		err = errors.New("whereFields required")
		return nil, err
	}
	table := s.repository.GetTable()
	identityIndex, exists := table.Indexs.SortByColCount(table.Columns).First()
	if !exists {
		return nil, errors.New("table must have unique index or primary key")
	}

	identityFields := identityIndex.Fields(table.Columns, table.Columns.Fields())
	identityCol := identityFields.MakeAsOneDBColumnWithAlias(sqlbuilder.GetFieldName(NewIdentity), table.Columns)
	stateCol := NewState("", nil).MakeDBColumnWithAlias(table.Columns)
	fs[0] = fs[0].SetSelectColumns(identityCol, stateCol)
	stateModels, err = s.repository.All(fs)
	if err != nil {
		return nil, err
	}
	return stateModels, nil
}

type ChangeStatusIn struct {
	NewState    string
	OldState    string
	WhereFields sqlbuilder.Fields
	ExtraFields sqlbuilder.Fields
}

func (in ChangeStatusIn) Fields(stateEnums sqlbuilder.Enums) sqlbuilder.Fields {
	fs := sqlbuilder.Fields{
		NewState(in.NewState, stateEnums).Apply(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			//查询条件值使用旧状态值
			f.WhereFns.ResetSetValueFn(func(inputValue any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) {
				return in.OldState, nil
			})
		}),
	}
	fs = fs.Add(in.WhereFields...)
	fs = fs.Add(in.ExtraFields...)
	return fs
}
