package tg

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/gotd/td/tg"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/pkg/errors"
)

const DefaultGroupsLimit = 10

// nolint:lll
type GroupsArguments struct {
	Search string `json:"search,omitempty" jsonschema:"description=Filter groups by name similarity (case-insensitive)"`
	Limit  int    `json:"limit,omitempty" jsonschema:"description=Max number of groups to return (default 50)"`
}

type GroupInfo struct {
	Title    string `json:"title"`
	Username string `json:"username,omitempty"`
}

type GroupsResponse struct {
	Groups []GroupInfo `json:"groups"`
}

// GetGroups returns a list of group chats (not channels) with optional similarity search.
func (c *Client) GetGroups(args GroupsArguments) (*mcp.ToolResponse, error) {
	limit := args.Limit
	if limit <= 0 {
		limit = DefaultGroupsLimit
	}

	var infos []DialogInfo
	client := c.T()
	if err := client.Run(context.Background(), func(ctx context.Context) (err error) {
		api := client.API()
		cur := DialogsOffset{Peer: &tg.InputPeerEmpty{}}
		for len(infos) < limit {
			dc, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
				OffsetPeer: cur.Peer,
				OffsetID:   cur.MsgID,
				OffsetDate: cur.Date,
				Limit:      100,
			})
			if err != nil {
				return errors.Wrap(err, "failed to get dialogs")
			}

			d, err := newDialogs(dc, DialogsArguments{Type: DialogTypeChat})
			if err != nil {
				return errors.Wrap(err, "failed to process dialogs")
			}

			infos = append(infos, d.Info()...)

			next := d.Offset()
			if next.String() == "end" || next.String() == cur.String() {
				break
			}
			cur = next
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to get dialogs")
	}

	query := strings.ToLower(strings.TrimSpace(args.Search))

	type groupScore struct {
		GroupInfo
		score int
	}

	groups := make([]groupScore, 0, len(infos))
	for _, info := range infos {
		rank := matchRank(info.Title, query)
		if query != "" && rank < 0 {
			continue
		}
		groups = append(groups, groupScore{
			GroupInfo: GroupInfo{Title: info.Title, Username: info.Name},
			score:     rank,
		})
	}

	if query != "" {
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].score < groups[j].score
		})
	}

	if limit > len(groups) {
		limit = len(groups)
	}

	result := make([]GroupInfo, limit)
	for i := range result {
		result[i] = groups[i].GroupInfo
	}

	jsonData, err := json.Marshal(GroupsResponse{Groups: result})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response")
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(jsonData))), nil
}
