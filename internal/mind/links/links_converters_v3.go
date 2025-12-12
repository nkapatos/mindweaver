package links

import (
	mindv3 "github.com/nkapatos/mindweaver/internal/mind/gen/v3"
	"github.com/nkapatos/mindweaver/internal/mind/store"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// StoreLinkToProto converts a store.NotesLink to a proto Link
func StoreLinkToProto(link store.NotesLink) *mindv3.Link {
	proto := &mindv3.Link{
		Name:       "links/" + string(link.ID),
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

// StoreLinksToProto converts a slice of store.NotesLink to proto Links
func StoreLinksToProto(links []store.NotesLink) []*mindv3.Link {
	if links == nil {
		return nil
	}

	result := make([]*mindv3.Link, len(links))
	for i, link := range links {
		result[i] = StoreLinkToProto(link)
	}
	return result
}
