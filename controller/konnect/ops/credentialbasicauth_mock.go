// Code generated by mockery. DO NOT EDIT.

package ops

import (
	context "context"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
	mock "github.com/stretchr/testify/mock"
)

// MockCredentialBasicAuthSDK is an autogenerated mock type for the CredentialBasicAuthSDK type
type MockCredentialBasicAuthSDK struct {
	mock.Mock
}

type MockCredentialBasicAuthSDK_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCredentialBasicAuthSDK) EXPECT() *MockCredentialBasicAuthSDK_Expecter {
	return &MockCredentialBasicAuthSDK_Expecter{mock: &_m.Mock}
}

// CreateBasicAuthWithConsumer provides a mock function with given fields: ctx, req, opts
func (_m *MockCredentialBasicAuthSDK) CreateBasicAuthWithConsumer(ctx context.Context, req operations.CreateBasicAuthWithConsumerRequest, opts ...operations.Option) (*operations.CreateBasicAuthWithConsumerResponse, error) {
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

// MockCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateBasicAuthWithConsumer'
type MockCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call struct {
	*mock.Call
}

// CreateBasicAuthWithConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - req operations.CreateBasicAuthWithConsumerRequest
//   - opts ...operations.Option
func (_e *MockCredentialBasicAuthSDK_Expecter) CreateBasicAuthWithConsumer(ctx interface{}, req interface{}, opts ...interface{}) *MockCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call {
	return &MockCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call{Call: _e.mock.On("CreateBasicAuthWithConsumer",
		append([]interface{}{ctx, req}, opts...)...)}
}

func (_c *MockCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call) Run(run func(ctx context.Context, req operations.CreateBasicAuthWithConsumerRequest, opts ...operations.Option)) *MockCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call {
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

func (_c *MockCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call) Return(_a0 *operations.CreateBasicAuthWithConsumerResponse, _a1 error) *MockCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call) RunAndReturn(run func(context.Context, operations.CreateBasicAuthWithConsumerRequest, ...operations.Option) (*operations.CreateBasicAuthWithConsumerResponse, error)) *MockCredentialBasicAuthSDK_CreateBasicAuthWithConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteBasicAuthWithConsumer provides a mock function with given fields: ctx, request, opts
func (_m *MockCredentialBasicAuthSDK) DeleteBasicAuthWithConsumer(ctx context.Context, request operations.DeleteBasicAuthWithConsumerRequest, opts ...operations.Option) (*operations.DeleteBasicAuthWithConsumerResponse, error) {
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

// MockCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteBasicAuthWithConsumer'
type MockCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call struct {
	*mock.Call
}

// DeleteBasicAuthWithConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.DeleteBasicAuthWithConsumerRequest
//   - opts ...operations.Option
func (_e *MockCredentialBasicAuthSDK_Expecter) DeleteBasicAuthWithConsumer(ctx interface{}, request interface{}, opts ...interface{}) *MockCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call {
	return &MockCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call{Call: _e.mock.On("DeleteBasicAuthWithConsumer",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call) Run(run func(ctx context.Context, request operations.DeleteBasicAuthWithConsumerRequest, opts ...operations.Option)) *MockCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call {
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

func (_c *MockCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call) Return(_a0 *operations.DeleteBasicAuthWithConsumerResponse, _a1 error) *MockCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call) RunAndReturn(run func(context.Context, operations.DeleteBasicAuthWithConsumerRequest, ...operations.Option) (*operations.DeleteBasicAuthWithConsumerResponse, error)) *MockCredentialBasicAuthSDK_DeleteBasicAuthWithConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertBasicAuthWithConsumer provides a mock function with given fields: ctx, request, opts
func (_m *MockCredentialBasicAuthSDK) UpsertBasicAuthWithConsumer(ctx context.Context, request operations.UpsertBasicAuthWithConsumerRequest, opts ...operations.Option) (*operations.UpsertBasicAuthWithConsumerResponse, error) {
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

// MockCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertBasicAuthWithConsumer'
type MockCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call struct {
	*mock.Call
}

// UpsertBasicAuthWithConsumer is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.UpsertBasicAuthWithConsumerRequest
//   - opts ...operations.Option
func (_e *MockCredentialBasicAuthSDK_Expecter) UpsertBasicAuthWithConsumer(ctx interface{}, request interface{}, opts ...interface{}) *MockCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call {
	return &MockCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call{Call: _e.mock.On("UpsertBasicAuthWithConsumer",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call) Run(run func(ctx context.Context, request operations.UpsertBasicAuthWithConsumerRequest, opts ...operations.Option)) *MockCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call {
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

func (_c *MockCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call) Return(_a0 *operations.UpsertBasicAuthWithConsumerResponse, _a1 error) *MockCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call) RunAndReturn(run func(context.Context, operations.UpsertBasicAuthWithConsumerRequest, ...operations.Option) (*operations.UpsertBasicAuthWithConsumerResponse, error)) *MockCredentialBasicAuthSDK_UpsertBasicAuthWithConsumer_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockCredentialBasicAuthSDK creates a new instance of MockCredentialBasicAuthSDK. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCredentialBasicAuthSDK(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCredentialBasicAuthSDK {
	mock := &MockCredentialBasicAuthSDK{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
