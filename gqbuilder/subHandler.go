package gqbuilder

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

type SubscriptionHandler struct {
	schema graphql.Schema
}

func GetSubscriptionHandler(schema graphql.Schema) *SubscriptionHandler {
	return &SubscriptionHandler{
		schema: schema,
	}
}

func (sh *SubscriptionHandler) SubscriptionsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed to do websocket upgrade: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	connectionACK, err := json.Marshal(map[string]string{
		"type": "connection_ack",
	})
	if err != nil {
		log.Error("failed to marshal ws connection ack: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, connectionACK); err != nil {
		log.Error("failed to write to ws connection: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go sh.handleSubscription(conn)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{"graphql-ws"},
}

type ConnectionACKMessage struct {
	OperationID string `json:"id,omitempty"`
	Type        string `json:"type"`
	Payload     struct {
		Query string `json:"query"`
	} `json:"payload,omitempty"`
}

func (sh *SubscriptionHandler) handleSubscription(conn *websocket.Conn) {
	var subscriber *Subscriber
	subscriptionCtx, subscriptionCancelFn := context.WithCancel(context.Background())

	handleClosedConnection := func() {
		log.Debug("[SubscriptionsHandler] subscriber closed connection")
		sh.unsubscribe(subscriptionCancelFn, subscriber)
		return
	}

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Error("failed to read websocket message: %v", err)
			return
		}

		var msg ConnectionACKMessage
		if err := json.Unmarshal(p, &msg); err != nil {
			log.Error("failed to unmarshal websocket message: %v", err)
			continue
		}

		if msg.Type == "stop" {
			handleClosedConnection()
			return
		}

		if msg.Type == "start" {
			subscriber = sh.subscribe(subscriptionCtx, subscriptionCancelFn, conn, msg)
		}
	}
}

type Subscriber struct {
	UUID          string
	Conn          *websocket.Conn
	RequestString string
	OperationID   string
}

var subscribers sync.Map

func (sh *SubscriptionHandler) subscribersSize() uint64 {
	var size uint64
	subscribers.Range(func(_, _ interface{}) bool {
		size++
		return true
	})
	return size
}

func (sh *SubscriptionHandler) unsubscribe(subscriptionCancelFn context.CancelFunc, subscriber *Subscriber) {
	subscriptionCancelFn()
	if subscriber != nil {
		subscriber.Conn.Close()
		subscribers.Delete(subscriber.UUID)
	}
	log.Debug("[SubscriptionsHandler] subscribers size: %+v", sh.subscribersSize())
}

func (sh *SubscriptionHandler) subscribe(ctx context.Context, subscriptionCancelFn context.CancelFunc, conn *websocket.Conn, msg ConnectionACKMessage) *Subscriber {
	subscriber := &Subscriber{
		UUID:          uuid.New().String(),
		Conn:          conn,
		RequestString: msg.Payload.Query,
		OperationID:   msg.OperationID,
	}
	subscribers.Store(subscriber.UUID, &subscriber)

	log.Debug("[SubscriptionsHandler] subscribers size: %+v", sh.subscribersSize())

	sendMessage := func(r *graphql.Result) error {
		message, err := json.Marshal(map[string]interface{}{
			"type":    "data",
			"id":      subscriber.OperationID,
			"payload": r.Data,
		})
		if err != nil {
			return err
		}

		if err := subscriber.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return err
		}

		return nil
	}

	go func() {
		subscribeParams := graphql.Params{
			Context:       ctx,
			RequestString: msg.Payload.Query,
			Schema:        sh.schema,
		}

		subscribeChannel := graphql.Subscribe(subscribeParams)

		for {
			select {
			case <-ctx.Done():
				log.Printf("[SubscriptionsHandler] subscription ctx done")
				return
			case r, isOpen := <-subscribeChannel:
				if !isOpen {
					log.Printf("[SubscriptionsHandler] subscription channel closed")
					sh.unsubscribe(subscriptionCancelFn, subscriber)
					return
				}
				if err := sendMessage(r); err != nil {
					if err == websocket.ErrCloseSent {
						sh.unsubscribe(subscriptionCancelFn, subscriber)
					}
					log.Printf("failed to send message: %v", err)
				}
			}
		}
	}()

	return subscriber
}
