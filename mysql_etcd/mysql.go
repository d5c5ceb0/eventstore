package mysql_etcd

import (
	"sync"

	"github.com/jmoiron/sqlx"
    "go.etcd.io/etcd/client/v3"
)

type MySQLBackend struct {
	sync.Mutex
	*sqlx.DB
    *clientv3.Client
    EtcdURL           string
	DatabaseURL       string
	QueryLimit        int
	QueryIDsLimit     int
	QueryAuthorsLimit int
	QueryKindsLimit   int
	QueryTagsLimit    int
}

func (b *MySQLBackend) Close() {
	b.DB.Close()
    b.Client.Close()
}
