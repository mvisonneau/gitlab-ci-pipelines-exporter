package exporter

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
	"github.com/vmihailenco/taskq/v3/redisq"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/storage"
)

var (
	// Config ..
	Config schemas.Config

	gitlabClient *gitlab.Client
	redisClient  *redis.Client
	taskFactory  taskq.Factory
	pollingQueue taskq.Queue
	store        storage.Storage
)

// ConfigureGitlabClient ..
func ConfigureGitlabClient(c *gitlab.Client) {
	gitlabClient = c
}

// ConfigureRedisClient ..
func ConfigureRedisClient(c *redis.Client) error {
	redisClient = c
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		return errors.Wrap(err, "connecting to redis")
	}
	return nil
}

// ConfigurePollingQueue ..
func ConfigurePollingQueue() {
	pollingQueueOptions := &taskq.QueueOptions{
		Name: "poll",
	}

	if redisClient != nil {
		taskFactory = redisq.NewFactory()
		pollingQueueOptions.Redis = redisClient
	} else {
		taskFactory = memqueue.NewFactory()
	}

	pollingQueue = taskFactory.RegisterQueue(pollingQueueOptions)

	// Purge the queue when we start
	// I am only partially convinced this will not cause issues in HA fashion
	if err := pollingQueue.Purge(); err != nil {
		log.WithField("error", err.Error()).Error("purging the polling queue")
	}
}

// ConfigureStore ..
func ConfigureStore() {
	if redisClient != nil {
		store = storage.NewRedisStorage(redisClient)
	} else {
		store = storage.NewLocalStorage()
	}

	// Load all the configured projects in the store
	for _, p := range Config.Projects {
		if err := store.SetProject(p); err != nil {
			log.WithFields(log.Fields{
				"project-name": p.Name,
				"error":        err.Error(),
			}).Error("writing project in the store")
		}
	}
}

// ProcessPollingQueue ..
func ProcessPollingQueue(ctx context.Context) {
	if redisClient != nil {
		if err := taskFactory.StartConsumers(ctx); err != nil {
			log.Fatal(err)
		}
	}
}
