// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	components "github.com/Kong/sdk-konnect-go/models/components"

	mock "github.com/stretchr/testify/mock"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
)

// MockConsumerGroupSDK is an autogenerated mock type for the ConsumerGroupSDK type
type MockConsumerGroupSDK struct {
	mock.Mock
}

type MockConsumerGroupSDK_Expecter struct {
	mock *mock.Mock
}

func (_m *MockConsumerGroupSDK) EXPECT() *MockConsumerGroupSDK_Expecter {
	return &MockConsumerGroupSDK_Expecter{mock: &_m.Mock}
}

// AddConsumerToGroup provides a mock function with given fields: ctx, request, opts
func (_m *MockConsumerGroupSDK) AddConsumerToGroup(ctx context.Context, request operations.AddConsumerToGroupRequest, opts ...operations.Option) (*operations.AddConsumerToGroupResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for AddConsumerToGroup")
	}

	var r0 *operations.AddConsumerToGroupResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.AddConsumerToGroupRequest, ...operations.Option) (*operations.AddConsumerToGroupResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.AddConsumerToGroupRequest, ...operations.Option) *operations.AddConsumerToGroupResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.AddConsumerToGroupResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.AddConsumerToGroupRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConsumerGroupSDK_AddConsumerToGroup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddConsumerToGroup'
type MockConsumerGroupSDK_AddConsumerToGroup_Call struct {
	*mock.Call
}

// AddConsumerToGroup is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.AddConsumerToGroupRequest
//   - opts ...operations.Option
func (_e *MockConsumerGroupSDK_Expecter) AddConsumerToGroup(ctx interface{}, request interface{}, opts ...interface{}) *MockConsumerGroupSDK_AddConsumerToGroup_Call {
	return &MockConsumerGroupSDK_AddConsumerToGroup_Call{Call: _e.mock.On("AddConsumerToGroup",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockConsumerGroupSDK_AddConsumerToGroup_Call) Run(run func(ctx context.Context, request operations.AddConsumerToGroupRequest, opts ...operations.Option)) *MockConsumerGroupSDK_AddConsumerToGroup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.AddConsumerToGroupRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockConsumerGroupSDK_AddConsumerToGroup_Call) Return(_a0 *operations.AddConsumerToGroupResponse, _a1 error) *MockConsumerGroupSDK_AddConsumerToGroup_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConsumerGroupSDK_AddConsumerToGroup_Call) RunAndReturn(run func(context.Context, operations.AddConsumerToGroupRequest, ...operations.Option) (*operations.AddConsumerToGroupResponse, error)) *MockConsumerGroupSDK_AddConsumerToGroup_Call {
	_c.Call.Return(run)
	return _c
}

// CreateConsumerGroup provides a mock function with given fields: ctx, controlPlaneID, consumerInput, opts
func (_m *MockConsumerGroupSDK) CreateConsumerGroup(ctx context.Context, controlPlaneID string, consumerInput components.ConsumerGroupInput, opts ...operations.Option) (*operations.CreateConsumerGroupResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, consumerInput)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateConsumerGroup")
	}

	var r0 *operations.CreateConsumerGroupResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, components.ConsumerGroupInput, ...operations.Option) (*operations.CreateConsumerGroupResponse, error)); ok {
		return rf(ctx, controlPlaneID, consumerInput, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, components.ConsumerGroupInput, ...operations.Option) *operations.CreateConsumerGroupResponse); ok {
		r0 = rf(ctx, controlPlaneID, consumerInput, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.CreateConsumerGroupResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, components.ConsumerGroupInput, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, consumerInput, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConsumerGroupSDK_CreateConsumerGroup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateConsumerGroup'
type MockConsumerGroupSDK_CreateConsumerGroup_Call struct {
	*mock.Call
}

// CreateConsumerGroup is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - consumerInput components.ConsumerGroupInput
//   - opts ...operations.Option
func (_e *MockConsumerGroupSDK_Expecter) CreateConsumerGroup(ctx interface{}, controlPlaneID interface{}, consumerInput interface{}, opts ...interface{}) *MockConsumerGroupSDK_CreateConsumerGroup_Call {
	return &MockConsumerGroupSDK_CreateConsumerGroup_Call{Call: _e.mock.On("CreateConsumerGroup",
		append([]interface{}{ctx, controlPlaneID, consumerInput}, opts...)...)}
}

func (_c *MockConsumerGroupSDK_CreateConsumerGroup_Call) Run(run func(ctx context.Context, controlPlaneID string, consumerInput components.ConsumerGroupInput, opts ...operations.Option)) *MockConsumerGroupSDK_CreateConsumerGroup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(components.ConsumerGroupInput), variadicArgs...)
	})
	return _c
}

func (_c *MockConsumerGroupSDK_CreateConsumerGroup_Call) Return(_a0 *operations.CreateConsumerGroupResponse, _a1 error) *MockConsumerGroupSDK_CreateConsumerGroup_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConsumerGroupSDK_CreateConsumerGroup_Call) RunAndReturn(run func(context.Context, string, components.ConsumerGroupInput, ...operations.Option) (*operations.CreateConsumerGroupResponse, error)) *MockConsumerGroupSDK_CreateConsumerGroup_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteConsumerGroup provides a mock function with given fields: ctx, controlPlaneID, consumerID, opts
func (_m *MockConsumerGroupSDK) DeleteConsumerGroup(ctx context.Context, controlPlaneID string, consumerID string, opts ...operations.Option) (*operations.DeleteConsumerGroupResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, consumerID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteConsumerGroup")
	}

	var r0 *operations.DeleteConsumerGroupResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) (*operations.DeleteConsumerGroupResponse, error)); ok {
		return rf(ctx, controlPlaneID, consumerID, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) *operations.DeleteConsumerGroupResponse); ok {
		r0 = rf(ctx, controlPlaneID, consumerID, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.DeleteConsumerGroupResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, consumerID, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConsumerGroupSDK_DeleteConsumerGroup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteConsumerGroup'
type MockConsumerGroupSDK_DeleteConsumerGroup_Call struct {
	*mock.Call
}

// DeleteConsumerGroup is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - consumerID string
//   - opts ...operations.Option
func (_e *MockConsumerGroupSDK_Expecter) DeleteConsumerGroup(ctx interface{}, controlPlaneID interface{}, consumerID interface{}, opts ...interface{}) *MockConsumerGroupSDK_DeleteConsumerGroup_Call {
	return &MockConsumerGroupSDK_DeleteConsumerGroup_Call{Call: _e.mock.On("DeleteConsumerGroup",
		append([]interface{}{ctx, controlPlaneID, consumerID}, opts...)...)}
}

func (_c *MockConsumerGroupSDK_DeleteConsumerGroup_Call) Run(run func(ctx context.Context, controlPlaneID string, consumerID string, opts ...operations.Option)) *MockConsumerGroupSDK_DeleteConsumerGroup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(string), variadicArgs...)
	})
	return _c
}

func (_c *MockConsumerGroupSDK_DeleteConsumerGroup_Call) Return(_a0 *operations.DeleteConsumerGroupResponse, _a1 error) *MockConsumerGroupSDK_DeleteConsumerGroup_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConsumerGroupSDK_DeleteConsumerGroup_Call) RunAndReturn(run func(context.Context, string, string, ...operations.Option) (*operations.DeleteConsumerGroupResponse, error)) *MockConsumerGroupSDK_DeleteConsumerGroup_Call {
	_c.Call.Return(run)
	return _c
}

// ListConsumerGroup provides a mock function with given fields: ctx, request, opts
func (_m *MockConsumerGroupSDK) ListConsumerGroup(ctx context.Context, request operations.ListConsumerGroupRequest, opts ...operations.Option) (*operations.ListConsumerGroupResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListConsumerGroup")
	}

	var r0 *operations.ListConsumerGroupResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListConsumerGroupRequest, ...operations.Option) (*operations.ListConsumerGroupResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListConsumerGroupRequest, ...operations.Option) *operations.ListConsumerGroupResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.ListConsumerGroupResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.ListConsumerGroupRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConsumerGroupSDK_ListConsumerGroup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListConsumerGroup'
type MockConsumerGroupSDK_ListConsumerGroup_Call struct {
	*mock.Call
}

// ListConsumerGroup is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.ListConsumerGroupRequest
//   - opts ...operations.Option
func (_e *MockConsumerGroupSDK_Expecter) ListConsumerGroup(ctx interface{}, request interface{}, opts ...interface{}) *MockConsumerGroupSDK_ListConsumerGroup_Call {
	return &MockConsumerGroupSDK_ListConsumerGroup_Call{Call: _e.mock.On("ListConsumerGroup",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockConsumerGroupSDK_ListConsumerGroup_Call) Run(run func(ctx context.Context, request operations.ListConsumerGroupRequest, opts ...operations.Option)) *MockConsumerGroupSDK_ListConsumerGroup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.ListConsumerGroupRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockConsumerGroupSDK_ListConsumerGroup_Call) Return(_a0 *operations.ListConsumerGroupResponse, _a1 error) *MockConsumerGroupSDK_ListConsumerGroup_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConsumerGroupSDK_ListConsumerGroup_Call) RunAndReturn(run func(context.Context, operations.ListConsumerGroupRequest, ...operations.Option) (*operations.ListConsumerGroupResponse, error)) *MockConsumerGroupSDK_ListConsumerGroup_Call {
	_c.Call.Return(run)
	return _c
}

// ListConsumerGroupsForConsumer provides a mock function with given fields: ctx, request, opts
func (_m *MockConsumerGroupSDK) ListConsumerGroupsForConsumer(ctx context.Context, request operations.ListConsumerGroupsForConsumerRequest, opts ...operations.Option) (*operations.ListConsumerGroupsForConsumerResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListConsumerGroupsForConsumer")
	}

	var r0 *operations.ListConsumerGroupsForConsumerResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListConsumerGroupsForConsumerRequest, ...operations.Option) (*operations.ListConsumerGroupsForConsumerResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListConsumerGroupsForConsumerRequest, ...operations.Option) *operations.ListConsumerGroupsForConsumerResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.ListConsumerGroupsForConsumerResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.ListConsumerGroupsForConsumerRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConsumerGroupSDK_ListConsumerGroupsForConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListConsumerGroupsForConsumer'
type MockConsumerGroupSDK_ListConsumerGroupsForConsumer_Call struct {
	*mock.Call
}

// ListConsumerGroupsForConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.ListConsumerGroupsForConsumerRequest
//   - opts ...operations.Option
func (_e *MockConsumerGroupSDK_Expecter) ListConsumerGroupsForConsumer(ctx interface{}, request interface{}, opts ...interface{}) *MockConsumerGroupSDK_ListConsumerGroupsForConsumer_Call {
	return &MockConsumerGroupSDK_ListConsumerGroupsForConsumer_Call{Call: _e.mock.On("ListConsumerGroupsForConsumer",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockConsumerGroupSDK_ListConsumerGroupsForConsumer_Call) Run(run func(ctx context.Context, request operations.ListConsumerGroupsForConsumerRequest, opts ...operations.Option)) *MockConsumerGroupSDK_ListConsumerGroupsForConsumer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.ListConsumerGroupsForConsumerRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockConsumerGroupSDK_ListConsumerGroupsForConsumer_Call) Return(_a0 *operations.ListConsumerGroupsForConsumerResponse, _a1 error) *MockConsumerGroupSDK_ListConsumerGroupsForConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConsumerGroupSDK_ListConsumerGroupsForConsumer_Call) RunAndReturn(run func(context.Context, operations.ListConsumerGroupsForConsumerRequest, ...operations.Option) (*operations.ListConsumerGroupsForConsumerResponse, error)) *MockConsumerGroupSDK_ListConsumerGroupsForConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveConsumerFromGroup provides a mock function with given fields: ctx, request, opts
func (_m *MockConsumerGroupSDK) RemoveConsumerFromGroup(ctx context.Context, request operations.RemoveConsumerFromGroupRequest, opts ...operations.Option) (*operations.RemoveConsumerFromGroupResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for RemoveConsumerFromGroup")
	}

	var r0 *operations.RemoveConsumerFromGroupResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.RemoveConsumerFromGroupRequest, ...operations.Option) (*operations.RemoveConsumerFromGroupResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.RemoveConsumerFromGroupRequest, ...operations.Option) *operations.RemoveConsumerFromGroupResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.RemoveConsumerFromGroupResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.RemoveConsumerFromGroupRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConsumerGroupSDK_RemoveConsumerFromGroup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveConsumerFromGroup'
type MockConsumerGroupSDK_RemoveConsumerFromGroup_Call struct {
	*mock.Call
}

// RemoveConsumerFromGroup is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.RemoveConsumerFromGroupRequest
//   - opts ...operations.Option
func (_e *MockConsumerGroupSDK_Expecter) RemoveConsumerFromGroup(ctx interface{}, request interface{}, opts ...interface{}) *MockConsumerGroupSDK_RemoveConsumerFromGroup_Call {
	return &MockConsumerGroupSDK_RemoveConsumerFromGroup_Call{Call: _e.mock.On("RemoveConsumerFromGroup",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockConsumerGroupSDK_RemoveConsumerFromGroup_Call) Run(run func(ctx context.Context, request operations.RemoveConsumerFromGroupRequest, opts ...operations.Option)) *MockConsumerGroupSDK_RemoveConsumerFromGroup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.RemoveConsumerFromGroupRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockConsumerGroupSDK_RemoveConsumerFromGroup_Call) Return(_a0 *operations.RemoveConsumerFromGroupResponse, _a1 error) *MockConsumerGroupSDK_RemoveConsumerFromGroup_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConsumerGroupSDK_RemoveConsumerFromGroup_Call) RunAndReturn(run func(context.Context, operations.RemoveConsumerFromGroupRequest, ...operations.Option) (*operations.RemoveConsumerFromGroupResponse, error)) *MockConsumerGroupSDK_RemoveConsumerFromGroup_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertConsumerGroup provides a mock function with given fields: ctx, upsertConsumerRequest, opts
func (_m *MockConsumerGroupSDK) UpsertConsumerGroup(ctx context.Context, upsertConsumerRequest operations.UpsertConsumerGroupRequest, opts ...operations.Option) (*operations.UpsertConsumerGroupResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, upsertConsumerRequest)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpsertConsumerGroup")
	}

	var r0 *operations.UpsertConsumerGroupResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertConsumerGroupRequest, ...operations.Option) (*operations.UpsertConsumerGroupResponse, error)); ok {
		return rf(ctx, upsertConsumerRequest, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertConsumerGroupRequest, ...operations.Option) *operations.UpsertConsumerGroupResponse); ok {
		r0 = rf(ctx, upsertConsumerRequest, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.UpsertConsumerGroupResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.UpsertConsumerGroupRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, upsertConsumerRequest, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConsumerGroupSDK_UpsertConsumerGroup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertConsumerGroup'
type MockConsumerGroupSDK_UpsertConsumerGroup_Call struct {
	*mock.Call
}

// UpsertConsumerGroup is a helper method to define mock.On call
//   - ctx context.Context
//   - upsertConsumerRequest operations.UpsertConsumerGroupRequest
//   - opts ...operations.Option
func (_e *MockConsumerGroupSDK_Expecter) UpsertConsumerGroup(ctx interface{}, upsertConsumerRequest interface{}, opts ...interface{}) *MockConsumerGroupSDK_UpsertConsumerGroup_Call {
	return &MockConsumerGroupSDK_UpsertConsumerGroup_Call{Call: _e.mock.On("UpsertConsumerGroup",
		append([]interface{}{ctx, upsertConsumerRequest}, opts...)...)}
}

func (_c *MockConsumerGroupSDK_UpsertConsumerGroup_Call) Run(run func(ctx context.Context, upsertConsumerRequest operations.UpsertConsumerGroupRequest, opts ...operations.Option)) *MockConsumerGroupSDK_UpsertConsumerGroup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.UpsertConsumerGroupRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockConsumerGroupSDK_UpsertConsumerGroup_Call) Return(_a0 *operations.UpsertConsumerGroupResponse, _a1 error) *MockConsumerGroupSDK_UpsertConsumerGroup_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConsumerGroupSDK_UpsertConsumerGroup_Call) RunAndReturn(run func(context.Context, operations.UpsertConsumerGroupRequest, ...operations.Option) (*operations.UpsertConsumerGroupResponse, error)) *MockConsumerGroupSDK_UpsertConsumerGroup_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockConsumerGroupSDK creates a new instance of MockConsumerGroupSDK. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockConsumerGroupSDK(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConsumerGroupSDK {
	mock := &MockConsumerGroupSDK{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}