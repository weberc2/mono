package testsupport

import "github.com/weberc2/auth/pkg/types"

type NotificationServiceFake struct{}

func (NotificationServiceFake) Notify(*types.Notification) error { return nil }
