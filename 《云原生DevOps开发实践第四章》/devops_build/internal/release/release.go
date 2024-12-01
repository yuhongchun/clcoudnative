package release

import (
	"context"

	"devops_build/internal/release/model"
)

type ReleaseInfo interface {
	GetDockerInfo(ctx context.Context, params interface{}) *model.DockerInfo
}
