package datastructure

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Entitlement is a binding between a resource and a user
// It grants the user access to the bound resource
// and may define some additional properties on top.
type Entitlement struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	// Kind represents what item this entitlement grants
	Kind EntitlementKind `json:"kind" bson:"kind"`
	// Data referencing the entitled item
	Data bson.Raw `json:"data" bson:"data"`
	// The user who is entitled to the item
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	// Wether this entitlement is currently inactive
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty"`
}

// A string representing an Entitlement Kind
type EntitlementKind string

var (
	EntitlementKindSubscription EntitlementKind // Subscription Entitlement
	EntitlementKindBadge        EntitlementKind // Badge Entitlement
	EntitlementKindRole         EntitlementKind // Role Entitlement
	EntitlementKindEmoteSet     EntitlementKind // Emote Set Entitlement
)

// (Data) Subscription binding in an Entitlement
type EntitledSubscription struct {
	// The ID of the subscription
	ItemID primitive.ObjectID `json:"item_id" bson:"item_id"`
}

// (Data) Badge binding in an Entitlement
type EntitledBadge struct {
	ItemID   primitive.ObjectID `json:"item_id" bson:"item_id"`
	Selected bool               `json:"selected" bson:"selected"`
}

// (Data) Role binding in an Entitlement
type EntitledRole struct {
	ItemID primitive.ObjectID `json:"item_id" bson:"item_id"`
	// Whether or not the entitlemet will cause the user's role to be overriden,
	// even if their current role has a higher position
	Override bool `json:"override" bson:"override"`
}

// (Data) Emote Set binding in an Entitlement
type EntitledEmoteSet struct {
	ItemID     primitive.ObjectID   `json:"item_id" bson:"item_id"`
	UnicodeTag string               `json:"unicode_tag" bson:"unicode_tag"`
	EmoteIDs   []primitive.ObjectID `json:"emote_ids" bson:"emotes"`

	// Relational

	// A list of emotes for this emote set entitlement
	Emotes []*Emote `json:"emotes" bson:"-"`
}
