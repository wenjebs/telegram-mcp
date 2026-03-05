package tg

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/gotd/td/tg"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/pkg/errors"
)

type DraftArguments struct {
	Name string `json:"name" jsonschema:"required,description=Username identifier of the dialog — use the 'username' field returned by tg_users or tg_groups"`
	Text string `json:"text" jsonschema:"required,description=Plain text of the message"`
}

type DraftResponse struct {
	Success bool `json:"success"`
}

func (c *Client) SendDraft(args DraftArguments) (*mcp.ToolResponse, error) {
	var ok bool
	client := c.T()
	if err := client.Run(context.Background(), func(ctx context.Context) (err error) {
		api := client.API()

		inputPeer, err := getInputPeerFromName(ctx, api, args.Name)
		if err != nil {
			return fmt.Errorf("get inputPeer from name: %w", err)
		}

		_, err = api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
			Peer:     inputPeer,
			Message:  args.Text,
			RandomID: rand.Int63(),
		})
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		ok = true

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to send message")
	}

	jsonData, err := json.Marshal(DraftResponse{Success: ok})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response")
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(jsonData))), nil
}
