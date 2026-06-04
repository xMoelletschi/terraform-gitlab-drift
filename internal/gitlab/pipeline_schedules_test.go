package gitlab

import (
	"context"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
	gitlabtesting "gitlab.com/gitlab-org/api/client-go/v2/testing"
	"go.uber.org/mock/gomock"
)

// ListPipelineSchedules fetches the schedule list per project and then a detail
// per schedule (the N+1). With multiple projects fetched concurrently this
// verifies the results compose per project and, under -race, that the
// concurrent writes are race-free.
func TestListPipelineSchedules_ComposesPerProject(t *testing.T) {
	tc := gitlabtesting.NewTestClient(t)
	c := NewClientFromAPI(tc.Client, "mygroup")

	tc.MockPipelineSchedules.EXPECT().
		ListPipelineSchedules(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*gl.PipelineSchedule{{ID: 10, Description: "nightly"}}, &gl.Response{}, nil).
		Times(2)
	tc.MockPipelineSchedules.EXPECT().
		GetPipelineSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&gl.PipelineSchedule{
			ID:          10,
			Description: "nightly",
			Variables:   []*gl.PipelineVariable{{Key: "K", Value: "V"}},
		}, &gl.Response{}, nil).
		Times(2)

	projects := []*gl.Project{
		{ID: 1, PathWithNamespace: "mygroup/p1"},
		{ID: 2, PathWithNamespace: "mygroup/p2"},
	}

	result, err := c.ListPipelineSchedules(context.Background(), projects)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected schedules for 2 projects, got %d", len(result))
	}
	for _, id := range []int64{1, 2} {
		if len(result[id]) != 1 {
			t.Fatalf("project %d: expected 1 schedule, got %d", id, len(result[id]))
		}
		if len(result[id][0].Variables) != 1 {
			t.Errorf("project %d: schedule detail (variables) not fetched", id)
		}
	}
}
