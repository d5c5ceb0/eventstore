package sqlite3

import (
	"sync"

	"github.com/jmoiron/sqlx"
    "go.etcd.io/etcd/client/v3"
)

type SQLite3Backend struct {
	sync.Mutex
	*sqlx.DB
    *clientv3.Client
	DatabaseURL       string
	QueryLimit        int
	QueryIDsLimit     int
	QueryAuthorsLimit int
	QueryKindsLimit   int
	QueryTagsLimit    int
}

func (b *SQLite3Backend) Close() {
	b.DB.Close()
    b.Client.Close()
}
