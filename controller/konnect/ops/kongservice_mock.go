// Code generated by mockery. DO NOT EDIT.

package ops

import (
	context "context"

	components "github.com/Kong/sdk-konnect-go/models/components"

	mock "github.com/stretchr/testify/mock"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
)

// MockServicesSDK is an autogenerated mock type for the ServicesSDK type
type MockServicesSDK struct {
	mock.Mock
}

type MockServicesSDK_Expecter struct {
	mock *mock.Mock
}

func (_m *MockServicesSDK) EXPECT() *MockServicesSDK_Expecter {
	return &MockServicesSDK_Expecter{mock: &_m.Mock}
}

// CreateService provides a mock function with given fields: ctx, controlPlaneID, service, opts
func (_m *MockServicesSDK) CreateService(ctx context.Context, controlPlaneID string, service components.ServiceInput, opts ...operations.Option) (*operations.CreateServiceResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, service)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateService")
	}

	var r0 *operations.CreateServiceResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, components.ServiceInput, ...operations.Option) (*operations.CreateServiceResponse, error)); ok {
		return rf(ctx, controlPlaneID, service, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, components.ServiceInput, ...operations.Option) *operations.CreateServiceResponse); ok {
		r0 = rf(ctx, controlPlaneID, service, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.CreateServiceResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, components.ServiceInput, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, service, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockServicesSDK_CreateService_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateService'
type MockServicesSDK_CreateService_Call struct {
	*mock.Call
}

// CreateService is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - service components.ServiceInput
//   - opts ...operations.Option
func (_e *MockServicesSDK_Expecter) CreateService(ctx interface{}, controlPlaneID interface{}, service interface{}, opts ...interface{}) *MockServicesSDK_CreateService_Call {
	return &MockServicesSDK_CreateService_Call{Call: _e.mock.On("CreateService",
		append([]interface{}{ctx, controlPlaneID, service}, opts...)...)}
}

func (_c *MockServicesSDK_CreateService_Call) Run(run func(ctx context.Context, controlPlaneID string, service components.ServiceInput, opts ...operations.Option)) *MockServicesSDK_CreateService_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(components.ServiceInput), variadicArgs...)
	})
	return _c
}

func (_c *MockServicesSDK_CreateService_Call) Return(_a0 *operations.CreateServiceResponse, _a1 error) *MockServicesSDK_CreateService_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServicesSDK_CreateService_Call) RunAndReturn(run func(context.Context, string, components.ServiceInput, ...operations.Option) (*operations.CreateServiceResponse, error)) *MockServicesSDK_CreateService_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteService provides a mock function with given fields: ctx, controlPlaneID, serviceID, opts
func (_m *MockServicesSDK) DeleteService(ctx context.Context, controlPlaneID string, serviceID string, opts ...operations.Option) (*operations.DeleteServiceResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, controlPlaneID, serviceID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteService")
	}

	var r0 *operations.DeleteServiceResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) (*operations.DeleteServiceResponse, error)); ok {
		return rf(ctx, controlPlaneID, serviceID, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...operations.Option) *operations.DeleteServiceResponse); ok {
		r0 = rf(ctx, controlPlaneID, serviceID, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.DeleteServiceResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, ...operations.Option) error); ok {
		r1 = rf(ctx, controlPlaneID, serviceID, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockServicesSDK_DeleteService_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteService'
type MockServicesSDK_DeleteService_Call struct {
	*mock.Call
}

// DeleteService is a helper method to define mock.On call
//   - ctx context.Context
//   - controlPlaneID string
//   - serviceID string
//   - opts ...operations.Option
func (_e *MockServicesSDK_Expecter) DeleteService(ctx interface{}, controlPlaneID interface{}, serviceID interface{}, opts ...interface{}) *MockServicesSDK_DeleteService_Call {
	return &MockServicesSDK_DeleteService_Call{Call: _e.mock.On("DeleteService",
		append([]interface{}{ctx, controlPlaneID, serviceID}, opts...)...)}
}

func (_c *MockServicesSDK_DeleteService_Call) Run(run func(ctx context.Context, controlPlaneID string, serviceID string, opts ...operations.Option)) *MockServicesSDK_DeleteService_Call {
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

func (_c *MockServicesSDK_DeleteService_Call) Return(_a0 *operations.DeleteServiceResponse, _a1 error) *MockServicesSDK_DeleteService_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServicesSDK_DeleteService_Call) RunAndReturn(run func(context.Context, string, string, ...operations.Option) (*operations.DeleteServiceResponse, error)) *MockServicesSDK_DeleteService_Call {
	_c.Call.Return(run)
	return _c
}

// ListService provides a mock function with given fields: ctx, request, opts
func (_m *MockServicesSDK) ListService(ctx context.Context, request operations.ListServiceRequest, opts ...operations.Option) (*operations.ListServiceResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListService")
	}

	var r0 *operations.ListServiceResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListServiceRequest, ...operations.Option) (*operations.ListServiceResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListServiceRequest, ...operations.Option) *operations.ListServiceResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.ListServiceResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.ListServiceRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockServicesSDK_ListService_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListService'
type MockServicesSDK_ListService_Call struct {
	*mock.Call
}

// ListService is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.ListServiceRequest
//   - opts ...operations.Option
func (_e *MockServicesSDK_Expecter) ListService(ctx interface{}, request interface{}, opts ...interface{}) *MockServicesSDK_ListService_Call {
	return &MockServicesSDK_ListService_Call{Call: _e.mock.On("ListService",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockServicesSDK_ListService_Call) Run(run func(ctx context.Context, request operations.ListServiceRequest, opts ...operations.Option)) *MockServicesSDK_ListService_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.ListServiceRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockServicesSDK_ListService_Call) Return(_a0 *operations.ListServiceResponse, _a1 error) *MockServicesSDK_ListService_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServicesSDK_ListService_Call) RunAndReturn(run func(context.Context, operations.ListServiceRequest, ...operations.Option) (*operations.ListServiceResponse, error)) *MockServicesSDK_ListService_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertService provides a mock function with given fields: ctx, req, opts
func (_m *MockServicesSDK) UpsertService(ctx context.Context, req operations.UpsertServiceRequest, opts ...operations.Option) (*operations.UpsertServiceResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, req)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpsertService")
	}

	var r0 *operations.UpsertServiceResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertServiceRequest, ...operations.Option) (*operations.UpsertServiceResponse, error)); ok {
		return rf(ctx, req, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.UpsertServiceRequest, ...operations.Option) *operations.UpsertServiceResponse); ok {
		r0 = rf(ctx, req, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.UpsertServiceResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.UpsertServiceRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, req, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockServicesSDK_UpsertService_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertService'
type MockServicesSDK_UpsertService_Call struct {
	*mock.Call
}

// UpsertService is a helper method to define mock.On call
//   - ctx context.Context
//   - req operations.UpsertServiceRequest
//   - opts ...operations.Option
func (_e *MockServicesSDK_Expecter) UpsertService(ctx interface{}, req interface{}, opts ...interface{}) *MockServicesSDK_UpsertService_Call {
	return &MockServicesSDK_UpsertService_Call{Call: _e.mock.On("UpsertService",
		append([]interface{}{ctx, req}, opts...)...)}
}

func (_c *MockServicesSDK_UpsertService_Call) Run(run func(ctx context.Context, req operations.UpsertServiceRequest, opts ...operations.Option)) *MockServicesSDK_UpsertService_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.UpsertServiceRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockServicesSDK_UpsertService_Call) Return(_a0 *operations.UpsertServiceResponse, _a1 error) *MockServicesSDK_UpsertService_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServicesSDK_UpsertService_Call) RunAndReturn(run func(context.Context, operations.UpsertServiceRequest, ...operations.Option) (*operations.UpsertServiceResponse, error)) *MockServicesSDK_UpsertService_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockServicesSDK creates a new instance of MockServicesSDK. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockServicesSDK(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockServicesSDK {
	mock := &MockServicesSDK{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
