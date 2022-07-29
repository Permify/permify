package database

import (
	"github.com/Permify/permify/internal/config"
	MNDatabase "github.com/Permify/permify/pkg/database/mongo"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

// DBFactory -
func DBFactory(conf config.Write) (db Database, err error) {
	switch conf.Connection {
	case POSTGRES.String():
		db, err = PQDatabase.New(conf.URI, conf.Database, PQDatabase.MaxPoolSize(conf.PoolMax))
		if err != nil {
			return nil, err
		}
		return
	case MONGO.String():
		db, err = MNDatabase.New(conf.URI, conf.Database, MNDatabase.MaxPoolSize(conf.PoolMax))
		if err != nil {
			return nil, err
		}
		return
	default:
		db, err = PQDatabase.New(conf.URI, conf.Database, PQDatabase.MaxPoolSize(conf.PoolMax))
		if err != nil {
			return nil, err
		}
		return
	}
}
