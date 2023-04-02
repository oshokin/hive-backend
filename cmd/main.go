package main

import (
	"context"

	"github.com/oshokin/hive-backend/internal/app"
	"github.com/oshokin/hive-backend/internal/common"
	"github.com/oshokin/hive-backend/internal/logger"
)

func main() {
	ctx := context.Background()

	a, err := app.NewApplication(ctx)
	if err != nil {
		logger.FatalKV(ctx, "failed to create app", common.ErrorTag, err)
	}

	a.Run(ctx)
}
