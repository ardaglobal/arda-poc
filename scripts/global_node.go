package scripts

import (
	"context"
	"fmt"

	"github.com/cometbft/cometbft/libs/pubsub/query"
	"github.com/cometbft/cometbft/rpc/client/http"
	cmttypes "github.com/cometbft/cometbft/types"
)

func GlobalNode() {
	// Connect to the local CometBFT node
	c, err := http.New("tcp://localhost:26657", "/websocket")
	if err != nil {
		panic(err)
	}
	err = c.Start()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = c.Stop()
	}()

	// Build the query: listen for Tx events where submission.valid = 'true'
	query, err := query.New("tm.event='Tx' AND submission.valid='true'")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	// Subscribe to events
	eventCh, err := c.Subscribe(ctx, "global-registry", query.String(), 100)
	if err != nil {
		panic(err)
	}
	fmt.Println("Subscribed to valid submission events...")

	for {
		select {
		case event := <-eventCh:
			txEvent, ok := event.Data.(cmttypes.EventDataTx)
			if !ok {
				fmt.Println("Received event is not a Tx event")
				continue
			}
			// Each txResult contains all events; we filter our module's event
			for _, ev := range txEvent.Result.Events {
				if ev.Type == "submission" {
					// Extract attributes
					var region, hash string
					var validAttr string
					for _, attr := range ev.Attributes {
						if string(attr.Key) == "region" {
							region = string(attr.Value)
						} else if string(attr.Key) == "hash" {
							hash = string(attr.Value)
						} else if string(attr.Key) == "valid" {
							validAttr = string(attr.Value)
						}
					}
					fmt.Printf("Indexed verified hash: %s (region: %s, valid: %s)\n", hash, region, validAttr)
					// Here, we could store this information in a database or memory structure.
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
