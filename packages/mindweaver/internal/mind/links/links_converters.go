package links

import (
	"fmt"

	mindv3 "github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/gen/store"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// StoreLinkToProto converts a store.Link to a proto Link
func StoreLinkToProto(link store.Link) *mindv3.Link {
	proto := &mindv3.Link{
		Name:       fmt.Sprintf("links/%d", link.ID),
		Id:         link.ID,
		SrcId:      link.SrcID,
		CreateTime: timestamppb.New(link.CreatedAt.Time),
		UpdateTime: timestamppb.New(link.UpdatedAt.Time),
	}

	if link.DestID.Valid {
		proto.DestId = &link.DestID.Int64
	}

	if link.DestTitle.Valid {
		proto.DestTitle = &link.DestTitle.String
	}

	if link.DisplayText.Valid {
		proto.DisplayText = &link.DisplayText.String
	}

	if link.IsEmbed.Valid {
		proto.IsEmbed = &link.IsEmbed.Bool
	}

	if link.Resolved.Valid {
		proto.Resolved = &link.Resolved.Int64
	}

	return proto
}

// StoreLinksToProto converts a slice of store.Link to proto Links
func StoreLinksToProto(links []store.Link) []*mindv3.Link {
	if links == nil {
		return nil
	}

	result := make([]*mindv3.Link, len(links))
	for i, link := range links {
		result[i] = StoreLinkToProto(link)
	}
	return result
}
