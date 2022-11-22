package postgres

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgproto3/v2"

	"github.com/Permify/permify/pkg/logger"
)

const (
	INSERT = "INSERT"
	UPDATE = "UPDATE"
	DELETE = "DELETE"
)

// Notification - Structure for Notification
type Notification struct {
	Entity  string                 `json:"entity"`
	Action  string                 `json:"action"`
	OldData map[string]interface{} `json:"old_data"`
	NewData map[string]interface{} `json:"new_data"`
}

// Event - Structure for Notification
var Event = struct {
	Entity  string
	Columns []string
}{}

// Publisher - Structure for Publisher
type Publisher struct {
	conn         *pgconn.PgConn
	slotName     string
	outputPlugin string
	tables       []string
	subscribers  []chan<- *Notification
	logger       logger.Interface
}

// NewPublisher - Creates new publisher
func NewPublisher(ctx context.Context, url string, sl string, op string, tables []string, logger logger.Interface) (publisher *Publisher, err error) {
	publisher = &Publisher{
		slotName:     sl,
		outputPlugin: op,
		tables:       tables,
		logger:       logger,
	}

	var conn *pgconn.PgConn
	conn, err = pgconn.Connect(ctx, url)
	if err != nil {
		return nil, err
	}

	publisher.conn = conn

	return
}

// Migrate -
func (p *Publisher) Migrate(ctx context.Context) (err error) {
	if _, err = p.conn.Exec(ctx, "DROP PUBLICATION IF EXISTS pub;").ReadAll(); err != nil {
		p.logger.Error(fmt.Errorf("permify - Run - failed to drop publication: %w", err))
	}

	if _, err = p.conn.Exec(ctx, fmt.Sprintf("CREATE PUBLICATION pub FOR TABLE %s;", strings.Join(p.tables, ","))).ReadAll(); err != nil {
		p.logger.Error(fmt.Errorf("permify - Run - failed to create publication: %w", err))
	}

	for _, table := range p.tables {
		if _, err = p.conn.Exec(ctx, fmt.Sprintf("ALTER TABLE %s REPLICA IDENTITY FULL;", table)).ReadAll(); err != nil {
			p.logger.Error(fmt.Errorf("permify - Run - failed to create publication: %w", err))
		}
	}

	return err
}

// CreateReplicationSlotServer -
func (p *Publisher) CreateReplicationSlotServer(ctx context.Context) (err error) {
	if _, err = pglogrepl.CreateReplicationSlot(ctx, p.conn, p.slotName, p.outputPlugin, pglogrepl.CreateReplicationSlotOptions{Temporary: true}); err != nil {
		p.logger.Error(fmt.Errorf("permify - Run - failed to create replication slot: %w", err))
	}
	return err
}

// Start -
func (p *Publisher) Start() {
	var err error

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err = p.Migrate(ctx)
	err = p.CreateReplicationSlotServer(ctx)

	var msgPointer pglogrepl.LSN
	pluginArguments := []string{"proto_version '1'", "publication_names 'pub'"}

	err = pglogrepl.StartReplication(ctx, p.conn, p.slotName, msgPointer, pglogrepl.StartReplicationOptions{PluginArgs: pluginArguments})
	if err != nil {
		p.logger.Error(fmt.Errorf("permify - Run - failed to establish start replication: %w", err))
	}

	var ping time.Time

	for ctx.Err() != context.Canceled {
		if time.Now().After(ping) {
			if err = pglogrepl.SendStandbyStatusUpdate(ctx, p.conn, pglogrepl.StandbyStatusUpdate{WALWritePosition: msgPointer}); err != nil {
				p.logger.Error("failed to send standby update: %v", err)
			}
			ping = time.Now().Add(10 * time.Second)
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		msg, err := p.conn.ReceiveMessage(ctx)
		if pgconn.Timeout(err) {
			continue
		}
		if err != nil {
			p.logger.Error(fmt.Errorf("something went wrong while listening for message: %v", err))
			continue
		}

		switch msg := msg.(type) {
		case *pgproto3.CopyData:
			switch msg.Data[0] {
			case pglogrepl.PrimaryKeepaliveMessageByteID:
				break
			case pglogrepl.XLogDataByteID:
				walLog, err := pglogrepl.ParseXLogData(msg.Data[1:])
				if err != nil {
					p.logger.Error(fmt.Errorf("publisher - Start - failed to parse logical WAL log: %w", err))
				}

				var msg pglogrepl.Message
				if msg, err = pglogrepl.Parse(walLog.WALData); err != nil {
					p.logger.Error(fmt.Errorf("publisher - Start - failed to parse logical replication message: %w", err))
				}

				switch m := msg.(type) {
				case *pglogrepl.RelationMessage:
					Event.Columns = []string{}
					for _, col := range m.Columns {
						Event.Columns = append(Event.Columns, col.Name)
					}
					Event.Entity = m.RelationName
				case *pglogrepl.InsertMessage:
					notification := &Notification{}
					notification.Action = INSERT
					notification.Entity = Event.Entity
					notification.NewData = map[string]interface{}{}

					for i := 0; i < len(Event.Columns); i++ {
						notification.NewData[Event.Columns[i]] = string(m.Tuple.Columns[i].Data)
					}

					for _, sub := range p.subscribers {
						sub <- notification
					}
				case *pglogrepl.UpdateMessage:
					notification := &Notification{}
					notification.Action = UPDATE
					notification.Entity = Event.Entity
					notification.NewData = map[string]interface{}{}
					notification.OldData = map[string]interface{}{}

					for i := 0; i < len(Event.Columns); i++ {
						notification.NewData[Event.Columns[i]] = string(m.NewTuple.Columns[i].Data)
						notification.OldData[Event.Columns[i]] = string(m.OldTuple.Columns[i].Data)
					}

					for _, sub := range p.subscribers {
						sub <- notification
					}
				case *pglogrepl.DeleteMessage:
					notification := &Notification{}
					notification.Action = DELETE
					notification.Entity = Event.Entity
					notification.OldData = map[string]interface{}{}

					for i := 0; i < len(Event.Columns); i++ {
						notification.OldData[Event.Columns[i]] = string(m.OldTuple.Columns[i].Data)
					}

					for _, sub := range p.subscribers {
						sub <- notification
					}
				case *pglogrepl.TruncateMessage:
					p.logger.Error("publisher - Start - ALL GONE (TRUNCATE)")
				}
			}
		default:
			p.logger.Error(fmt.Errorf("received unexpected message: %T", msg))
			continue
		}
	}
}

// Subscribe -
func (p *Publisher) Subscribe(c chan<- *Notification) {
	p.subscribers = append(p.subscribers, c)
}

// Unsubscribe -
func (p *Publisher) Unsubscribe(c chan *Notification) {
	terminated := make(chan struct{})
	go func() {
		for {
			select {
			case <-c:
			case <-terminated:
				return
			}
		}
	}()
	newSubscribers := make([]chan<- *Notification, 0)
	for _, existing := range p.subscribers {
		if existing != c {
			newSubscribers = append(newSubscribers, existing)
		}
	}
	p.subscribers = newSubscribers
	close(terminated)
}
