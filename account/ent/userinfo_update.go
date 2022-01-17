// Code generated by entc, DO NOT EDIT.

package ent

import (
	"account/ent/predicate"
	"account/ent/userinfo"
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserInfoUpdate is the builder for updating UserInfo entities.
type UserInfoUpdate struct {
	config
	hooks    []Hook
	mutation *UserInfoMutation
}

// Where appends a list predicates to the UserInfoUpdate builder.
func (uiu *UserInfoUpdate) Where(ps ...predicate.UserInfo) *UserInfoUpdate {
	uiu.mutation.Where(ps...)
	return uiu
}

// SetUUID sets the "uuid" field.
func (uiu *UserInfoUpdate) SetUUID(u uuid.UUID) *UserInfoUpdate {
	uiu.mutation.SetUUID(u)
	return uiu
}

// SetGmtCreate sets the "gmt_create" field.
func (uiu *UserInfoUpdate) SetGmtCreate(t time.Time) *UserInfoUpdate {
	uiu.mutation.SetGmtCreate(t)
	return uiu
}

// SetNillableGmtCreate sets the "gmt_create" field if the given value is not nil.
func (uiu *UserInfoUpdate) SetNillableGmtCreate(t *time.Time) *UserInfoUpdate {
	if t != nil {
		uiu.SetGmtCreate(*t)
	}
	return uiu
}

// SetGmtModified sets the "gmt_modified" field.
func (uiu *UserInfoUpdate) SetGmtModified(t time.Time) *UserInfoUpdate {
	uiu.mutation.SetGmtModified(t)
	return uiu
}

// SetNillableGmtModified sets the "gmt_modified" field if the given value is not nil.
func (uiu *UserInfoUpdate) SetNillableGmtModified(t *time.Time) *UserInfoUpdate {
	if t != nil {
		uiu.SetGmtModified(*t)
	}
	return uiu
}

// SetNickname sets the "nickname" field.
func (uiu *UserInfoUpdate) SetNickname(s string) *UserInfoUpdate {
	uiu.mutation.SetNickname(s)
	return uiu
}

// SetAvatar sets the "avatar" field.
func (uiu *UserInfoUpdate) SetAvatar(s string) *UserInfoUpdate {
	uiu.mutation.SetAvatar(s)
	return uiu
}

// SetNillableAvatar sets the "avatar" field if the given value is not nil.
func (uiu *UserInfoUpdate) SetNillableAvatar(s *string) *UserInfoUpdate {
	if s != nil {
		uiu.SetAvatar(*s)
	}
	return uiu
}

// SetEmail sets the "email" field.
func (uiu *UserInfoUpdate) SetEmail(s string) *UserInfoUpdate {
	uiu.mutation.SetEmail(s)
	return uiu
}

// SetPhone sets the "phone" field.
func (uiu *UserInfoUpdate) SetPhone(s string) *UserInfoUpdate {
	uiu.mutation.SetPhone(s)
	return uiu
}

// Mutation returns the UserInfoMutation object of the builder.
func (uiu *UserInfoUpdate) Mutation() *UserInfoMutation {
	return uiu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (uiu *UserInfoUpdate) Save(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	if len(uiu.hooks) == 0 {
		if err = uiu.check(); err != nil {
			return 0, err
		}
		affected, err = uiu.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*UserInfoMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = uiu.check(); err != nil {
				return 0, err
			}
			uiu.mutation = mutation
			affected, err = uiu.sqlSave(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(uiu.hooks) - 1; i >= 0; i-- {
			if uiu.hooks[i] == nil {
				return 0, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = uiu.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, uiu.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// SaveX is like Save, but panics if an error occurs.
func (uiu *UserInfoUpdate) SaveX(ctx context.Context) int {
	affected, err := uiu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (uiu *UserInfoUpdate) Exec(ctx context.Context) error {
	_, err := uiu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (uiu *UserInfoUpdate) ExecX(ctx context.Context) {
	if err := uiu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (uiu *UserInfoUpdate) check() error {
	if v, ok := uiu.mutation.Nickname(); ok {
		if err := userinfo.NicknameValidator(v); err != nil {
			return &ValidationError{Name: "nickname", err: fmt.Errorf("ent: validator failed for field \"nickname\": %w", err)}
		}
	}
	return nil
}

func (uiu *UserInfoUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   userinfo.Table,
			Columns: userinfo.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: userinfo.FieldID,
			},
		},
	}
	if ps := uiu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := uiu.mutation.UUID(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeUUID,
			Value:  value,
			Column: userinfo.FieldUUID,
		})
	}
	if value, ok := uiu.mutation.GmtCreate(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: userinfo.FieldGmtCreate,
		})
	}
	if value, ok := uiu.mutation.GmtModified(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: userinfo.FieldGmtModified,
		})
	}
	if value, ok := uiu.mutation.Nickname(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: userinfo.FieldNickname,
		})
	}
	if value, ok := uiu.mutation.Avatar(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: userinfo.FieldAvatar,
		})
	}
	if value, ok := uiu.mutation.Email(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: userinfo.FieldEmail,
		})
	}
	if value, ok := uiu.mutation.Phone(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: userinfo.FieldPhone,
		})
	}
	if n, err = sqlgraph.UpdateNodes(ctx, uiu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{userinfo.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return 0, err
	}
	return n, nil
}

// UserInfoUpdateOne is the builder for updating a single UserInfo entity.
type UserInfoUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *UserInfoMutation
}

// SetUUID sets the "uuid" field.
func (uiuo *UserInfoUpdateOne) SetUUID(u uuid.UUID) *UserInfoUpdateOne {
	uiuo.mutation.SetUUID(u)
	return uiuo
}

// SetGmtCreate sets the "gmt_create" field.
func (uiuo *UserInfoUpdateOne) SetGmtCreate(t time.Time) *UserInfoUpdateOne {
	uiuo.mutation.SetGmtCreate(t)
	return uiuo
}

// SetNillableGmtCreate sets the "gmt_create" field if the given value is not nil.
func (uiuo *UserInfoUpdateOne) SetNillableGmtCreate(t *time.Time) *UserInfoUpdateOne {
	if t != nil {
		uiuo.SetGmtCreate(*t)
	}
	return uiuo
}

// SetGmtModified sets the "gmt_modified" field.
func (uiuo *UserInfoUpdateOne) SetGmtModified(t time.Time) *UserInfoUpdateOne {
	uiuo.mutation.SetGmtModified(t)
	return uiuo
}

// SetNillableGmtModified sets the "gmt_modified" field if the given value is not nil.
func (uiuo *UserInfoUpdateOne) SetNillableGmtModified(t *time.Time) *UserInfoUpdateOne {
	if t != nil {
		uiuo.SetGmtModified(*t)
	}
	return uiuo
}

// SetNickname sets the "nickname" field.
func (uiuo *UserInfoUpdateOne) SetNickname(s string) *UserInfoUpdateOne {
	uiuo.mutation.SetNickname(s)
	return uiuo
}

// SetAvatar sets the "avatar" field.
func (uiuo *UserInfoUpdateOne) SetAvatar(s string) *UserInfoUpdateOne {
	uiuo.mutation.SetAvatar(s)
	return uiuo
}

// SetNillableAvatar sets the "avatar" field if the given value is not nil.
func (uiuo *UserInfoUpdateOne) SetNillableAvatar(s *string) *UserInfoUpdateOne {
	if s != nil {
		uiuo.SetAvatar(*s)
	}
	return uiuo
}

// SetEmail sets the "email" field.
func (uiuo *UserInfoUpdateOne) SetEmail(s string) *UserInfoUpdateOne {
	uiuo.mutation.SetEmail(s)
	return uiuo
}

// SetPhone sets the "phone" field.
func (uiuo *UserInfoUpdateOne) SetPhone(s string) *UserInfoUpdateOne {
	uiuo.mutation.SetPhone(s)
	return uiuo
}

// Mutation returns the UserInfoMutation object of the builder.
func (uiuo *UserInfoUpdateOne) Mutation() *UserInfoMutation {
	return uiuo.mutation
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (uiuo *UserInfoUpdateOne) Select(field string, fields ...string) *UserInfoUpdateOne {
	uiuo.fields = append([]string{field}, fields...)
	return uiuo
}

// Save executes the query and returns the updated UserInfo entity.
func (uiuo *UserInfoUpdateOne) Save(ctx context.Context) (*UserInfo, error) {
	var (
		err  error
		node *UserInfo
	)
	if len(uiuo.hooks) == 0 {
		if err = uiuo.check(); err != nil {
			return nil, err
		}
		node, err = uiuo.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*UserInfoMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = uiuo.check(); err != nil {
				return nil, err
			}
			uiuo.mutation = mutation
			node, err = uiuo.sqlSave(ctx)
			mutation.done = true
			return node, err
		})
		for i := len(uiuo.hooks) - 1; i >= 0; i-- {
			if uiuo.hooks[i] == nil {
				return nil, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = uiuo.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, uiuo.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX is like Save, but panics if an error occurs.
func (uiuo *UserInfoUpdateOne) SaveX(ctx context.Context) *UserInfo {
	node, err := uiuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (uiuo *UserInfoUpdateOne) Exec(ctx context.Context) error {
	_, err := uiuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (uiuo *UserInfoUpdateOne) ExecX(ctx context.Context) {
	if err := uiuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (uiuo *UserInfoUpdateOne) check() error {
	if v, ok := uiuo.mutation.Nickname(); ok {
		if err := userinfo.NicknameValidator(v); err != nil {
			return &ValidationError{Name: "nickname", err: fmt.Errorf("ent: validator failed for field \"nickname\": %w", err)}
		}
	}
	return nil
}

func (uiuo *UserInfoUpdateOne) sqlSave(ctx context.Context) (_node *UserInfo, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   userinfo.Table,
			Columns: userinfo.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: userinfo.FieldID,
			},
		},
	}
	id, ok := uiuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "ID", err: fmt.Errorf("missing UserInfo.ID for update")}
	}
	_spec.Node.ID.Value = id
	if fields := uiuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, userinfo.FieldID)
		for _, f := range fields {
			if !userinfo.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != userinfo.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := uiuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := uiuo.mutation.UUID(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeUUID,
			Value:  value,
			Column: userinfo.FieldUUID,
		})
	}
	if value, ok := uiuo.mutation.GmtCreate(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: userinfo.FieldGmtCreate,
		})
	}
	if value, ok := uiuo.mutation.GmtModified(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: userinfo.FieldGmtModified,
		})
	}
	if value, ok := uiuo.mutation.Nickname(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: userinfo.FieldNickname,
		})
	}
	if value, ok := uiuo.mutation.Avatar(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: userinfo.FieldAvatar,
		})
	}
	if value, ok := uiuo.mutation.Email(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: userinfo.FieldEmail,
		})
	}
	if value, ok := uiuo.mutation.Phone(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: userinfo.FieldPhone,
		})
	}
	_node = &UserInfo{config: uiuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, uiuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{userinfo.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return nil, err
	}
	return _node, nil
}