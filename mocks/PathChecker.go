package mocks

import "github.com/stretchr/testify/mock"

// PathChecker is an autogenerated mock type for the PathChecker type
type PathChecker struct {
	mock.Mock
}

// IsPathExists provides a mock function with given fields: pth
func (_m *PathChecker) IsPathExists(pth string) (bool, error) {
	args := _m.Called(pth)
	var err error
	if len(args) > 1 {
		err = args.Error(1)
	}
	return args.Bool(0), err
}

// IsDirExists provides a mock function with given fields: pth
func (_m *PathChecker) IsDirExists(pth string) (bool, error) {
	args := _m.Called(pth)
	var err error
	if len(args) > 1 {
		err = args.Error(1)
	}
	return args.Bool(0), err
}
