package helpers

import (
	"context"
	"fmt"
	"net/http"
	"seneca/api/constants"
	st "seneca/api/type"
	"seneca/internal/authenticator"
	"seneca/internal/client/logging"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/cenkalti/backoff"
)

const (
	goPath          = "$HOME/go"
	path            = "$PATH:$HOME/go/bin"
	restartAttempts = 5
)

type Medic struct {
	coreProjectID   string
	datastoreClient *datastore.Client
	logger          logging.LoggingInterface
}

func NewMedic(coreProjectID string, logger logging.LoggingInterface) (*Medic, error) {
	client, err := datastore.NewClient(context.TODO(), coreProjectID)
	if err != nil {
		return nil, fmt.Errorf("error initializing datastore client: %v", err)
	}
	return &Medic{
		coreProjectID:   coreProjectID,
		datastoreClient: client,
		logger:          logger,
	}, nil
}

func (med *Medic) Run() error {
	senecaServers, err := listSenecaServers(med.datastoreClient)
	if err != nil {
		return fmt.Errorf("error listing senecaServers: %w", err)
	}

	var lastErr error
	lastErr = nil
	for _, ss := range senecaServers {
		err := func() error {
			if err := checkHeartbeat(ss); err != nil {
				fmt.Printf("checkHeartbeat() for server with VM name %q in project %q returns err: %v\n", ss.ServerVmName, ss.ProjectId, err)

				b := backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second*10), restartAttempts)
				operation := func() error {
					return restartVMAndServer(ss)
				}

				if err := backoff.Retry(operation, b); err != nil {
					return fmt.Errorf("server cannot be restarted after %d attempts. last err: %w", restartAttempts, err)
				}

				newExternalIP, err := getNewExternalIP(ss)
				if err != nil {
					return fmt.Errorf("error getting new external IP: %w", err)
				}
				ss.ServerExternalIp = newExternalIP

				if err := updateSenecaServer(med.datastoreClient, ss); err != nil {
					return fmt.Errorf("error updating seneca server: %w", err)
				}

				if err := checkHeartbeat(ss); err != nil {
					return fmt.Errorf("server is still not responding to heartbeat after successful restart - err: %w", err)
				}
			}
			return nil
		}()
		if err != nil {
			lastErr = err
			med.logger.Critical(fmt.Sprintf("Outage for server with vm Name %q in project %q: %v", ss.ServerVmName, ss.ProjectId, err))
		}
	}
	return lastErr
}

func checkHeartbeat(ss *st.SenecaServer) error {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s/%s", ss.ServerExternalIp, ss.ServerPort, constants.HeartbeatEndpoint), nil)
	if err != nil {
		return fmt.Errorf("error initializing GET request: %w", err)
	}

	req = authenticator.AddRequestAuth(req)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("heartbeat returns status %q", resp.Status)
	}

	fmt.Printf("Server responded to heartbeat message")
	return nil
}

func restartVMAndServer(ss *st.SenecaServer) error {
	stopCommand := fmt.Sprintf("gcloud compute instances stop %s --zone %s --project %s", ss.ServerVmName, ss.ServerVmZone, ss.ProjectId)

	if _, err := gcloudExec(stopCommand); err != nil {
		return err
	}

	// Give it some time before restarting.
	time.Sleep(time.Second * 10)

	startCommand := fmt.Sprintf("gcloud compute instances start %s --zone %s --project %s", ss.ServerVmName, ss.ServerVmZone, ss.ProjectId)
	if _, err := gcloudExec(startCommand); err != nil {
		return err
	}

	// Give it some time before starting server.
	time.Sleep(time.Second * 30)

	return startServer(ss)
}

func getNewExternalIP(ss *st.SenecaServer) (string, error) {
	getInstanceInfo := fmt.Sprintf("gcloud compute instances describe %s --project %s --zone %s", ss.ServerVmName, ss.ProjectId, ss.ServerVmZone)

	output, err := gcloudExec(getInstanceInfo)
	if err != nil {
		return "", err
	}

	found := false
	externalIP := ""
	for _, line := range output {
		if strings.Contains(line, "natIP") {
			found = true
			lineParts := strings.Split(line, " ")
			if len(lineParts) != 2 {
				return "", fmt.Errorf("want a line with 2 parts when split on ' ', got line with %d parts: %q", len(lineParts), line)
			}
			externalIP = lineParts[1]
			break
		}
	}

	if !found {
		return "", fmt.Errorf("failed to find new external IP")
	}

	return externalIP, nil
}

func updateSenecaServer(datastoreClient *datastore.Client, ss *st.SenecaServer) error {
	idKey := datastore.IDKey(senecaServerKey.Kind, ss.DatastoreId, &senecaServerKey)
	if _, err := datastoreClient.Put(context.TODO(), idKey, ss); err != nil {
		return fmt.Errorf("datastoreClient.Put() returns err: %w", err)
	}
	return nil
}
