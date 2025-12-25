package tags

import (
	"fmt"

	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func StoreTagToProto(tag store.Tag) *mindv3.Tag {
	proto := &mindv3.Tag{
		Name:        fmt.Sprintf("tags/%d", tag.ID),
		Id:          tag.ID,
		DisplayName: tag.Name,
	}

	if tag.CreatedAt.Valid {
		proto.CreateTime = timestamppb.New(tag.CreatedAt.Time)
	}
	if tag.UpdatedAt.Valid {
		proto.UpdateTime = timestamppb.New(tag.UpdatedAt.Time)
	}

	return proto
}

func StoreTagsToProto(tags []store.Tag) []*mindv3.Tag {
	result := make([]*mindv3.Tag, len(tags))
	for i, tag := range tags {
		result[i] = StoreTagToProto(tag)
	}
	return result
}
