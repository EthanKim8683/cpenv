package extension

import (
	"github.com/EthanKim8683/cpenv/internal/gen/status/v1/statusv1connect"
	bolt "go.etcd.io/bbolt"
)

type StatusService struct {
	DB *bolt.DB
}

var _ statusv1connect.StatusServiceHandler = (*StatusService)(nil)
