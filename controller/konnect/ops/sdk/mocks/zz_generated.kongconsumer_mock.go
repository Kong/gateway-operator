// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	components "github.com/Kong/sdk-konnect-go/models/components"

	mock "github.com/stretchr/testify/mock"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
)

// MockConsumersSDK is an autogenerated mock type for the ConsumersSDK type
type MockConsumersSDK struct {
	mock.Mock
}

type MockConsumersSDK_Expecter struct {
	mock *mock.Mock
}

func (_m *MockConsumersSDK) EXPECT() *MockConsumersSDK_Expecter {
	return &MockConsumersSDK_Expecter{mock: &_m.Mock}
}

// CreateConsumer provides a mock function with given fields: ctx, controlPlaneID, consumerInput, opts
func (_m *MockConsumersSDK) CreateConsumer(ctx context.Context, controlPlaneID string, consumerInput components.ConsumerInput, opts ...operations.Option) (*operations.CreateConsumerResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, consumerInput)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateConsumer")
	}

	var r0 *operations.CreateConsumerResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, components.ConsumerInput, ...operations.Option) (*operations.CreateConsumerResponse, error)); ok {
		return rf(ctx, controlPlaneID, consumerInput, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, components.ConsumerInput, ...operations.Option) *operations.CreateConsumerResponse); ok {
		r0 = rf(ctx, controlPlaneID, consumerInput, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.CreateConsumerResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, components.ConsumerInput, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, consumerInput, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConsumersSDK_CreateConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateConsumer'
type MockConsumersSDK_CreateConsumer_Call struct {
	*mock.Call
}

// CreateConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - consumerInput components.ConsumerInput
//   - opts ...operations.Option
func (_e *MockConsumersSDK_Expecter) CreateConsumer(ctx interface{}, controlPlaneID interface{}, consumerInput interface{}, opts ...interface{}) *MockConsumersSDK_CreateConsumer_Call {
	return &MockConsumersSDK_CreateConsumer_Call{Call: _e.mock.On("CreateConsumer",
		append([]interface{}{ctx, controlPlaneID, consumerInput}, opts...)...)}
}

func (_c *MockConsumersSDK_CreateConsumer_Call) Run(run func(ctx context.Context, controlPlaneID string, consumerInput components.ConsumerInput, opts ...operations.Option)) *MockConsumersSDK_CreateConsumer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(components.ConsumerInput), variadicArgs...)
	})
	return _c
}

func (_c *MockConsumersSDK_CreateConsumer_Call) Return(_a0 *operations.CreateConsumerResponse, _a1 error) *MockConsumersSDK_CreateConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConsumersSDK_CreateConsumer_Call) RunAndReturn(run func(context.Context, string, components.ConsumerInput, ...operations.Option) (*operations.CreateConsumerResponse, error)) *MockConsumersSDK_CreateConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteConsumer provides a mock function with given fields: ctx, controlPlaneID, consumerID, opts
func (_m *MockConsumersSDK) DeleteConsumer(ctx context.Context, controlPlaneID string, consumerID string, opts ...operations.Option) (*operations.DeleteConsumerResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, consumerID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteConsumer")
	}

	var r0 *operations.DeleteConsumerResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) (*operations.DeleteConsumerResponse, error)); ok {
		return rf(ctx, controlPlaneID, consumerID, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) *operations.DeleteConsumerResponse); ok {
		r0 = rf(ctx, controlPlaneID, consumerID, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.DeleteConsumerResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, consumerID, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConsumersSDK_DeleteConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteConsumer'
type MockConsumersSDK_DeleteConsumer_Call struct {
	*mock.Call
}

// DeleteConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - consumerID string
//   - opts ...operations.Option
func (_e *MockConsumersSDK_Expecter) DeleteConsumer(ctx interface{}, controlPlaneID interface{}, consumerID interface{}, opts ...interface{}) *MockConsumersSDK_DeleteConsumer_Call {
	return &MockConsumersSDK_DeleteConsumer_Call{Call: _e.mock.On("DeleteConsumer",
		append([]interface{}{ctx, controlPlaneID, consumerID}, opts...)...)}
}

func (_c *MockConsumersSDK_DeleteConsumer_Call) Run(run func(ctx context.Context, controlPlaneID string, consumerID string, opts ...operations.Option)) *MockConsumersSDK_DeleteConsumer_Call {
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

func (_c *MockConsumersSDK_DeleteConsumer_Call) Return(_a0 *operations.DeleteConsumerResponse, _a1 error) *MockConsumersSDK_DeleteConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConsumersSDK_DeleteConsumer_Call) RunAndReturn(run func(context.Context, string, string, ...operations.Option) (*operations.DeleteConsumerResponse, error)) *MockConsumersSDK_DeleteConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// ListConsumer provides a mock function with given fields: ctx, request, opts
func (_m *MockConsumersSDK) ListConsumer(ctx context.Context, request operations.ListConsumerRequest, opts ...operations.Option) (*operations.ListConsumerResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListConsumer")
	}

	var r0 *operations.ListConsumerResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListConsumerRequest, ...operations.Option) (*operations.ListConsumerResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListConsumerRequest, ...operations.Option) *operations.ListConsumerResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.ListConsumerResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.ListConsumerRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConsumersSDK_ListConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListConsumer'
type MockConsumersSDK_ListConsumer_Call struct {
	*mock.Call
}

// ListConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.ListConsumerRequest
//   - opts ...operations.Option
func (_e *MockConsumersSDK_Expecter) ListConsumer(ctx interface{}, request interface{}, opts ...interface{}) *MockConsumersSDK_ListConsumer_Call {
	return &MockConsumersSDK_ListConsumer_Call{Call: _e.mock.On("ListConsumer",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockConsumersSDK_ListConsumer_Call) Run(run func(ctx context.Context, request operations.ListConsumerRequest, opts ...operations.Option)) *MockConsumersSDK_ListConsumer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.ListConsumerRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockConsumersSDK_ListConsumer_Call) Return(_a0 *operations.ListConsumerResponse, _a1 error) *MockConsumersSDK_ListConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConsumersSDK_ListConsumer_Call) RunAndReturn(run func(context.Context, operations.ListConsumerRequest, ...operations.Option) (*operations.ListConsumerResponse, error)) *MockConsumersSDK_ListConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertConsumer provides a mock function with given fields: ctx, upsertConsumerRequest, opts
func (_m *MockConsumersSDK) UpsertConsumer(ctx context.Context, upsertConsumerRequest operations.UpsertConsumerRequest, opts ...operations.Option) (*operations.UpsertConsumerResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, upsertConsumerRequest)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpsertConsumer")
	}

	var r0 *operations.UpsertConsumerResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertConsumerRequest, ...operations.Option) (*operations.UpsertConsumerResponse, error)); ok {
		return rf(ctx, upsertConsumerRequest, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertConsumerRequest, ...operations.Option) *operations.UpsertConsumerResponse); ok {
		r0 = rf(ctx, upsertConsumerRequest, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.UpsertConsumerResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.UpsertConsumerRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, upsertConsumerRequest, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConsumersSDK_UpsertConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertConsumer'
type MockConsumersSDK_UpsertConsumer_Call struct {
	*mock.Call
}

// UpsertConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - upsertConsumerRequest operations.UpsertConsumerRequest
//   - opts ...operations.Option
func (_e *MockConsumersSDK_Expecter) UpsertConsumer(ctx interface{}, upsertConsumerRequest interface{}, opts ...interface{}) *MockConsumersSDK_UpsertConsumer_Call {
	return &MockConsumersSDK_UpsertConsumer_Call{Call: _e.mock.On("UpsertConsumer",
		append([]interface{}{ctx, upsertConsumerRequest}, opts...)...)}
}

func (_c *MockConsumersSDK_UpsertConsumer_Call) Run(run func(ctx context.Context, upsertConsumerRequest operations.UpsertConsumerRequest, opts ...operations.Option)) *MockConsumersSDK_UpsertConsumer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.UpsertConsumerRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockConsumersSDK_UpsertConsumer_Call) Return(_a0 *operations.UpsertConsumerResponse, _a1 error) *MockConsumersSDK_UpsertConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConsumersSDK_UpsertConsumer_Call) RunAndReturn(run func(context.Context, operations.UpsertConsumerRequest, ...operations.Option) (*operations.UpsertConsumerResponse, error)) *MockConsumersSDK_UpsertConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockConsumersSDK creates a new instance of MockConsumersSDK. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockConsumersSDK(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConsumersSDK {
	mock := &MockConsumersSDK{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
