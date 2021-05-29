package helpers

import (
	"context"
	"fmt"
	st "seneca/api/type"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/cenkalti/backoff"
)

const (
	pushAttempts = 2
	gitHubToken  = "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"
)

// TODO(lucaloncar): use two servers and alernate push so we don't have any downtime

type Pusher struct {
	coreProjectID   string
	datastoreClient *datastore.Client
}

func NewPusher(coreProjectID string) (*Pusher, error) {
	client, err := datastore.NewClient(context.TODO(), coreProjectID)
	if err != nil {
		return nil, fmt.Errorf("error initializing datastore client: %v", err)
	}
	return &Pusher{
		coreProjectID:   coreProjectID,
		datastoreClient: client,
	}, nil
}

func (psh *Pusher) Run() error {
	senecaServers, err := listSenecaServers(psh.datastoreClient)
	if err != nil {
		return fmt.Errorf("error listing senecaServers: %w", err)
	}

	for _, ss := range senecaServers {
		if !ss.ReceiveMainPushes {
			continue
		}

		b := backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second*10), pushAttempts)
		operation := func() error {
			return push(ss)
		}

		if err := backoff.Retry(operation, b); err != nil {
			return fmt.Errorf("server with VM name %q in project %q cannot be restarted after %d attempts. last err: %w", ss.ServerVmName, ss.ProjectId, pushAttempts, err)
		}
	}

	return nil
}

func push(ss *st.SenecaServer) error {
	if err := pullRepo(ss); err != nil {
		return fmt.Errorf("error pulling repo: %w", err)
	}

	if err := killServer(ss); err != nil {
		// The server may already be did, so just print a warning.
		fmt.Printf("Error killing server: %v\n", err)
	}

	if err := startServer(ss); err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}

func pullRepo(ss *st.SenecaServer) error {
	pullCommand := fmt.Sprintf("gcloud compute ssh lucaloncar@%s --zone %s --project %s -- -T cd seneca && git pull https://%s@github.com/Seneca-AI/seneca.git", ss.ServerVmName, ss.ServerVmZone, ss.ProjectId, gitHubToken)
	if _, err := gcloudExec(pullCommand); err != nil {
		return err
	}

	return nil
}

func killServer(ss *st.SenecaServer) error {
	getPIDCommand := fmt.Sprintf("gcloud compute ssh lucaloncar@%s --zone %s --project %s -- -T pgrep -f 'singleserver'", ss.ServerVmName, ss.ServerVmZone, ss.ProjectId)

	output, err := gcloudExec(getPIDCommand)
	if err != nil {
		return err
	}

	// Output should have 1 PID, and the message "connection closed."
	if len(output) < 2 {
		return fmt.Errorf("no PIDs returned")
	}

	killCommand := fmt.Sprintf("gcloud compute ssh lucaloncar@%s --zone %s --project %s -- -T sudo kill -9 %s", ss.ServerVmName, ss.ServerVmZone, ss.ProjectId, output[0])
	if _, err := gcloudExec(killCommand); err != nil {
		return err
	}

	return nil
}
