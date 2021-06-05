package emotes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SevenTV/ServerGo/src/cache"
	"github.com/SevenTV/ServerGo/src/mongo/datastructure"
	"github.com/SevenTV/ServerGo/src/server/api/v2/rest/restutil"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetChannelEmotesRoute(router fiber.Router) {
	router.Get("/channels/:channel", func(c *fiber.Ctx) error {
		channelIdentifier := c.Params("channel")

		// Find channel user
		var channel *datastructure.User
		if err := cache.FindOne(c.Context(), "users", "", bson.M{
			"$or": bson.A{
				bson.M{"id": channelIdentifier},
				bson.M{"login": strings.ToLower(channelIdentifier)},
			},
		}, &channel); err != nil {
			fmt.Println("er", err)
			return restutil.ErrUnknownUser.Send(c)
		}

		// Find emotes
		var emotes []*datastructure.Emote
		if err := cache.Find(c.Context(), "emotes", "", bson.M{
			"_id": bson.M{
				"$in": channel.EmoteIDs,
			},
		}, &emotes); err != nil {
			return restutil.ErrInternalServer.Send(c, err.Error())
		}

		// Find IDs of emote owners
		ownerUserIDMap := make(map[*primitive.ObjectID]int)
		ownerIDs := []*primitive.ObjectID{}
		for _, emote := range emotes {
			if &emote.OwnerID == nil || ownerUserIDMap[&emote.OwnerID] == 1 {
				continue
			}

			ownerUserIDMap[&emote.OwnerID] = 1
			ownerIDs = append(ownerIDs, &emote.OwnerID)
		}

		// Map IDs to struct
		var owners []*datastructure.User
		ownerMap := make(map[primitive.ObjectID]*datastructure.User, len(owners))
		if err := cache.Find(c.Context(), "users", "", bson.M{
			"_id": bson.M{
				"$in": ownerIDs,
			},
		}, &owners); err != nil {
			return restutil.ErrInternalServer.Send(c, err.Error())
		}
		for _, o := range owners {
			ownerMap[o.ID] = o
		}

		// Create final response
		response := make([]restutil.EmoteResponse, len(emotes))
		for i, emote := range emotes {
			owner := ownerMap[emote.OwnerID]
			response[i] = restutil.CreateEmoteResponse(*emote, owner)
		}

		j, err := json.Marshal(response)
		if err != nil {
			return restutil.ErrInternalServer.Send(c, err.Error())
		}

		return c.Send(j)
	})
}