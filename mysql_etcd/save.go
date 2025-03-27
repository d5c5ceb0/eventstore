package mysql_etcd

import (
    "fmt"
    "time"
	"context"
	"encoding/json"

	"github.com/fiatjaf/eventstore"
	"github.com/nbd-wtf/go-nostr"
)

const (
    //<MODELDAO_ID_PREFIX> + <event_id> -> <event_json>
    MODELDAO_ID_PREFIX = "/modeldao/id/"
    //<MODELDAO_PUBKEY_PREFIX> + <pubkey> + <event_id> -> <MODELDAO_REF_PREFIX> + <MODELDAO_ID_PREFIX> + <event_id>
    MODELDAO_PUBKEY_PREFIX = "/modeldao/pubkey/"
    //<MODELDAO_TAG_PREFIX> + <tag> + <event_id> -> <MODELDAO_REF_PREFIX> + <MODELDAO_ID_PREFIX> + <event_id>
    MODELDAO_TAG_PREFIX = "/modeldao/tag/"
    //reference to ID key
    MODELDAO_REF_PREFIX = "ref:"
)

func (b *MySQLBackend) SaveEvent(ctx context.Context, evt *nostr.Event) error {
	deleteQuery, deleteParams, shouldDelete := deleteBeforeSaveSql(evt)
	if shouldDelete {
		_, _ = b.DB.ExecContext(ctx, deleteQuery, deleteParams...)
	}

	sql, params, _ := saveEventSql(evt)
	res, err := b.DB.ExecContext(ctx, sql, params...)
	if err != nil {
		return err
	}

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    id_key := MODELDAO_ID_PREFIX + evt.ID
    pubkey_key := MODELDAO_PUBKEY_PREFIX + evt.PubKey + evt.ID
    pubkey_value := MODELDAO_REF_PREFIX + id_key

	evtsj, _ := json.Marshal(evt)
    _, err = b.Client.Put(ctx, id_key, string(evtsj))
    if err != nil {
        return err
    }
    fmt.Println("id_key: ", id_key)

    _, err = b.Client.Put(ctx, pubkey_key, pubkey_value)
    if err != nil {
        return err
    }
    fmt.Println("pubkey_key: ", pubkey_key)

	nr, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if nr == 0 {
		return eventstore.ErrDupEvent
	}

	return nil
}

func deleteBeforeSaveSql(evt *nostr.Event) (string, []any, bool) {
	// react to different kinds of events
	var (
		query        = ""
		params       []any
		shouldDelete bool
	)
	if evt.Kind == nostr.KindProfileMetadata || evt.Kind == nostr.KindFollowList || (10000 <= evt.Kind && evt.Kind < 20000) {
		// delete past events from this user
		query = `DELETE FROM event WHERE pubkey = ? AND kind = ?`
		params = []any{evt.PubKey, evt.Kind}
		shouldDelete = true
	} else if evt.Kind == nostr.KindRecommendServer {
		// delete past recommend_server events equal to this one
		query = `DELETE FROM event WHERE pubkey = ? AND kind = ? AND content = ?`
		params = []any{evt.PubKey, evt.Kind, evt.Content}
		shouldDelete = true
	} else if evt.Kind >= 30000 && evt.Kind < 40000 {
		// NIP-33
		d := evt.Tags.GetFirst([]string{"d"})
		if d != nil {
			query = `DELETE FROM event WHERE pubkey = ? AND kind = ? AND tags LIKE ?`
			params = []any{evt.PubKey, evt.Kind, d.Value()}
			shouldDelete = true
		}
	}

	return query, params, shouldDelete
}

func saveEventSql(evt *nostr.Event) (string, []any, error) {
	const query = `INSERT INTO event (
	id, pubkey, created_at, kind, tags, content, sig)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	var (
		tagsj, _ = json.Marshal(evt.Tags)
		params   = []any{evt.ID, evt.PubKey, evt.CreatedAt, evt.Kind, tagsj, evt.Content, evt.Sig}
	)

	return query, params, nil
}
