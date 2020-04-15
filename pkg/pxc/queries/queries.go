package queries

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// value of writer group is hardcoded in ProxySQL config inside docker image
// https://github.com/percona/percona-docker/blob/pxc-operator-1.3.0/proxysql/dockerdir/etc/proxysql-admin.cnf#L23
const writerID = 11

type database struct {
	db *sql.DB
}

func New(client client.Client, namespace, secretName, user, host string, port int) (database, error) {
	secretObj := corev1.Secret{}
	err := client.Get(context.TODO(),
		types.NamespacedName{
			Namespace: namespace,
			Name:      secretName,
		},
		&secretObj,
	)
	if err != nil {
		return database{}, err
	}

	pass := string(secretObj.Data[user])
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql?interpolateParams=true", user, pass, host, port)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return database{}, err
	}

	err = db.Ping()
	if err != nil {
		return database{}, err
	}

	return database{
		db: db,
	}, nil
}

func (p *database) PrimaryHost() (string, error) {
	var host string
	err := p.db.QueryRow("SELECT hostname FROM runtime_mysql_servers WHERE hostgroup_id = ?", writerID).Scan(&host)
	if err != nil {
		return "", err
	}

	return host, nil
}

func (p *database) Close() error {
	return p.db.Close()
}
