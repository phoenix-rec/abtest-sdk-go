package abtest

import "errors"

var (
	ErrClientStopped       = errors.New("client stopped")
	ErrClientUninitialized = errors.New("client uninitialized")
	ErrClientSettingErr    = errors.New("client project_id is empty")
	ErrExperimentDisabled  = errors.New("experiment disabled")
	ErrExperimentNotFound  = errors.New("experiment not found")
	ErrProjectNotFound     = errors.New("project not found")
	ErrExperimentNotMatch  = errors.New("experiment not match")
	ErrKeyNotFound         = errors.New("key not found")
	ErrSnapshotDisabled    = errors.New("snapshot disabled")
	ErrAllDefault          = errors.New("A/B Server is unavailable. All of the experiments are using the default value in code!")
)
