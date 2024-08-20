// Code generated by mockery. DO NOT EDIT.

package ops

import (
	context "context"

	components "github.com/Kong/sdk-konnect-go/models/components"

	mock "github.com/stretchr/testify/mock"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
)

// MockControlPlaneSDK is an autogenerated mock type for the ControlPlaneSDK type
type MockControlPlaneSDK struct {
	mock.Mock
}

type MockControlPlaneSDK_Expecter struct {
	mock *mock.Mock
}

func (_m *MockControlPlaneSDK) EXPECT() *MockControlPlaneSDK_Expecter {
	return &MockControlPlaneSDK_Expecter{mock: &_m.Mock}
}

// CreateControlPlane provides a mock function with given fields: ctx, req, opts
func (_m *MockControlPlaneSDK) CreateControlPlane(ctx context.Context, req components.CreateControlPlaneRequest, opts ...operations.Option) (*operations.CreateControlPlaneResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, req)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateControlPlane")
	}

	var r0 *operations.CreateControlPlaneResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, components.CreateControlPlaneRequest, ...operations.Option) (*operations.CreateControlPlaneResponse, error)); ok {
		return rf(ctx, req, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, components.CreateControlPlaneRequest, ...operations.Option) *operations.CreateControlPlaneResponse); ok {
		r0 = rf(ctx, req, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.CreateControlPlaneResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, components.CreateControlPlaneRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, req, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockControlPlaneSDK_CreateControlPlane_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateControlPlane'
type MockControlPlaneSDK_CreateControlPlane_Call struct {
	*mock.Call
}

// CreateControlPlane is a helper method to define mock.On call
//   - ctx context.Context
//   - req components.CreateControlPlaneRequest
//   - opts ...operations.Option
func (_e *MockControlPlaneSDK_Expecter) CreateControlPlane(ctx interface{}, req interface{}, opts ...interface{}) *MockControlPlaneSDK_CreateControlPlane_Call {
	return &MockControlPlaneSDK_CreateControlPlane_Call{Call: _e.mock.On("CreateControlPlane",
		append([]interface{}{ctx, req}, opts...)...)}
}

func (_c *MockControlPlaneSDK_CreateControlPlane_Call) Run(run func(ctx context.Context, req components.CreateControlPlaneRequest, opts ...operations.Option)) *MockControlPlaneSDK_CreateControlPlane_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(components.CreateControlPlaneRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockControlPlaneSDK_CreateControlPlane_Call) Return(_a0 *operations.CreateControlPlaneResponse, _a1 error) *MockControlPlaneSDK_CreateControlPlane_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockControlPlaneSDK_CreateControlPlane_Call) RunAndReturn(run func(context.Context, components.CreateControlPlaneRequest, ...operations.Option) (*operations.CreateControlPlaneResponse, error)) *MockControlPlaneSDK_CreateControlPlane_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteControlPlane provides a mock function with given fields: ctx, id, opts
func (_m *MockControlPlaneSDK) DeleteControlPlane(ctx context.Context, id string, opts ...operations.Option) (*operations.DeleteControlPlaneResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, id)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteControlPlane")
	}

	var r0 *operations.DeleteControlPlaneResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...operations.Option) (*operations.DeleteControlPlaneResponse, error)); ok {
		return rf(ctx, id, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...operations.Option) *operations.DeleteControlPlaneResponse); ok {
		r0 = rf(ctx, id, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.DeleteControlPlaneResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...operations.Option) error); ok {
		r1 = rf(ctx, id, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockControlPlaneSDK_DeleteControlPlane_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteControlPlane'
type MockControlPlaneSDK_DeleteControlPlane_Call struct {
	*mock.Call
}

// DeleteControlPlane is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - opts ...operations.Option
func (_e *MockControlPlaneSDK_Expecter) DeleteControlPlane(ctx interface{}, id interface{}, opts ...interface{}) *MockControlPlaneSDK_DeleteControlPlane_Call {
	return &MockControlPlaneSDK_DeleteControlPlane_Call{Call: _e.mock.On("DeleteControlPlane",
		append([]interface{}{ctx, id}, opts...)...)}
}

func (_c *MockControlPlaneSDK_DeleteControlPlane_Call) Run(run func(ctx context.Context, id string, opts ...operations.Option)) *MockControlPlaneSDK_DeleteControlPlane_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *MockControlPlaneSDK_DeleteControlPlane_Call) Return(_a0 *operations.DeleteControlPlaneResponse, _a1 error) *MockControlPlaneSDK_DeleteControlPlane_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockControlPlaneSDK_DeleteControlPlane_Call) RunAndReturn(run func(context.Context, string, ...operations.Option) (*operations.DeleteControlPlaneResponse, error)) *MockControlPlaneSDK_DeleteControlPlane_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateControlPlane provides a mock function with given fields: ctx, id, req, opts
func (_m *MockControlPlaneSDK) UpdateControlPlane(ctx context.Context, id string, req components.UpdateControlPlaneRequest, opts ...operations.Option) (*operations.UpdateControlPlaneResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, id, req)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpdateControlPlane")
	}

	var r0 *operations.UpdateControlPlaneResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, components.UpdateControlPlaneRequest, ...operations.Option) (*operations.UpdateControlPlaneResponse, error)); ok {
		return rf(ctx, id, req, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, components.UpdateControlPlaneRequest, ...operations.Option) *operations.UpdateControlPlaneResponse); ok {
		r0 = rf(ctx, id, req, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.UpdateControlPlaneResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, components.UpdateControlPlaneRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, id, req, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockControlPlaneSDK_UpdateControlPlane_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateControlPlane'
type MockControlPlaneSDK_UpdateControlPlane_Call struct {
	*mock.Call
}

// UpdateControlPlane is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - req components.UpdateControlPlaneRequest
//   - opts ...operations.Option
func (_e *MockControlPlaneSDK_Expecter) UpdateControlPlane(ctx interface{}, id interface{}, req interface{}, opts ...interface{}) *MockControlPlaneSDK_UpdateControlPlane_Call {
	return &MockControlPlaneSDK_UpdateControlPlane_Call{Call: _e.mock.On("UpdateControlPlane",
		append([]interface{}{ctx, id, req}, opts...)...)}
}

func (_c *MockControlPlaneSDK_UpdateControlPlane_Call) Run(run func(ctx context.Context, id string, req components.UpdateControlPlaneRequest, opts ...operations.Option)) *MockControlPlaneSDK_UpdateControlPlane_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(components.UpdateControlPlaneRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockControlPlaneSDK_UpdateControlPlane_Call) Return(_a0 *operations.UpdateControlPlaneResponse, _a1 error) *MockControlPlaneSDK_UpdateControlPlane_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockControlPlaneSDK_UpdateControlPlane_Call) RunAndReturn(run func(context.Context, string, components.UpdateControlPlaneRequest, ...operations.Option) (*operations.UpdateControlPlaneResponse, error)) *MockControlPlaneSDK_UpdateControlPlane_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockControlPlaneSDK creates a new instance of MockControlPlaneSDK. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockControlPlaneSDK(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockControlPlaneSDK {
	mock := &MockControlPlaneSDK{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
