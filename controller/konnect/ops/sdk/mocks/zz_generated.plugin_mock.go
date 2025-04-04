// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	components "github.com/Kong/sdk-konnect-go/models/components"

	mock "github.com/stretchr/testify/mock"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
)

// MockPluginSDK is an autogenerated mock type for the PluginSDK type
type MockPluginSDK struct {
	mock.Mock
}

type MockPluginSDK_Expecter struct {
	mock *mock.Mock
}

func (_m *MockPluginSDK) EXPECT() *MockPluginSDK_Expecter {
	return &MockPluginSDK_Expecter{mock: &_m.Mock}
}

// CreatePlugin provides a mock function with given fields: ctx, controlPlaneID, plugin, opts
func (_m *MockPluginSDK) CreatePlugin(ctx context.Context, controlPlaneID string, plugin components.PluginInput, opts ...operations.Option) (*operations.CreatePluginResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, plugin)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreatePlugin")
	}

	var r0 *operations.CreatePluginResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, components.PluginInput, ...operations.Option) (*operations.CreatePluginResponse, error)); ok {
		return rf(ctx, controlPlaneID, plugin, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, components.PluginInput, ...operations.Option) *operations.CreatePluginResponse); ok {
		r0 = rf(ctx, controlPlaneID, plugin, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.CreatePluginResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, components.PluginInput, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, plugin, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockPluginSDK_CreatePlugin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreatePlugin'
type MockPluginSDK_CreatePlugin_Call struct {
	*mock.Call
}

// CreatePlugin is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - plugin components.PluginInput
//   - opts ...operations.Option
func (_e *MockPluginSDK_Expecter) CreatePlugin(ctx interface{}, controlPlaneID interface{}, plugin interface{}, opts ...interface{}) *MockPluginSDK_CreatePlugin_Call {
	return &MockPluginSDK_CreatePlugin_Call{Call: _e.mock.On("CreatePlugin",
		append([]interface{}{ctx, controlPlaneID, plugin}, opts...)...)}
}

func (_c *MockPluginSDK_CreatePlugin_Call) Run(run func(ctx context.Context, controlPlaneID string, plugin components.PluginInput, opts ...operations.Option)) *MockPluginSDK_CreatePlugin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(components.PluginInput), variadicArgs...)
	})
	return _c
}

func (_c *MockPluginSDK_CreatePlugin_Call) Return(_a0 *operations.CreatePluginResponse, _a1 error) *MockPluginSDK_CreatePlugin_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPluginSDK_CreatePlugin_Call) RunAndReturn(run func(context.Context, string, components.PluginInput, ...operations.Option) (*operations.CreatePluginResponse, error)) *MockPluginSDK_CreatePlugin_Call {
	_c.Call.Return(run)
	return _c
}

// DeletePlugin provides a mock function with given fields: ctx, controlPlaneID, pluginID, opts
func (_m *MockPluginSDK) DeletePlugin(ctx context.Context, controlPlaneID string, pluginID string, opts ...operations.Option) (*operations.DeletePluginResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, pluginID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeletePlugin")
	}

	var r0 *operations.DeletePluginResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) (*operations.DeletePluginResponse, error)); ok {
		return rf(ctx, controlPlaneID, pluginID, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) *operations.DeletePluginResponse); ok {
		r0 = rf(ctx, controlPlaneID, pluginID, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.DeletePluginResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, pluginID, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockPluginSDK_DeletePlugin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeletePlugin'
type MockPluginSDK_DeletePlugin_Call struct {
	*mock.Call
}

// DeletePlugin is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - pluginID string
//   - opts ...operations.Option
func (_e *MockPluginSDK_Expecter) DeletePlugin(ctx interface{}, controlPlaneID interface{}, pluginID interface{}, opts ...interface{}) *MockPluginSDK_DeletePlugin_Call {
	return &MockPluginSDK_DeletePlugin_Call{Call: _e.mock.On("DeletePlugin",
		append([]interface{}{ctx, controlPlaneID, pluginID}, opts...)...)}
}

func (_c *MockPluginSDK_DeletePlugin_Call) Run(run func(ctx context.Context, controlPlaneID string, pluginID string, opts ...operations.Option)) *MockPluginSDK_DeletePlugin_Call {
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

func (_c *MockPluginSDK_DeletePlugin_Call) Return(_a0 *operations.DeletePluginResponse, _a1 error) *MockPluginSDK_DeletePlugin_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPluginSDK_DeletePlugin_Call) RunAndReturn(run func(context.Context, string, string, ...operations.Option) (*operations.DeletePluginResponse, error)) *MockPluginSDK_DeletePlugin_Call {
	_c.Call.Return(run)
	return _c
}

// ListPlugin provides a mock function with given fields: ctx, request, opts
func (_m *MockPluginSDK) ListPlugin(ctx context.Context, request operations.ListPluginRequest, opts ...operations.Option) (*operations.ListPluginResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListPlugin")
	}

	var r0 *operations.ListPluginResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListPluginRequest, ...operations.Option) (*operations.ListPluginResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListPluginRequest, ...operations.Option) *operations.ListPluginResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.ListPluginResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.ListPluginRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockPluginSDK_ListPlugin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListPlugin'
type MockPluginSDK_ListPlugin_Call struct {
	*mock.Call
}

// ListPlugin is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.ListPluginRequest
//   - opts ...operations.Option
func (_e *MockPluginSDK_Expecter) ListPlugin(ctx interface{}, request interface{}, opts ...interface{}) *MockPluginSDK_ListPlugin_Call {
	return &MockPluginSDK_ListPlugin_Call{Call: _e.mock.On("ListPlugin",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockPluginSDK_ListPlugin_Call) Run(run func(ctx context.Context, request operations.ListPluginRequest, opts ...operations.Option)) *MockPluginSDK_ListPlugin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.ListPluginRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockPluginSDK_ListPlugin_Call) Return(_a0 *operations.ListPluginResponse, _a1 error) *MockPluginSDK_ListPlugin_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPluginSDK_ListPlugin_Call) RunAndReturn(run func(context.Context, operations.ListPluginRequest, ...operations.Option) (*operations.ListPluginResponse, error)) *MockPluginSDK_ListPlugin_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertPlugin provides a mock function with given fields: ctx, request, opts
func (_m *MockPluginSDK) UpsertPlugin(ctx context.Context, request operations.UpsertPluginRequest, opts ...operations.Option) (*operations.UpsertPluginResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpsertPlugin")
	}

	var r0 *operations.UpsertPluginResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertPluginRequest, ...operations.Option) (*operations.UpsertPluginResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertPluginRequest, ...operations.Option) *operations.UpsertPluginResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.UpsertPluginResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.UpsertPluginRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockPluginSDK_UpsertPlugin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertPlugin'
type MockPluginSDK_UpsertPlugin_Call struct {
	*mock.Call
}

// UpsertPlugin is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.UpsertPluginRequest
//   - opts ...operations.Option
func (_e *MockPluginSDK_Expecter) UpsertPlugin(ctx interface{}, request interface{}, opts ...interface{}) *MockPluginSDK_UpsertPlugin_Call {
	return &MockPluginSDK_UpsertPlugin_Call{Call: _e.mock.On("UpsertPlugin",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockPluginSDK_UpsertPlugin_Call) Run(run func(ctx context.Context, request operations.UpsertPluginRequest, opts ...operations.Option)) *MockPluginSDK_UpsertPlugin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.UpsertPluginRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockPluginSDK_UpsertPlugin_Call) Return(_a0 *operations.UpsertPluginResponse, _a1 error) *MockPluginSDK_UpsertPlugin_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPluginSDK_UpsertPlugin_Call) RunAndReturn(run func(context.Context, operations.UpsertPluginRequest, ...operations.Option) (*operations.UpsertPluginResponse, error)) *MockPluginSDK_UpsertPlugin_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockPluginSDK creates a new instance of MockPluginSDK. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPluginSDK(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPluginSDK {
	mock := &MockPluginSDK{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
