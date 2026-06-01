package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestListTags_Unit(t *testing.T) {
	mockTags := &OrganizationTagsAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.OrganizationTagsListOptions) (*tfe.OrganizationTagsList, error) {
			if options.PageNumber == 1 {
				return &tfe.OrganizationTagsList{
					Items: []*tfe.OrganizationTag{
						{ID: "tag-1", Name: "Tag 1", InstanceCount: 5},
					},
					Pagination: &tfe.Pagination{
						NextPage: 2,
					},
				}, nil
			}
			return &tfe.OrganizationTagsList{
				Items: []*tfe.OrganizationTag{
					{ID: "tag-2", Name: "Tag 2", InstanceCount: 10},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	tags, err := listTags(mockTags, "my-org", "", "")

	assert.NoError(t, err)
	assert.Len(t, tags, 2)
	assert.Equal(t, "Tag 1", tags[0].Name)
	assert.Equal(t, "Tag 2", tags[1].Name)
}

func TestListTags_Filter_Unit(t *testing.T) {
	mockTags := &OrganizationTagsAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.OrganizationTagsListOptions) (*tfe.OrganizationTagsList, error) {
			assert.Equal(t, "ws-123", options.Filter)
			assert.Equal(t, "search-query", options.Query)
			return &tfe.OrganizationTagsList{
				Items: []*tfe.OrganizationTag{
					{ID: "tag-1", Name: "Filtered Tag"},
				},
				Pagination: &tfe.Pagination{NextPage: 0},
			}, nil
		},
	}

	tags, err := listTags(mockTags, "my-org", "ws-123", "search-query")
	assert.NoError(t, err)
	assert.Len(t, tags, 1)
}

func TestTagStruct_Unit(t *testing.T) {
	tag := Tag{
		ID:            "tag-id",
		Name:          "tag-name",
		InstanceCount: 42,
	}
	assert.Equal(t, "tag-id", tag.ID)
	assert.Equal(t, "tag-name", tag.Name)
	assert.Equal(t, 42, tag.InstanceCount)
}
