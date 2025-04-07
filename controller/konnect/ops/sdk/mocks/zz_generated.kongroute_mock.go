// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	components "github.com/Kong/sdk-konnect-go/models/components"

	mock "github.com/stretchr/testify/mock"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
)

// MockRoutesSDK is an autogenerated mock type for the RoutesSDK type
type MockRoutesSDK struct {
	mock.Mock
}

type MockRoutesSDK_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRoutesSDK) EXPECT() *MockRoutesSDK_Expecter {
	return &MockRoutesSDK_Expecter{mock: &_m.Mock}
}

// CreateRoute provides a mock function with given fields: ctx, controlPlaneID, route, opts
func (_m *MockRoutesSDK) CreateRoute(ctx context.Context, controlPlaneID string, route components.Route, opts ...operations.Option) (*operations.CreateRouteResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, route)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateRoute")
	}

	var r0 *operations.CreateRouteResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, components.Route, ...operations.Option) (*operations.CreateRouteResponse, error)); ok {
		return rf(ctx, controlPlaneID, route, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, components.Route, ...operations.Option) *operations.CreateRouteResponse); ok {
		r0 = rf(ctx, controlPlaneID, route, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.CreateRouteResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, components.Route, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, route, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockRoutesSDK_CreateRoute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateRoute'
type MockRoutesSDK_CreateRoute_Call struct {
	*mock.Call
}

// CreateRoute is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - route components.Route
//   - opts ...operations.Option
func (_e *MockRoutesSDK_Expecter) CreateRoute(ctx interface{}, controlPlaneID interface{}, route interface{}, opts ...interface{}) *MockRoutesSDK_CreateRoute_Call {
	return &MockRoutesSDK_CreateRoute_Call{Call: _e.mock.On("CreateRoute",
		append([]interface{}{ctx, controlPlaneID, route}, opts...)...)}
}

func (_c *MockRoutesSDK_CreateRoute_Call) Run(run func(ctx context.Context, controlPlaneID string, route components.Route, opts ...operations.Option)) *MockRoutesSDK_CreateRoute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(components.Route), variadicArgs...)
	})
	return _c
}

func (_c *MockRoutesSDK_CreateRoute_Call) Return(_a0 *operations.CreateRouteResponse, _a1 error) *MockRoutesSDK_CreateRoute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockRoutesSDK_CreateRoute_Call) RunAndReturn(run func(context.Context, string, components.Route, ...operations.Option) (*operations.CreateRouteResponse, error)) *MockRoutesSDK_CreateRoute_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteRoute provides a mock function with given fields: ctx, controlPlaneID, routeID, opts
func (_m *MockRoutesSDK) DeleteRoute(ctx context.Context, controlPlaneID string, routeID string, opts ...operations.Option) (*operations.DeleteRouteResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, routeID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteRoute")
	}

	var r0 *operations.DeleteRouteResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) (*operations.DeleteRouteResponse, error)); ok {
		return rf(ctx, controlPlaneID, routeID, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) *operations.DeleteRouteResponse); ok {
		r0 = rf(ctx, controlPlaneID, routeID, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.DeleteRouteResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, routeID, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockRoutesSDK_DeleteRoute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteRoute'
type MockRoutesSDK_DeleteRoute_Call struct {
	*mock.Call
}

// DeleteRoute is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - routeID string
//   - opts ...operations.Option
func (_e *MockRoutesSDK_Expecter) DeleteRoute(ctx interface{}, controlPlaneID interface{}, routeID interface{}, opts ...interface{}) *MockRoutesSDK_DeleteRoute_Call {
	return &MockRoutesSDK_DeleteRoute_Call{Call: _e.mock.On("DeleteRoute",
		append([]interface{}{ctx, controlPlaneID, routeID}, opts...)...)}
}

func (_c *MockRoutesSDK_DeleteRoute_Call) Run(run func(ctx context.Context, controlPlaneID string, routeID string, opts ...operations.Option)) *MockRoutesSDK_DeleteRoute_Call {
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

func (_c *MockRoutesSDK_DeleteRoute_Call) Return(_a0 *operations.DeleteRouteResponse, _a1 error) *MockRoutesSDK_DeleteRoute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockRoutesSDK_DeleteRoute_Call) RunAndReturn(run func(context.Context, string, string, ...operations.Option) (*operations.DeleteRouteResponse, error)) *MockRoutesSDK_DeleteRoute_Call {
	_c.Call.Return(run)
	return _c
}

// ListRoute provides a mock function with given fields: ctx, request, opts
func (_m *MockRoutesSDK) ListRoute(ctx context.Context, request operations.ListRouteRequest, opts ...operations.Option) (*operations.ListRouteResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListRoute")
	}

	var r0 *operations.ListRouteResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListRouteRequest, ...operations.Option) (*operations.ListRouteResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListRouteRequest, ...operations.Option) *operations.ListRouteResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.ListRouteResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.ListRouteRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockRoutesSDK_ListRoute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListRoute'
type MockRoutesSDK_ListRoute_Call struct {
	*mock.Call
}

// ListRoute is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.ListRouteRequest
//   - opts ...operations.Option
func (_e *MockRoutesSDK_Expecter) ListRoute(ctx interface{}, request interface{}, opts ...interface{}) *MockRoutesSDK_ListRoute_Call {
	return &MockRoutesSDK_ListRoute_Call{Call: _e.mock.On("ListRoute",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockRoutesSDK_ListRoute_Call) Run(run func(ctx context.Context, request operations.ListRouteRequest, opts ...operations.Option)) *MockRoutesSDK_ListRoute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.ListRouteRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockRoutesSDK_ListRoute_Call) Return(_a0 *operations.ListRouteResponse, _a1 error) *MockRoutesSDK_ListRoute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockRoutesSDK_ListRoute_Call) RunAndReturn(run func(context.Context, operations.ListRouteRequest, ...operations.Option) (*operations.ListRouteResponse, error)) *MockRoutesSDK_ListRoute_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertRoute provides a mock function with given fields: ctx, req, opts
func (_m *MockRoutesSDK) UpsertRoute(ctx context.Context, req operations.UpsertRouteRequest, opts ...operations.Option) (*operations.UpsertRouteResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, req)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpsertRoute")
	}

	var r0 *operations.UpsertRouteResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertRouteRequest, ...operations.Option) (*operations.UpsertRouteResponse, error)); ok {
		return rf(ctx, req, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertRouteRequest, ...operations.Option) *operations.UpsertRouteResponse); ok {
		r0 = rf(ctx, req, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.UpsertRouteResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.UpsertRouteRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, req, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockRoutesSDK_UpsertRoute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertRoute'
type MockRoutesSDK_UpsertRoute_Call struct {
	*mock.Call
}

// UpsertRoute is a helper method to define mock.On call
//   - ctx context.Context
//   - req operations.UpsertRouteRequest
//   - opts ...operations.Option
func (_e *MockRoutesSDK_Expecter) UpsertRoute(ctx interface{}, req interface{}, opts ...interface{}) *MockRoutesSDK_UpsertRoute_Call {
	return &MockRoutesSDK_UpsertRoute_Call{Call: _e.mock.On("UpsertRoute",
		append([]interface{}{ctx, req}, opts...)...)}
}

func (_c *MockRoutesSDK_UpsertRoute_Call) Run(run func(ctx context.Context, req operations.UpsertRouteRequest, opts ...operations.Option)) *MockRoutesSDK_UpsertRoute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.UpsertRouteRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockRoutesSDK_UpsertRoute_Call) Return(_a0 *operations.UpsertRouteResponse, _a1 error) *MockRoutesSDK_UpsertRoute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockRoutesSDK_UpsertRoute_Call) RunAndReturn(run func(context.Context, operations.UpsertRouteRequest, ...operations.Option) (*operations.UpsertRouteResponse, error)) *MockRoutesSDK_UpsertRoute_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockRoutesSDK creates a new instance of MockRoutesSDK. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRoutesSDK(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRoutesSDK {
	mock := &MockRoutesSDK{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
