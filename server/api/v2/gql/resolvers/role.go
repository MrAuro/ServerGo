package resolvers

import (
	"context"
	"fmt"

	"github.com/SevenTV/ServerGo/cache"
	"github.com/SevenTV/ServerGo/mongo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type roleResolver struct {
	ctx context.Context
	v   *mongo.Role

	fields map[string]*SelectedField
}

func GenerateRoleResolver(ctx context.Context, roleID *primitive.ObjectID, fields map[string]*SelectedField) (*roleResolver, error) {
	role := &mongo.Role{}

	if role.ID.IsZero() {
		if err := cache.FindOne("roles", "", bson.M{
			"_id": roleID,
		}, role); err != nil {
			if err != mongo.ErrNoDocuments {
				log.Errorf("mongo, err=%v", err)
				return nil, errInternalServer
			}
			return nil, nil
		}
	}

	if role == nil {
		return nil, nil
	}

	r := &roleResolver{
		ctx:    ctx,
		v:      role,
		fields: fields,
	}
	return r, nil
}

func (r *roleResolver) ID() string {
	return r.v.ID.Hex()
}

func (r *roleResolver) Name() string {
	return r.v.Name
}

func (r *roleResolver) Color() int32 {
	return r.v.Color
}

func (r *roleResolver) Allowed() string {
	return fmt.Sprint(r.v.Allowed)
}

func (r *roleResolver) Denied() string {
	return fmt.Sprint(r.v.Denied)
}