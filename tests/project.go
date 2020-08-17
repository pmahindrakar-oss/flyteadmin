// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

func TestCreateProject(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	req := admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:   "potato",
			Name: "spud",
		},
	}

	_, err := client.RegisterProject(ctx, &req)
	assert.Nil(t, err)

	projects, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, projects.Projects)
	var sawNewProject bool
	for _, project := range projects.Projects {
		if project.Id == "potato" {
			sawNewProject = true
			assert.Equal(t, "spud", project.Name)
		}
		// Assert all projects include expected system-wide domains
		for _, domain := range project.Domains {
			assert.Contains(t, []string{"development", "domain", "staging", "production"}, domain.Id)
			assert.Contains(t, []string{"development", "domain", "staging", "production"}, domain.Name)
		}
	}
	assert.True(t, sawNewProject)
}

func TestUpdateProjectDescription(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	// Create a new project.
	req := admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:   "potato",
			Name: "spud",
			Labels: &admin.Labels{
				Values: map[string]string{
					"foo": "bar",
					"bar": "baz",
				},
			},
		},
	}
	_, err := client.RegisterProject(ctx, &req)
	assert.Nil(t, err)

	// Verify the project has been registered.
	projects, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, projects.Projects)

	// Attempt to modify the name of the Project. Modifying the Name should be a
	// no-op, while the Description is modified. Labels should be a no-op.
	_, err = client.UpdateProject(ctx, &admin.Project{
		Id: "potato",
		Name: "foobar",
		Description: "a-new-description",
		Labels: &admin.Labels{},
	})

	// Fetch updated projects.
	projectsUpdated, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, projectsUpdated.Projects)

	// Verify that the project's Name has not been modified but the Description has.
	updatedProject := projectsUpdated.Projects[0]
	assert.Equal(t, updatedProject.Id, "potato")
	assert.Equal(t, updatedProject.Name, "spud") // unchanged
	assert.Equal(t, updatedProject.Description, "a-new-description") // changed

	// Verify that project labels are not removed.
	labelsMap := updatedProject.Labels
	fooVal, fooExists := labelsMap.Values["foo"]
	barVal, barExists := labelsMap.Values["bar"]
	assert.Equal(t, fooExists, true)
	assert.Equal(t, fooVal, "bar")
	assert.Equal(t, barExists, true)
	assert.Equal(t, barVal, "baz")
}

func TestUpdateProjectLabels(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	// Create a new project.
	req := admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:   "potato",
			Name: "spud",
		},
	}
	_, err := client.RegisterProject(ctx, &req)
	assert.Nil(t, err)

	// Verify the project has been registered.
	projects, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, projects.Projects)

	// Attempt to modify the name of the Project. Modifying the Name should be a
	// no-op, while the Labels are modified.
	_, err = client.UpdateProject(ctx, &admin.Project{
		Id: "potato",
		Name: "foobar",
		Labels: &admin.Labels{
			Values: map[string]string{
				"foo": "bar",
				"bar": "baz",
			},
		},
	})

	// Fetch updated projects.
	projectsUpdated, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, projectsUpdated.Projects)

	// Verify that the expected labels have been added to the project.
	updatedProject := projectsUpdated.Projects[0]
	labelsMap := updatedProject.Labels
	fooVal, fooExists := labelsMap.Values["foo"]
	barVal, barExists := labelsMap.Values["bar"]
	assert.Equal(t, fooExists, true)
	assert.Equal(t, fooVal, "bar")
	assert.Equal(t, barExists, true)
	assert.Equal(t, barVal, "baz")
}
