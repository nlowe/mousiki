// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	audio "github.com/nlowe/mousiki/audio"
	mock "github.com/stretchr/testify/mock"
)

// Player is an autogenerated mock type for the Player type
type Player struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Player) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DoneChan provides a mock function with given fields:
func (_m *Player) DoneChan() <-chan error {
	ret := _m.Called()

	var r0 <-chan error
	if rf, ok := ret.Get(0).(func() <-chan error); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan error)
		}
	}

	return r0
}

// Pause provides a mock function with given fields:
func (_m *Player) Pause() {
	_m.Called()
}

// Play provides a mock function with given fields:
func (_m *Player) Play() {
	_m.Called()
}

// ProgressChan provides a mock function with given fields:
func (_m *Player) ProgressChan() <-chan audio.PlaybackProgress {
	ret := _m.Called()

	var r0 <-chan audio.PlaybackProgress
	if rf, ok := ret.Get(0).(func() <-chan audio.PlaybackProgress); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan audio.PlaybackProgress)
		}
	}

	return r0
}

// UpdateStream provides a mock function with given fields: url, volumeAdjustment
func (_m *Player) UpdateStream(url string, volumeAdjustment float64) {
	_m.Called(url, volumeAdjustment)
}
