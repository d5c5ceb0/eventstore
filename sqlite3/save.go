package sqlite3

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

func (b *SQLite3Backend) SaveEvent(ctx context.Context, evt *nostr.Event) error {
	// insert
	tagsj, _ := json.Marshal(evt.Tags)
	res, err := b.DB.ExecContext(ctx, `
        INSERT OR IGNORE INTO event (id, pubkey, created_at, kind, tags, content, sig)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, evt.ID, evt.PubKey, evt.CreatedAt, evt.Kind, tagsj, evt.Content, evt.Sig)
	if err != nil {
		return err
	}

	nr, err := res.RowsAffected()
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
    fmt.Println("id set successfully!, key = ", id_key)

    _, err = b.Client.Put(ctx, pubkey_key, pubkey_value)
    if err != nil {
        return err
    }
    fmt.Println("pubkey set successfully!, key = ", pubkey_key)


	if nr == 0 {
		return eventstore.ErrDupEvent
	}

	return nil
}
