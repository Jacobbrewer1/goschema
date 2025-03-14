// Code generated by mockery. DO NOT EDIT.

package patcher

import (
	reflect "reflect"

	mock "github.com/stretchr/testify/mock"
)

// MockIgnoreFieldsFunc is an autogenerated mock type for the IgnoreFieldsFunc type
type MockIgnoreFieldsFunc struct {
	mock.Mock
}

// Execute provides a mock function with given fields: field
func (_m *MockIgnoreFieldsFunc) Execute(field *reflect.StructField) bool {
	ret := _m.Called(field)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(*reflect.StructField) bool); ok {
		r0 = rf(field)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// NewMockIgnoreFieldsFunc creates a new instance of MockIgnoreFieldsFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockIgnoreFieldsFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockIgnoreFieldsFunc {
	mock := &MockIgnoreFieldsFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
