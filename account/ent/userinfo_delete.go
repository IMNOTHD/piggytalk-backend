// Code generated by entc, DO NOT EDIT.

package ent

import (
	"account/ent/predicate"
	"account/ent/userinfo"
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// UserInfoDelete is the builder for deleting a UserInfo entity.
type UserInfoDelete struct {
	config
	hooks    []Hook
	mutation *UserInfoMutation
}

// Where appends a list predicates to the UserInfoDelete builder.
func (uid *UserInfoDelete) Where(ps ...predicate.UserInfo) *UserInfoDelete {
	uid.mutation.Where(ps...)
	return uid
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (uid *UserInfoDelete) Exec(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	if len(uid.hooks) == 0 {
		affected, err = uid.sqlExec(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*UserInfoMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			uid.mutation = mutation
			affected, err = uid.sqlExec(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(uid.hooks) - 1; i >= 0; i-- {
			if uid.hooks[i] == nil {
				return 0, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = uid.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, uid.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// ExecX is like Exec, but panics if an error occurs.
func (uid *UserInfoDelete) ExecX(ctx context.Context) int {
	n, err := uid.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (uid *UserInfoDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := &sqlgraph.DeleteSpec{
		Node: &sqlgraph.NodeSpec{
			Table: userinfo.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: userinfo.FieldID,
			},
		},
	}
	if ps := uid.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return sqlgraph.DeleteNodes(ctx, uid.driver, _spec)
}

// UserInfoDeleteOne is the builder for deleting a single UserInfo entity.
type UserInfoDeleteOne struct {
	uid *UserInfoDelete
}

// Exec executes the deletion query.
func (uido *UserInfoDeleteOne) Exec(ctx context.Context) error {
	n, err := uido.uid.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{userinfo.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (uido *UserInfoDeleteOne) ExecX(ctx context.Context) {
	uido.uid.ExecX(ctx)
}
