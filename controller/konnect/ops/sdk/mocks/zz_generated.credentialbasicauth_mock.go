// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
	mock "github.com/stretchr/testify/mock"
)

// MockKongCredentialBasicAuthSDK is an autogenerated mock type for the KongCredentialBasicAuthSDK type
type MockKongCredentialBasicAuthSDK struct {
	mock.Mock
}

type MockKongCredentialBasicAuthSDK_Expecter struct {
	mock *mock.Mock
}

func (_m *MockKongCredentialBasicAuthSDK) EXPECT() *MockKongCredentialBasicAuthSDK_Expecter {
	return &MockKongCredentialBasicAuthSDK_Expecter{mock: &_m.Mock}
}

// CreateBasicAuthWithConsumer provides a mock function with given fields: ctx, req, opts
func (_m *MockKongCredentialBasicAuthSDK) CreateBasicAuthWithConsumer(ctx context.Context, req operations.CreateBasicAuthWithConsumerRequest, opts ...operations.Option) (*operations.CreateBasicAuthWithConsumerResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, req)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateBasicAuthWithConsumer")
	}

	var r0 *operations.CreateBasicAuthWithConsumerResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.CreateBasicAuthWithConsumerRequest, ...operations.Option) (*operations.CreateBasicAuthWithConsumerResponse, error)); ok {
		return rf(ctx, req, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.CreateBasicAuthWithConsumerRequest, ...operations.Option) *operations.CreateBasicAuthWithConsumerResponse); ok {
		r0 = rf(ctx, req, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.CreateBasicAuthWithConsumerResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.CreateBasicAuthWithConsumerRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, req, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKongCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateBasicAuthWithConsumer'
type MockKongCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call struct {
	*mock.Call
}

// CreateBasicAuthWithConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - req operations.CreateBasicAuthWithConsumerRequest
//   - opts ...operations.Option
func (_e *MockKongCredentialBasicAuthSDK_Expecter) CreateBasicAuthWithConsumer(ctx interface{}, req interface{}, opts ...interface{}) *MockKongCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call {
	return &MockKongCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call{Call: _e.mock.On("CreateBasicAuthWithConsumer",
		append([]interface{}{ctx, req}, opts...)...)}
}

func (_c *MockKongCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call) Run(run func(ctx context.Context, req operations.CreateBasicAuthWithConsumerRequest, opts ...operations.Option)) *MockKongCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.CreateBasicAuthWithConsumerRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockKongCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call) Return(_a0 *operations.CreateBasicAuthWithConsumerResponse, _a1 error) *MockKongCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKongCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call) RunAndReturn(run func(context.Context, operations.CreateBasicAuthWithConsumerRequest, ...operations.Option) (*operations.CreateBasicAuthWithConsumerResponse, error)) *MockKongCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteBasicAuthWithConsumer provides a mock function with given fields: ctx, request, opts
func (_m *MockKongCredentialBasicAuthSDK) DeleteBasicAuthWithConsumer(ctx context.Context, request operations.DeleteBasicAuthWithConsumerRequest, opts ...operations.Option) (*operations.DeleteBasicAuthWithConsumerResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteBasicAuthWithConsumer")
	}

	var r0 *operations.DeleteBasicAuthWithConsumerResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.DeleteBasicAuthWithConsumerRequest, ...operations.Option) (*operations.DeleteBasicAuthWithConsumerResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.DeleteBasicAuthWithConsumerRequest, ...operations.Option) *operations.DeleteBasicAuthWithConsumerResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.DeleteBasicAuthWithConsumerResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.DeleteBasicAuthWithConsumerRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKongCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteBasicAuthWithConsumer'
type MockKongCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call struct {
	*mock.Call
}

// DeleteBasicAuthWithConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.DeleteBasicAuthWithConsumerRequest
//   - opts ...operations.Option
func (_e *MockKongCredentialBasicAuthSDK_Expecter) DeleteBasicAuthWithConsumer(ctx interface{}, request interface{}, opts ...interface{}) *MockKongCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call {
	return &MockKongCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call{Call: _e.mock.On("DeleteBasicAuthWithConsumer",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockKongCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call) Run(run func(ctx context.Context, request operations.DeleteBasicAuthWithConsumerRequest, opts ...operations.Option)) *MockKongCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.DeleteBasicAuthWithConsumerRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockKongCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call) Return(_a0 *operations.DeleteBasicAuthWithConsumerResponse, _a1 error) *MockKongCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKongCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call) RunAndReturn(run func(context.Context, operations.DeleteBasicAuthWithConsumerRequest, ...operations.Option) (*operations.DeleteBasicAuthWithConsumerResponse, error)) *MockKongCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// ListBasicAuth provides a mock function with given fields: ctx, request, opts
func (_m *MockKongCredentialBasicAuthSDK) ListBasicAuth(ctx context.Context, request operations.ListBasicAuthRequest, opts ...operations.Option) (*operations.ListBasicAuthResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListBasicAuth")
	}

	var r0 *operations.ListBasicAuthResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListBasicAuthRequest, ...operations.Option) (*operations.ListBasicAuthResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListBasicAuthRequest, ...operations.Option) *operations.ListBasicAuthResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.ListBasicAuthResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.ListBasicAuthRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKongCredentialBasicAuthSDK_ListBasicAuth_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListBasicAuth'
type MockKongCredentialBasicAuthSDK_ListBasicAuth_Call struct {
	*mock.Call
}

// ListBasicAuth is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.ListBasicAuthRequest
//   - opts ...operations.Option
func (_e *MockKongCredentialBasicAuthSDK_Expecter) ListBasicAuth(ctx interface{}, request interface{}, opts ...interface{}) *MockKongCredentialBasicAuthSDK_ListBasicAuth_Call {
	return &MockKongCredentialBasicAuthSDK_ListBasicAuth_Call{Call: _e.mock.On("ListBasicAuth",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockKongCredentialBasicAuthSDK_ListBasicAuth_Call) Run(run func(ctx context.Context, request operations.ListBasicAuthRequest, opts ...operations.Option)) *MockKongCredentialBasicAuthSDK_ListBasicAuth_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.ListBasicAuthRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockKongCredentialBasicAuthSDK_ListBasicAuth_Call) Return(_a0 *operations.ListBasicAuthResponse, _a1 error) *MockKongCredentialBasicAuthSDK_ListBasicAuth_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKongCredentialBasicAuthSDK_ListBasicAuth_Call) RunAndReturn(run func(context.Context, operations.ListBasicAuthRequest, ...operations.Option) (*operations.ListBasicAuthResponse, error)) *MockKongCredentialBasicAuthSDK_ListBasicAuth_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertBasicAuthWithConsumer provides a mock function with given fields: ctx, request, opts
func (_m *MockKongCredentialBasicAuthSDK) UpsertBasicAuthWithConsumer(ctx context.Context, request operations.UpsertBasicAuthWithConsumerRequest, opts ...operations.Option) (*operations.UpsertBasicAuthWithConsumerResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpsertBasicAuthWithConsumer")
	}

	var r0 *operations.UpsertBasicAuthWithConsumerResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertBasicAuthWithConsumerRequest, ...operations.Option) (*operations.UpsertBasicAuthWithConsumerResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertBasicAuthWithConsumerRequest, ...operations.Option) *operations.UpsertBasicAuthWithConsumerResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.UpsertBasicAuthWithConsumerResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.UpsertBasicAuthWithConsumerRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKongCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertBasicAuthWithConsumer'
type MockKongCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call struct {
	*mock.Call
}

// UpsertBasicAuthWithConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.UpsertBasicAuthWithConsumerRequest
//   - opts ...operations.Option
func (_e *MockKongCredentialBasicAuthSDK_Expecter) UpsertBasicAuthWithConsumer(ctx interface{}, request interface{}, opts ...interface{}) *MockKongCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call {
	return &MockKongCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call{Call: _e.mock.On("UpsertBasicAuthWithConsumer",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockKongCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call) Run(run func(ctx context.Context, request operations.UpsertBasicAuthWithConsumerRequest, opts ...operations.Option)) *MockKongCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.UpsertBasicAuthWithConsumerRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockKongCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call) Return(_a0 *operations.UpsertBasicAuthWithConsumerResponse, _a1 error) *MockKongCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKongCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call) RunAndReturn(run func(context.Context, operations.UpsertBasicAuthWithConsumerRequest, ...operations.Option) (*operations.UpsertBasicAuthWithConsumerResponse, error)) *MockKongCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockKongCredentialBasicAuthSDK creates a new instance of MockKongCredentialBasicAuthSDK. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockKongCredentialBasicAuthSDK(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockKongCredentialBasicAuthSDK {
	mock := &MockKongCredentialBasicAuthSDK{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}