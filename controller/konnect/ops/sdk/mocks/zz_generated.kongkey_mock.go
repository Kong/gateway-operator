// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	components "github.com/Kong/sdk-konnect-go/models/components"

	mock "github.com/stretchr/testify/mock"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
)

// MockKeysSDK is an autogenerated mock type for the KeysSDK type
type MockKeysSDK struct {
	mock.Mock
}

type MockKeysSDK_Expecter struct {
	mock *mock.Mock
}

func (_m *MockKeysSDK) EXPECT() *MockKeysSDK_Expecter {
	return &MockKeysSDK_Expecter{mock: &_m.Mock}
}

// CreateKey provides a mock function with given fields: ctx, controlPlaneID, Key, opts
func (_m *MockKeysSDK) CreateKey(ctx context.Context, controlPlaneID string, Key components.KeyInput, opts ...operations.Option) (*operations.CreateKeyResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, Key)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateKey")
	}

	var r0 *operations.CreateKeyResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, components.KeyInput, ...operations.Option) (*operations.CreateKeyResponse, error)); ok {
		return rf(ctx, controlPlaneID, Key, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, components.KeyInput, ...operations.Option) *operations.CreateKeyResponse); ok {
		r0 = rf(ctx, controlPlaneID, Key, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.CreateKeyResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, components.KeyInput, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, Key, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKeysSDK_CreateKey_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateKey'
type MockKeysSDK_CreateKey_Call struct {
	*mock.Call
}

// CreateKey is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - Key components.KeyInput
//   - opts ...operations.Option
func (_e *MockKeysSDK_Expecter) CreateKey(ctx interface{}, controlPlaneID interface{}, Key interface{}, opts ...interface{}) *MockKeysSDK_CreateKey_Call {
	return &MockKeysSDK_CreateKey_Call{Call: _e.mock.On("CreateKey",
		append([]interface{}{ctx, controlPlaneID, Key}, opts...)...)}
}

func (_c *MockKeysSDK_CreateKey_Call) Run(run func(ctx context.Context, controlPlaneID string, Key components.KeyInput, opts ...operations.Option)) *MockKeysSDK_CreateKey_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(components.KeyInput), variadicArgs...)
	})
	return _c
}

func (_c *MockKeysSDK_CreateKey_Call) Return(_a0 *operations.CreateKeyResponse, _a1 error) *MockKeysSDK_CreateKey_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKeysSDK_CreateKey_Call) RunAndReturn(run func(context.Context, string, components.KeyInput, ...operations.Option) (*operations.CreateKeyResponse, error)) *MockKeysSDK_CreateKey_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteKey provides a mock function with given fields: ctx, controlPlaneID, KeyID, opts
func (_m *MockKeysSDK) DeleteKey(ctx context.Context, controlPlaneID string, KeyID string, opts ...operations.Option) (*operations.DeleteKeyResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, KeyID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteKey")
	}

	var r0 *operations.DeleteKeyResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) (*operations.DeleteKeyResponse, error)); ok {
		return rf(ctx, controlPlaneID, KeyID, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) *operations.DeleteKeyResponse); ok {
		r0 = rf(ctx, controlPlaneID, KeyID, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.DeleteKeyResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, KeyID, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKeysSDK_DeleteKey_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteKey'
type MockKeysSDK_DeleteKey_Call struct {
	*mock.Call
}

// DeleteKey is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - KeyID string
//   - opts ...operations.Option
func (_e *MockKeysSDK_Expecter) DeleteKey(ctx interface{}, controlPlaneID interface{}, KeyID interface{}, opts ...interface{}) *MockKeysSDK_DeleteKey_Call {
	return &MockKeysSDK_DeleteKey_Call{Call: _e.mock.On("DeleteKey",
		append([]interface{}{ctx, controlPlaneID, KeyID}, opts...)...)}
}

func (_c *MockKeysSDK_DeleteKey_Call) Run(run func(ctx context.Context, controlPlaneID string, KeyID string, opts ...operations.Option)) *MockKeysSDK_DeleteKey_Call {
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

func (_c *MockKeysSDK_DeleteKey_Call) Return(_a0 *operations.DeleteKeyResponse, _a1 error) *MockKeysSDK_DeleteKey_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKeysSDK_DeleteKey_Call) RunAndReturn(run func(context.Context, string, string, ...operations.Option) (*operations.DeleteKeyResponse, error)) *MockKeysSDK_DeleteKey_Call {
	_c.Call.Return(run)
	return _c
}

// ListKey provides a mock function with given fields: ctx, request, opts
func (_m *MockKeysSDK) ListKey(ctx context.Context, request operations.ListKeyRequest, opts ...operations.Option) (*operations.ListKeyResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListKey")
	}

	var r0 *operations.ListKeyResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListKeyRequest, ...operations.Option) (*operations.ListKeyResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListKeyRequest, ...operations.Option) *operations.ListKeyResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.ListKeyResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.ListKeyRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKeysSDK_ListKey_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListKey'
type MockKeysSDK_ListKey_Call struct {
	*mock.Call
}

// ListKey is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.ListKeyRequest
//   - opts ...operations.Option
func (_e *MockKeysSDK_Expecter) ListKey(ctx interface{}, request interface{}, opts ...interface{}) *MockKeysSDK_ListKey_Call {
	return &MockKeysSDK_ListKey_Call{Call: _e.mock.On("ListKey",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockKeysSDK_ListKey_Call) Run(run func(ctx context.Context, request operations.ListKeyRequest, opts ...operations.Option)) *MockKeysSDK_ListKey_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.ListKeyRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockKeysSDK_ListKey_Call) Return(_a0 *operations.ListKeyResponse, _a1 error) *MockKeysSDK_ListKey_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKeysSDK_ListKey_Call) RunAndReturn(run func(context.Context, operations.ListKeyRequest, ...operations.Option) (*operations.ListKeyResponse, error)) *MockKeysSDK_ListKey_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertKey provides a mock function with given fields: ctx, request, opts
func (_m *MockKeysSDK) UpsertKey(ctx context.Context, request operations.UpsertKeyRequest, opts ...operations.Option) (*operations.UpsertKeyResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpsertKey")
	}

	var r0 *operations.UpsertKeyResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertKeyRequest, ...operations.Option) (*operations.UpsertKeyResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertKeyRequest, ...operations.Option) *operations.UpsertKeyResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.UpsertKeyResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.UpsertKeyRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKeysSDK_UpsertKey_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertKey'
type MockKeysSDK_UpsertKey_Call struct {
	*mock.Call
}

// UpsertKey is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.UpsertKeyRequest
//   - opts ...operations.Option
func (_e *MockKeysSDK_Expecter) UpsertKey(ctx interface{}, request interface{}, opts ...interface{}) *MockKeysSDK_UpsertKey_Call {
	return &MockKeysSDK_UpsertKey_Call{Call: _e.mock.On("UpsertKey",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockKeysSDK_UpsertKey_Call) Run(run func(ctx context.Context, request operations.UpsertKeyRequest, opts ...operations.Option)) *MockKeysSDK_UpsertKey_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.UpsertKeyRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockKeysSDK_UpsertKey_Call) Return(_a0 *operations.UpsertKeyResponse, _a1 error) *MockKeysSDK_UpsertKey_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKeysSDK_UpsertKey_Call) RunAndReturn(run func(context.Context, operations.UpsertKeyRequest, ...operations.Option) (*operations.UpsertKeyResponse, error)) *MockKeysSDK_UpsertKey_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockKeysSDK creates a new instance of MockKeysSDK. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockKeysSDK(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockKeysSDK {
	mock := &MockKeysSDK{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
