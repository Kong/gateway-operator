// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	components "github.com/Kong/sdk-konnect-go/models/components"

	mock "github.com/stretchr/testify/mock"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
)

// MockVaultSDK is an autogenerated mock type for the VaultSDK type
type MockVaultSDK struct {
	mock.Mock
}

type MockVaultSDK_Expecter struct {
	mock *mock.Mock
}

func (_m *MockVaultSDK) EXPECT() *MockVaultSDK_Expecter {
	return &MockVaultSDK_Expecter{mock: &_m.Mock}
}

// CreateVault provides a mock function with given fields: ctx, controlPlaneID, vault, opts
func (_m *MockVaultSDK) CreateVault(ctx context.Context, controlPlaneID string, vault components.VaultInput, opts ...operations.Option) (*operations.CreateVaultResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, vault)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateVault")
	}

	var r0 *operations.CreateVaultResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, components.VaultInput, ...operations.Option) (*operations.CreateVaultResponse, error)); ok {
		return rf(ctx, controlPlaneID, vault, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, components.VaultInput, ...operations.Option) *operations.CreateVaultResponse); ok {
		r0 = rf(ctx, controlPlaneID, vault, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.CreateVaultResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, components.VaultInput, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, vault, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockVaultSDK_CreateVault_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateVault'
type MockVaultSDK_CreateVault_Call struct {
	*mock.Call
}

// CreateVault is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - vault components.VaultInput
//   - opts ...operations.Option
func (_e *MockVaultSDK_Expecter) CreateVault(ctx interface{}, controlPlaneID interface{}, vault interface{}, opts ...interface{}) *MockVaultSDK_CreateVault_Call {
	return &MockVaultSDK_CreateVault_Call{Call: _e.mock.On("CreateVault",
		append([]interface{}{ctx, controlPlaneID, vault}, opts...)...)}
}

func (_c *MockVaultSDK_CreateVault_Call) Run(run func(ctx context.Context, controlPlaneID string, vault components.VaultInput, opts ...operations.Option)) *MockVaultSDK_CreateVault_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(components.VaultInput), variadicArgs...)
	})
	return _c
}

func (_c *MockVaultSDK_CreateVault_Call) Return(_a0 *operations.CreateVaultResponse, _a1 error) *MockVaultSDK_CreateVault_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockVaultSDK_CreateVault_Call) RunAndReturn(run func(context.Context, string, components.VaultInput, ...operations.Option) (*operations.CreateVaultResponse, error)) *MockVaultSDK_CreateVault_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteVault provides a mock function with given fields: ctx, controlPlaneID, vaultID, opts
func (_m *MockVaultSDK) DeleteVault(ctx context.Context, controlPlaneID string, vaultID string, opts ...operations.Option) (*operations.DeleteVaultResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, vaultID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteVault")
	}

	var r0 *operations.DeleteVaultResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) (*operations.DeleteVaultResponse, error)); ok {
		return rf(ctx, controlPlaneID, vaultID, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) *operations.DeleteVaultResponse); ok {
		r0 = rf(ctx, controlPlaneID, vaultID, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.DeleteVaultResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, vaultID, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockVaultSDK_DeleteVault_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteVault'
type MockVaultSDK_DeleteVault_Call struct {
	*mock.Call
}

// DeleteVault is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - vaultID string
//   - opts ...operations.Option
func (_e *MockVaultSDK_Expecter) DeleteVault(ctx interface{}, controlPlaneID interface{}, vaultID interface{}, opts ...interface{}) *MockVaultSDK_DeleteVault_Call {
	return &MockVaultSDK_DeleteVault_Call{Call: _e.mock.On("DeleteVault",
		append([]interface{}{ctx, controlPlaneID, vaultID}, opts...)...)}
}

func (_c *MockVaultSDK_DeleteVault_Call) Run(run func(ctx context.Context, controlPlaneID string, vaultID string, opts ...operations.Option)) *MockVaultSDK_DeleteVault_Call {
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

func (_c *MockVaultSDK_DeleteVault_Call) Return(_a0 *operations.DeleteVaultResponse, _a1 error) *MockVaultSDK_DeleteVault_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockVaultSDK_DeleteVault_Call) RunAndReturn(run func(context.Context, string, string, ...operations.Option) (*operations.DeleteVaultResponse, error)) *MockVaultSDK_DeleteVault_Call {
	_c.Call.Return(run)
	return _c
}

// ListVault provides a mock function with given fields: ctx, request, opts
func (_m *MockVaultSDK) ListVault(ctx context.Context, request operations.ListVaultRequest, opts ...operations.Option) (*operations.ListVaultResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListVault")
	}

	var r0 *operations.ListVaultResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListVaultRequest, ...operations.Option) (*operations.ListVaultResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListVaultRequest, ...operations.Option) *operations.ListVaultResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.ListVaultResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.ListVaultRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockVaultSDK_ListVault_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListVault'
type MockVaultSDK_ListVault_Call struct {
	*mock.Call
}

// ListVault is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.ListVaultRequest
//   - opts ...operations.Option
func (_e *MockVaultSDK_Expecter) ListVault(ctx interface{}, request interface{}, opts ...interface{}) *MockVaultSDK_ListVault_Call {
	return &MockVaultSDK_ListVault_Call{Call: _e.mock.On("ListVault",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockVaultSDK_ListVault_Call) Run(run func(ctx context.Context, request operations.ListVaultRequest, opts ...operations.Option)) *MockVaultSDK_ListVault_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.ListVaultRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockVaultSDK_ListVault_Call) Return(_a0 *operations.ListVaultResponse, _a1 error) *MockVaultSDK_ListVault_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockVaultSDK_ListVault_Call) RunAndReturn(run func(context.Context, operations.ListVaultRequest, ...operations.Option) (*operations.ListVaultResponse, error)) *MockVaultSDK_ListVault_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertVault provides a mock function with given fields: ctx, request, opts
func (_m *MockVaultSDK) UpsertVault(ctx context.Context, request operations.UpsertVaultRequest, opts ...operations.Option) (*operations.UpsertVaultResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpsertVault")
	}

	var r0 *operations.UpsertVaultResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertVaultRequest, ...operations.Option) (*operations.UpsertVaultResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertVaultRequest, ...operations.Option) *operations.UpsertVaultResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.UpsertVaultResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.UpsertVaultRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockVaultSDK_UpsertVault_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertVault'
type MockVaultSDK_UpsertVault_Call struct {
	*mock.Call
}

// UpsertVault is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.UpsertVaultRequest
//   - opts ...operations.Option
func (_e *MockVaultSDK_Expecter) UpsertVault(ctx interface{}, request interface{}, opts ...interface{}) *MockVaultSDK_UpsertVault_Call {
	return &MockVaultSDK_UpsertVault_Call{Call: _e.mock.On("UpsertVault",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockVaultSDK_UpsertVault_Call) Run(run func(ctx context.Context, request operations.UpsertVaultRequest, opts ...operations.Option)) *MockVaultSDK_UpsertVault_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.UpsertVaultRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockVaultSDK_UpsertVault_Call) Return(_a0 *operations.UpsertVaultResponse, _a1 error) *MockVaultSDK_UpsertVault_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockVaultSDK_UpsertVault_Call) RunAndReturn(run func(context.Context, operations.UpsertVaultRequest, ...operations.Option) (*operations.UpsertVaultResponse, error)) *MockVaultSDK_UpsertVault_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockVaultSDK creates a new instance of MockVaultSDK. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockVaultSDK(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockVaultSDK {
	mock := &MockVaultSDK{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
