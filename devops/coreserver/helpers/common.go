package helpers

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	st "seneca/api/type"
	"strings"

	"cloud.google.com/go/datastore"
)

var (
	senecaServerKey = datastore.Key{
		Kind: "SenecaServer",
		Name: "SenecaServers",
	}
)

func listSenecaServers(datastoreClient *datastore.Client) ([]*st.SenecaServer, error) {
	keys, err := datastoreClient.GetAll(context.TODO(), datastore.NewQuery(senecaServerKey.Kind).KeysOnly(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing keys: %w", err)
	}

	serverList := []*st.SenecaServer{}
	for _, k := range keys {
		ss := &st.SenecaServer{}
		if err := datastoreClient.Get(context.TODO(), k, ss); err != nil {
			return nil, fmt.Errorf("datastoreClient.Get(_, %s, _) returns err: %w", k, err)
		}
		ss.DatastoreId = k.ID
		serverList = append(serverList, ss)
	}
	return serverList, nil
}

func startServer(ss *st.SenecaServer) error {
	startProcessCommand := fmt.Sprintf("gcloud compute ssh lucaloncar@%s --zone %s --project %s -- cd seneca/devops/setup && bash setup.sh start_singleserver %s %s %s", ss.ServerVmName, ss.ServerVmZone, ss.ProjectId, ss.ProjectId, ss.ServerPathToGoogleApplicationCredentials, ss.ServerPathToGoogleOauthCredentials)
	if _, err := gcloudExec(startProcessCommand); err != nil {
		return err
	}

	return nil
}

func PushSenecaServerObject(datastoreClient *datastore.Client, ss *st.SenecaServer) error {
	incompleteKey := datastore.IncompleteKey(senecaServerKey.Kind, &senecaServerKey)
	if _, err := datastoreClient.Put(context.TODO(), incompleteKey, ss); err != nil {
		return fmt.Errorf("Put() returns err: %w", err)
	}
	return nil
}

func gcloudExec(command string) ([]string, error) {
	fmt.Printf("Running command %q\n", command)

	commmandParts := strings.Split(command, " ")
	cmd := exec.Command(commmandParts[0], commmandParts[1:]...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error output: %s\n", stderr.String())
		return nil, fmt.Errorf("error running command %s. OUTPUT: %s , ERR: %w", command, stderr.String(), err)
	}

	fmt.Printf("Output: %s\n", out.String())

	return strings.Split(out.String(), "\n"), nil
}
