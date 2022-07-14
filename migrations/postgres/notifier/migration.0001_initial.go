package notifier

import (
	"fmt"
)

// CreateNotifyFunctionMigration -
func CreateNotifyFunctionMigration() string {
	return `create or replace function notify_authorization_event() returns trigger
   language plpgsql
as
$$
DECLARE
   notification json;
BEGIN

   notification = json_build_object(
           'entity', TG_TABLE_NAME,
           'action', TG_OP,
           'old_data', row_to_json(OLD),
           'new_data', row_to_json(NEW));
   PERFORM pg_notify('authorization_events', notification::text);

   RETURN NULL;
END;
$$;
`
}

// CreateNotifyTriggerMigration -
func CreateNotifyTriggerMigration(table string) string {
	return fmt.Sprintf(`
	drop trigger if exists %s_notify_authorization_event on "%s";

	create trigger %s_notify_authorization_event
  after insert or update or delete on "%s"
  for each row execute procedure notify_authorization_event();`, table, table, table, table)
}

//func CreatePublicationMigration(tables []string) string {
//	var sb strings.Builder
//	sb.WriteString("DROP PUBLICATION IF EXISTS authorization_pub;\n")
//	if len(tables) < 1 {
//		sb.WriteString("CREATE PUBLICATION authorization_pub FOR ALL TABLES;")
//		return sb.String()
//	}
//	sb.WriteString("CREATE PUBLICATION authorization_pub FOR TABLE ")
//	for i := 0; i < len(tables); i++ {
//		if i == len(tables)-1 {
//			sb.WriteString("\"" + tables[i] + "\";")
//		} else {
//			sb.WriteString("\"" + tables[i] + "\",")
//		}
//	}
//	return sb.String()
//}
