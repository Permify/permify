package factories

import (
	"errors"
	"fmt"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/repositories/memory"
	"github.com/Permify/permify/pkg/database"
	IMDatabase "github.com/Permify/permify/pkg/database/memory"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

// DatabaseFactory -
func DatabaseFactory(conf config.Write) (db database.Database, err error) {
	switch conf.Connection {
	case database.POSTGRES.String():
		db, err = PQDatabase.New(conf.URI, conf.Database, PQDatabase.MaxPoolSize(conf.PoolMax))
		if err != nil {
			return nil, err
		}
		return
	case database.MEMORY.String():
		db, err = IMDatabase.New(memory.Schema)
		if err != nil {
			return nil, err
		}
		return
	default:
		return nil, errors.New(fmt.Sprintf("%s connection is unsupported", conf.Connection))
	}
}
