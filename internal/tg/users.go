package tg

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/lithammer/fuzzysearch/fuzzy"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/pkg/errors"
)

const DefaultUsersLimit = 50

// nolint:lll
type UsersArguments struct {
	Search string `json:"search,omitempty" jsonschema:"description=Filter users by name similarity (case-insensitive)"`
	Offset string `json:"offset,omitempty" jsonschema:"description=Offset for continuation"`
	Limit  int    `json:"limit,omitempty" jsonschema:"description=Max number of users to return (default 20)"`
}

type UserInfo struct {
	Title    string `json:"title"`
	Username string `json:"username,omitempty"`
}

type UsersResponse struct {
	Users []UserInfo `json:"users"`
}

// GetUsers returns a lean list of users (people) you can message, with optional similarity search.
func (c *Client) GetUsers(args UsersArguments) (*mcp.ToolResponse, error) {
	var offset DialogsOffset
	if args.Offset != "" {
		if err := offset.UnmarshalJSON([]byte(args.Offset)); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal offset")
		}
	}
	if offset.Peer == nil {
		offset.Peer = &tg.InputPeerEmpty{}
	}

	limit := args.Limit
	if limit <= 0 {
		limit = DefaultUsersLimit
	}

	var infos []DialogInfo
	client := c.T()
	if err := client.Run(context.Background(), func(ctx context.Context) (err error) {
		api := client.API()
		cur := offset
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

			d, err := newDialogs(dc, DialogsArguments{Type: DialogTypeUser})
			if err != nil {
				return errors.Wrap(err, "failed to process dialogs")
			}

			batch := d.Info()
			infos = append(infos, batch...)

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
	users := make([]userScore, 0, len(infos))
	query := strings.ToLower(strings.TrimSpace(args.Search))

	for _, info := range infos {
		rank := min(matchRank(info.Title, query), matchRank(info.Name, query))
		if query != "" && rank < 0 {
			continue
		}
		users = append(users, userScore{
			UserInfo: UserInfo{Title: info.Title, Username: info.Name},
			score:    rank,
		})
	}

	if query != "" {
		sort.Slice(users, func(i, j int) bool {
			return users[i].score < users[j].score
		})
	}

	if limit > len(users) {
		limit = len(users)
	}

	result := make([]UserInfo, limit)
	for i := range result {
		result[i] = users[i].UserInfo
	}

	rsp := UsersResponse{
		Users: result,
	}

	jsonData, err := json.Marshal(rsp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response")
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(jsonData))), nil
}

type userScore struct {
	UserInfo
	score int
}

const maxFuzzyDistance = 3

// matchRank returns a rank where lower is better, and -1 means no match.
// When query is empty returns 0 so all entries pass through.
// Tiers: 0=exact, 1=prefix, 2=contains, 3..=levenshtein distance (capped at maxFuzzyDistance).
func matchRank(source, query string) int {
	if query == "" {
		return 0
	}
	if source == "" {
		return -1
	}
	lower := strings.ToLower(source)
	switch {
	case lower == query:
		return 0
	case strings.HasPrefix(lower, query):
		return 1
	case strings.Contains(lower, query):
		return 2
	default:
		dist := fuzzy.LevenshteinDistance(query, lower)
		if dist > maxFuzzyDistance {
			return -1
		}
		return 2 + dist
	}
}
