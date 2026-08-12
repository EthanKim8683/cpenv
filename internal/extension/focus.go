package extension

import (
	"github.com/EthanKim8683/cpenv/internal/gen/focus/v1/focusv1connect"
	bolt "go.etcd.io/bbolt"
)

type FocusService struct {
	DB *bolt.DB
}

var _ focusv1connect.FocusServiceHandler = (*FocusService)(nil)
