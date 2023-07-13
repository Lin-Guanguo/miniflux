// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package worker // import "miniflux.app/worker"

import (
	"math/rand"
	"time"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/metric"
	"miniflux.app/model"
	feedHandler "miniflux.app/reader/handler"
	"miniflux.app/storage"
)

// Worker refreshes a feed in the background.
type Worker struct {
	id    int
	store *storage.Storage
}

// Run wait for a job and refresh the given feed.
func (w *Worker) Run(c chan model.Job) {
	logger.Debug("[Worker] #%d started", w.id)

	for {
		job := <-c
		logger.Debug("[Worker #%d] Received feed #%d for user #%d", w.id, job.FeedID, job.UserID)

		startTime := time.Now()
		refreshErr := feedHandler.RefreshFeed(w.store, job.UserID, job.FeedID)

		if config.Opts.HasMetricsCollector() {
			status := "success"
			if refreshErr != nil {
				status = "error"
			}
			metric.BackgroundFeedRefreshDuration.WithLabelValues(status).Observe(time.Since(startTime).Seconds())
		}

		if refreshErr != nil {
			go func() {
				for i := 0; i < 3; i++ {
					retryDelay := time.Duration(rand.Intn(60*5) + 1) // 生成5分钟的随机延迟
					logger.Error("[Worker] Refreshing the feed #%d returned this error: %v. will retry(%d) after %d seconds", job.FeedID, refreshErr, i, retryDelay)
					time.Sleep(retryDelay * time.Second)
					refreshErr := feedHandler.RefreshFeed(w.store, job.UserID, job.FeedID)
					if refreshErr == nil {
						return
					}
				}
				logger.Error("[Worker] Refreshing the feed #%d returned this error: %v. already retry 3 times", job.FeedID, refreshErr)
			}()
		}

		time.Sleep(time.Duration(5) * time.Second)
	}
}
