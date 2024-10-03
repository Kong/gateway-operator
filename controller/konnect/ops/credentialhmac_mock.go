// Code generated by mockery. DO NOT EDIT.

package ops

import (
	context "context"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
	mock "github.com/stretchr/testify/mock"
)

// MockKongCredentialHMACSDK is an autogenerated mock type for the KongCredentialHMACSDK type
type MockKongCredentialHMACSDK struct {
	mock.Mock
}

type MockKongCredentialHMACSDK_Expecter struct {
	mock *mock.Mock
}

func (_m *MockKongCredentialHMACSDK) EXPECT() *MockKongCredentialHMACSDK_Expecter {
	return &MockKongCredentialHMACSDK_Expecter{mock: &_m.Mock}
}

// CreateHmacAuthWithConsumer provides a mock function with given fields: ctx, req, opts
func (_m *MockKongCredentialHMACSDK) CreateHmacAuthWithConsumer(ctx context.Context, req operations.CreateHmacAuthWithConsumerRequest, opts ...operations.Option) (*operations.CreateHmacAuthWithConsumerResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, req)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateHmacAuthWithConsumer")
	}

	var r0 *operations.CreateHmacAuthWithConsumerResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.CreateHmacAuthWithConsumerRequest, ...operations.Option) (*operations.CreateHmacAuthWithConsumerResponse, error)); ok {
		return rf(ctx, req, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.CreateHmacAuthWithConsumerRequest, ...operations.Option) *operations.CreateHmacAuthWithConsumerResponse); ok {
		r0 = rf(ctx, req, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.CreateHmacAuthWithConsumerResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.CreateHmacAuthWithConsumerRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, req, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKongCredentialHMACSDK_CreateHmacAuthWithConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateHmacAuthWithConsumer'
type MockKongCredentialHMACSDK_CreateHmacAuthWithConsumer_Call struct {
	*mock.Call
}

// CreateHmacAuthWithConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - req operations.CreateHmacAuthWithConsumerRequest
//   - opts ...operations.Option
func (_e *MockKongCredentialHMACSDK_Expecter) CreateHmacAuthWithConsumer(ctx interface{}, req interface{}, opts ...interface{}) *MockKongCredentialHMACSDK_CreateHmacAuthWithConsumer_Call {
	return &MockKongCredentialHMACSDK_CreateHmacAuthWithConsumer_Call{Call: _e.mock.On("CreateHmacAuthWithConsumer",
		append([]interface{}{ctx, req}, opts...)...)}
}

func (_c *MockKongCredentialHMACSDK_CreateHmacAuthWithConsumer_Call) Run(run func(ctx context.Context, req operations.CreateHmacAuthWithConsumerRequest, opts ...operations.Option)) *MockKongCredentialHMACSDK_CreateHmacAuthWithConsumer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.CreateHmacAuthWithConsumerRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockKongCredentialHMACSDK_CreateHmacAuthWithConsumer_Call) Return(_a0 *operations.CreateHmacAuthWithConsumerResponse, _a1 error) *MockKongCredentialHMACSDK_CreateHmacAuthWithConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKongCredentialHMACSDK_CreateHmacAuthWithConsumer_Call) RunAndReturn(run func(context.Context, operations.CreateHmacAuthWithConsumerRequest, ...operations.Option) (*operations.CreateHmacAuthWithConsumerResponse, error)) *MockKongCredentialHMACSDK_CreateHmacAuthWithConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteHmacAuthWithConsumer provides a mock function with given fields: ctx, request, opts
func (_m *MockKongCredentialHMACSDK) DeleteHmacAuthWithConsumer(ctx context.Context, request operations.DeleteHmacAuthWithConsumerRequest, opts ...operations.Option) (*operations.DeleteHmacAuthWithConsumerResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteHmacAuthWithConsumer")
	}

	var r0 *operations.DeleteHmacAuthWithConsumerResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.DeleteHmacAuthWithConsumerRequest, ...operations.Option) (*operations.DeleteHmacAuthWithConsumerResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.DeleteHmacAuthWithConsumerRequest, ...operations.Option) *operations.DeleteHmacAuthWithConsumerResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.DeleteHmacAuthWithConsumerResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.DeleteHmacAuthWithConsumerRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKongCredentialHMACSDK_DeleteHmacAuthWithConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteHmacAuthWithConsumer'
type MockKongCredentialHMACSDK_DeleteHmacAuthWithConsumer_Call struct {
	*mock.Call
}

// DeleteHmacAuthWithConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.DeleteHmacAuthWithConsumerRequest
//   - opts ...operations.Option
func (_e *MockKongCredentialHMACSDK_Expecter) DeleteHmacAuthWithConsumer(ctx interface{}, request interface{}, opts ...interface{}) *MockKongCredentialHMACSDK_DeleteHmacAuthWithConsumer_Call {
	return &MockKongCredentialHMACSDK_DeleteHmacAuthWithConsumer_Call{Call: _e.mock.On("DeleteHmacAuthWithConsumer",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockKongCredentialHMACSDK_DeleteHmacAuthWithConsumer_Call) Run(run func(ctx context.Context, request operations.DeleteHmacAuthWithConsumerRequest, opts ...operations.Option)) *MockKongCredentialHMACSDK_DeleteHmacAuthWithConsumer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.DeleteHmacAuthWithConsumerRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockKongCredentialHMACSDK_DeleteHmacAuthWithConsumer_Call) Return(_a0 *operations.DeleteHmacAuthWithConsumerResponse, _a1 error) *MockKongCredentialHMACSDK_DeleteHmacAuthWithConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKongCredentialHMACSDK_DeleteHmacAuthWithConsumer_Call) RunAndReturn(run func(context.Context, operations.DeleteHmacAuthWithConsumerRequest, ...operations.Option) (*operations.DeleteHmacAuthWithConsumerResponse, error)) *MockKongCredentialHMACSDK_DeleteHmacAuthWithConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertHmacAuthWithConsumer provides a mock function with given fields: ctx, request, opts
func (_m *MockKongCredentialHMACSDK) UpsertHmacAuthWithConsumer(ctx context.Context, request operations.UpsertHmacAuthWithConsumerRequest, opts ...operations.Option) (*operations.UpsertHmacAuthWithConsumerResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpsertHmacAuthWithConsumer")
	}

	var r0 *operations.UpsertHmacAuthWithConsumerResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertHmacAuthWithConsumerRequest, ...operations.Option) (*operations.UpsertHmacAuthWithConsumerResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertHmacAuthWithConsumerRequest, ...operations.Option) *operations.UpsertHmacAuthWithConsumerResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.UpsertHmacAuthWithConsumerResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.UpsertHmacAuthWithConsumerRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKongCredentialHMACSDK_UpsertHmacAuthWithConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertHmacAuthWithConsumer'
type MockKongCredentialHMACSDK_UpsertHmacAuthWithConsumer_Call struct {
	*mock.Call
}

// UpsertHmacAuthWithConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.UpsertHmacAuthWithConsumerRequest
//   - opts ...operations.Option
func (_e *MockKongCredentialHMACSDK_Expecter) UpsertHmacAuthWithConsumer(ctx interface{}, request interface{}, opts ...interface{}) *MockKongCredentialHMACSDK_UpsertHmacAuthWithConsumer_Call {
	return &MockKongCredentialHMACSDK_UpsertHmacAuthWithConsumer_Call{Call: _e.mock.On("UpsertHmacAuthWithConsumer",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockKongCredentialHMACSDK_UpsertHmacAuthWithConsumer_Call) Run(run func(ctx context.Context, request operations.UpsertHmacAuthWithConsumerRequest, opts ...operations.Option)) *MockKongCredentialHMACSDK_UpsertHmacAuthWithConsumer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.UpsertHmacAuthWithConsumerRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockKongCredentialHMACSDK_UpsertHmacAuthWithConsumer_Call) Return(_a0 *operations.UpsertHmacAuthWithConsumerResponse, _a1 error) *MockKongCredentialHMACSDK_UpsertHmacAuthWithConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKongCredentialHMACSDK_UpsertHmacAuthWithConsumer_Call) RunAndReturn(run func(context.Context, operations.UpsertHmacAuthWithConsumerRequest, ...operations.Option) (*operations.UpsertHmacAuthWithConsumerResponse, error)) *MockKongCredentialHMACSDK_UpsertHmacAuthWithConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockKongCredentialHMACSDK creates a new instance of MockKongCredentialHMACSDK. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockKongCredentialHMACSDK(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockKongCredentialHMACSDK {
	mock := &MockKongCredentialHMACSDK{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
