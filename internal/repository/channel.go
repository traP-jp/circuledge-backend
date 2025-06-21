package repository

import (
	"context"
	"errors"
	"fmt"

	traqforest "github.com/comavius/traq-channel-forest-go"
	traq "github.com/traPtitech/go-traq"
)

type Channel struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

func (r *Repository) GetChannels() ([]Channel, error) {
	client := traq.NewAPIClient(traq.NewConfiguration())
	auth := context.WithValue(context.Background(), traq.ContextAccessToken, r.token)
	fmt.Println(r.token)
	c, _, err := client.ChannelApi.GetChannels(auth).Execute()
	if err != nil {
		return nil, err
	}
	forest, err := traqforest.NewForest(client, &auth)
	if err != nil {
		return nil, err
	}
	var channels []Channel
	for _, t := range c.Public {
		path, ok := forest.GetPath(t.Id)
		if !ok {
			return nil, errors.New("failed to get path for channel " + t.Id)
		}
		channels = append(channels, Channel{
			ID:   t.Id,
			Path: path,
		})
	}

	return channels, nil
}
