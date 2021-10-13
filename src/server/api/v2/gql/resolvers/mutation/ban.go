package mutation_resolvers

import (
	"context"
	"time"

	"github.com/SevenTV/ServerGo/src/mongo"
	"github.com/SevenTV/ServerGo/src/mongo/datastructure"
	"github.com/SevenTV/ServerGo/src/server/api/actions"
	"github.com/SevenTV/ServerGo/src/server/api/v2/gql/resolvers"
	"github.com/SevenTV/ServerGo/src/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//
// BAN USER
//
func (*MutationResolver) BanUser(ctx context.Context, args struct {
	VictimID string
	ExpireAt *string
	Reason   *string
}) (*response, error) {
	usr, ok := ctx.Value(utils.UserKey).(*datastructure.User)
	if !ok {
		return nil, resolvers.ErrLoginRequired
	}

	// Verify actor has permission to ban
	if !usr.HasPermission(datastructure.RolePermissionBanUsers) {
		return nil, resolvers.ErrAccessDenied
	}

	// Serialize id to ObjectID
	id, err := primitive.ObjectIDFromHex(args.VictimID)
	if err != nil {
		return nil, resolvers.ErrUnknownUser
	}

	// Is actor silly?
	if id.Hex() == usr.ID.Hex() {
		return nil, resolvers.ErrYourself
	}

	// Check if ban already exists on victim
	banned, _ := actions.Bans.IsUserBanned(id)
	if banned {
		return nil, resolvers.ErrUserBanned
	}

	// Find user
	res := mongo.Collection(mongo.CollectionNameUsers).FindOne(ctx, bson.M{
		"_id": id,
	})
	user := &datastructure.User{}
	err = res.Err()
	if err == nil {
		err = res.Decode(user)
		role := datastructure.GetRole(user.RoleID)
		user.Role = &role
	}

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, resolvers.ErrUnknownUser
		}
		logrus.Errorf("mongo, err=%v", err)
		return nil, resolvers.ErrInternalServer
	}

	// Check if actor has a higher role than victim
	if user.Role.Position >= usr.Role.Position {
		return nil, resolvers.ErrAccessDenied
	}

	reasonN := "no reason"
	if args.Reason != nil {
		reasonN = *args.Reason
	}

	expireAt := time.Time{}
	if args.ExpireAt != nil {
		expireAt, _ = time.Parse("2006-01-02T15:04:05.999Z07:00", *args.ExpireAt)
	}

	ban := &datastructure.Ban{
		UserID:     &user.ID,
		Reason:     reasonN,
		IssuedByID: &usr.ID,
		ExpireAt:   expireAt,
	}

	_, err = mongo.Collection(mongo.CollectionNameBans).InsertOne(ctx, ban)
	if err != nil {
		logrus.Errorf("mongo, err=%v", err)
		return nil, resolvers.ErrInternalServer
	}

	actions.Bans.BannedUsers[id] = ban
	_, err = mongo.Collection(mongo.CollectionNameAudit).InsertOne(ctx, &datastructure.AuditLog{
		Type:      datastructure.AuditLogTypeUserBan,
		CreatedBy: usr.ID,
		Target:    &datastructure.Target{ID: &id, Type: "users"},
		Changes:   nil,
		Reason:    args.Reason,
	})

	if err != nil {
		logrus.Errorf("mongo, err=%v", err)
	}

	return &response{
		OK:      true,
		Status:  200,
		Message: "success",
	}, nil
}

//
// UNBAN USER
//

func (*MutationResolver) UnbanUser(ctx context.Context, args struct {
	VictimID string
	Reason   *string
}) (*response, error) {
	usr, ok := ctx.Value(utils.UserKey).(*datastructure.User)
	if !ok {
		return nil, resolvers.ErrLoginRequired
	}

	if !usr.HasPermission(datastructure.RolePermissionBanUsers) {
		return nil, resolvers.ErrAccessDenied
	}

	id, err := primitive.ObjectIDFromHex(args.VictimID)
	if err != nil {
		return nil, resolvers.ErrUnknownUser
	}

	if id.Hex() == usr.ID.Hex() {
		return nil, resolvers.ErrYourself
	}

	banned, _ := actions.Bans.IsUserBanned(id)
	if !banned {
		return nil, resolvers.ErrUserNotBanned
	}

	res := mongo.Collection(mongo.CollectionNameUsers).FindOne(ctx, bson.M{
		"_id": id,
	})

	user := &datastructure.User{}

	err = res.Err()

	if err == nil {
		err = res.Decode(user)
	}

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, resolvers.ErrUnknownUser
		}
		logrus.Errorf("mongo, err=%v", err)
		return nil, resolvers.ErrInternalServer
	}

	_, err = mongo.Collection(mongo.CollectionNameBans).UpdateMany(ctx, bson.M{
		"user_id":   user.ID,
		"expire_at": bson.M{"$not": bson.M{"$eq": time.Time{}}},
	}, bson.M{
		"$set": bson.M{
			"expire_at": time.Time{},
		},
	})
	if err != nil {
		logrus.Errorf("mongo, err=%v", err)
		return nil, resolvers.ErrInternalServer
	}

	delete(actions.Bans.BannedUsers, id)
	_, err = mongo.Collection(mongo.CollectionNameAudit).InsertOne(ctx, &datastructure.AuditLog{
		Type:      datastructure.AuditLogTypeUserUnban,
		CreatedBy: usr.ID,
		Target:    &datastructure.Target{ID: &id, Type: "users"},
		Changes:   nil,
		Reason:    args.Reason,
	})

	if err != nil {
		logrus.Errorf("mongo, err=%v", err)
	}

	return &response{
		OK:      true,
		Status:  200,
		Message: "success",
	}, nil
}
