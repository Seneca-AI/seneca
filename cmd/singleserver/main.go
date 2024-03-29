// Package main in singleserver starts the entire Seneca application on
// a single server utilizing channels to mimic HTTP request routing.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"seneca/api/constants"
	st "seneca/api/type"
	"seneca/env"
	"seneca/internal/authenticator"
	"seneca/internal/client/cloud/gcp"
	"seneca/internal/client/cloud/gcp/datastore"
	"seneca/internal/client/googledrive"
	"seneca/internal/client/intraseneca"
	senecahttp "seneca/internal/client/intraseneca/http"
	"seneca/internal/client/logging"
	weatherservice "seneca/internal/client/weather/service"
	"seneca/internal/controller/apiserver"
	"seneca/internal/controller/runner"
	"seneca/internal/controller/syncer"
	"seneca/internal/dao"
	"seneca/internal/dao/drivingconditiondao"
	"seneca/internal/dao/eventdao"
	"seneca/internal/dao/rawframedao"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/dao/tripdao"
	"seneca/internal/dao/userdao"
	"seneca/internal/dataaggregator/sanitizer"
	"seneca/internal/datagatherer/rawvideohandler"
	"seneca/internal/dataprocessor"
	"seneca/internal/dataprocessor/algorithms"
	"seneca/internal/util"
	"seneca/internal/util/mp4"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	port = "6060"
)

func main() {
	if err := env.ValidateEnvironmentVariables(); err != nil {
		log.Fatalf("Error in ValidateEnvironmentVariables: %v", err)
	}

	serverConfig := &intraseneca.ServerConfig{
		MLServerHostName: "34.136.176.46",
		MLServerHostPort: "5000",
		MLServerTimeout:  time.Second * 60,
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		fmt.Fprintf(os.Stderr, "GOOGLE_CLOUD_PROJECT environment variable must be set.\n")
		os.Exit(1)
	}

	logger, err := logging.NewGCPLogger(context.TODO(), "singleserver", projectID)
	if err != nil {
		fmt.Printf("logging.NewGCPLogger() returns - err: %v", err)
		return
	}

	intraSenecaClient, err := senecahttp.New(serverConfig)
	if err != nil {
		logger.Critical(fmt.Sprintf("senecahttp.New() returns - err: %v", err))
		return
	}

	logger.Log(fmt.Sprintf("Starting singleserver"))

	gcsc, err := gcp.NewGoogleCloudStorageClient(context.TODO(), projectID, time.Second*10, time.Minute)
	if err != nil {
		logger.Critical(fmt.Sprintf("cloud.NewGoogleCloudStorageClient() returns - err: %v", err))
		return
	}

	sqlService, err := datastore.New(context.TODO(), projectID)
	if err != nil {
		logger.Critical(fmt.Sprintf("datastore.New() returns - err: %v", err))
		return
	}

	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(sqlService, logger, (time.Second * 5))
	rawLocationDAO := rawlocationdao.NewSQLRawLocationDAO(sqlService)
	rawMotionDAO := rawmotiondao.NewSQLRawMotionDAO(sqlService, logger)
	rawFrameDAO := rawframedao.NewSQLRawFrameDAO(sqlService)
	userDAO := userdao.NewSQLUserDAO(sqlService)
	tripDAO := tripdao.NewSQLTripDAO(sqlService, logger)
	eventDAO := eventdao.NewSQLEventDAO(sqlService, tripDAO, logger)
	drivingConditionDAO := drivingconditiondao.NewSQLDrivingConditionDAO(sqlService, tripDAO, eventDAO)
	allDAOSet := &dao.AllDAOSet{
		UserDAO:             userDAO,
		RawVideoDAO:         rawVideoDAO,
		RawLocationDAO:      rawLocationDAO,
		RawMotionDAO:        rawMotionDAO,
		RawFrameDAO:         rawFrameDAO,
		TripDAO:             tripDAO,
		EventDAO:            eventDAO,
		DrivingConditionDAO: drivingConditionDAO,
	}

	mp4Tool, err := mp4.NewMP4Tool(logger)
	if err != nil {
		logger.Critical(fmt.Sprintf("mp4.NewMP4Tool() returns - err: %v", err))
		return
	}
	rawVideoHandler, err := rawvideohandler.NewRawVideoHandler(gcsc, mp4Tool, rawVideoDAO, rawLocationDAO, rawMotionDAO, rawFrameDAO, logger, projectID)
	if err != nil {
		logger.Critical(fmt.Sprintf("cloud.NewRawVideoHandler() returns - err: %v", err))
		return
	}

	gDriveFactory := &googledrive.UserClientFactory{}
	syncer := syncer.New(rawVideoHandler, gDriveFactory, userDAO, logger)

	algoFactory, err := algorithms.NewFactory(weatherservice.NewWeatherStackService(time.Second*10), intraSenecaClient)
	if err != nil {
		logger.Critical(fmt.Sprintf("algorithms.NewFactory() returns err: %v", err))
		return
	}
	algos := []dataprocessor.AlgorithmInterface{}
	algoTags := []string{"00000", "00001", "00002", "00003", "00004"}
	for _, tag := range algoTags {
		algo, err := algoFactory.GetAlgorithm(tag)
		if err != nil {
			logger.Critical(fmt.Sprintf("GetAlgorithm(%q) returns err: %v", tag, err))
		}
		algos = append(algos, algo)
	}

	dataprocessor, err := dataprocessor.New(algos, allDAOSet, logger)
	if err != nil {
		logger.Critical(fmt.Sprintf("dataprocessor.New() returns - err: %v", err))
		return
	}
	runner := runner.New(userDAO, dataprocessor, logger)
	sanitizer := sanitizer.New(rawMotionDAO, rawLocationDAO, rawVideoDAO, rawFrameDAO, eventDAO, drivingConditionDAO)
	apiserver := apiserver.New(sanitizer, tripDAO)

	handler := &HTTPHandler{
		syncer:              syncer,
		runner:              runner,
		eventDAO:            eventDAO,
		drivingconditionDAO: drivingConditionDAO,
		apiserver:           apiserver,
		logger:              logger,
	}

	http.HandleFunc("/", handler.handleHTTP)

	fmt.Printf("Starting server at port %s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}

type HTTPHandler struct {
	syncer              *syncer.Syncer
	runner              *runner.Runner
	eventDAO            dao.EventDAO
	drivingconditionDAO dao.DrivingConditionDAO
	apiserver           *apiserver.APIServer
	logger              logging.LoggingInterface
}

func (handler *HTTPHandler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	handler.logger.Log(fmt.Sprintf("Received %s request to %s", r.Method, r.URL))

	if err := authenticator.AuthorizeHTTPRequest(w, r); err != nil {
		return
	}

	if matchesRoute(fmt.Sprintf("/%s", constants.HeartbeatEndpoint), r.URL.Path) {
		handler.handleHeartbeat(w, r)
	} else if matchesRoute("/syncer", r.URL.Path) {
		handler.runSyncer(w, r)
	} else if matchesRoute("/runner", r.URL.Path) {
		handler.runRunner(w, r)
	} else if matchesRoute("/users/*/events", r.URL.Path) {
		handler.handleEventRequest(w, r)
	} else if matchesRoute("/users/*/driving_conditions", r.URL.Path) {
		handler.handleDrivingConditionRequest(w, r)
	} else if matchesRoute("/users/*/trips", r.URL.Path) {
		handler.handleTripsRequest(w, r)
	} else {
		fmt.Fprintf(w, "Unsupported request URL path.  Refer to discovery/discovery.json in the common repo.")
		w.WriteHeader(400)
	}
}

func (handler *HTTPHandler) runSyncer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Fprintf(w, "/syncer only supports POST methods")
		w.WriteHeader(400)
		return
	}
	go handler.syncer.ScanAllUsers()
	w.WriteHeader(200)
}

func (handler *HTTPHandler) runRunner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Fprintf(w, "/runner only supports POST methods")
		w.WriteHeader(400)
		return
	}
	go handler.runner.Run()
	w.WriteHeader(200)
}

func (handler *HTTPHandler) handleTripsRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fmt.Fprintf(w, "/users/*/trips only supports GET methods")
		w.WriteHeader(400)
		return
	}

	response, err := func() (*st.TripListResponse, error) {
		request := &st.TripListRequest{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to parse body of request")
		}

		if err := proto.UnmarshalText(string(bodyBytes), request); err != nil {
			return nil, fmt.Errorf("unable to unmarshal request body into TripListRequest")
		}

		trips, err := handler.apiserver.ListTrips(request.UserId, util.MillisecondsToTime(request.StartTimeMs), util.MillisecondsToTime(request.EndTimeMs))
		if err != nil {
			return nil, fmt.Errorf("error listing trips: %w", err)
		}

		return &st.TripListResponse{
			UserId: request.UserId,
			Trip:   trips,
		}, nil
	}()

	if err != nil {
		response = &st.TripListResponse{
			Header: &st.Header{
				Code:    400,
				Message: fmt.Sprintf("Error: %v", err),
			},
		}
	} else {
		response.Header = &st.Header{
			Code: 200,
		}
	}

	buffer := &bytes.Buffer{}
	err = proto.MarshalText(buffer, response)
	if err != nil {
		fmt.Fprintf(w, "Error marshalling response body: %v", err)
		w.WriteHeader(500)
		return
	}

	w.Write(buffer.Bytes())
	w.WriteHeader(int(response.Header.Code))
}

func (handler *HTTPHandler) handleEventRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Fprintf(w, "/users/*/events only supports POST methods")
		w.WriteHeader(400)
		return
	}

	response, err := func() (*st.EventCreateResponse, error) {
		request := &st.EventCreateRequest{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to parse body of request")
		}

		if err := proto.UnmarshalText(string(bodyBytes), request); err != nil {
			return nil, fmt.Errorf("unable to unmarshal request body into EventCreateRequest")
		}

		eventWithID, err := handler.eventDAO.CreateEvent(context.TODO(), request.Event)
		if err != nil {
			return nil, fmt.Errorf("error creating event: %w", err)
		}

		return &st.EventCreateResponse{
			UserId: request.UserId,
			Event:  eventWithID,
		}, nil
	}()

	if err != nil {
		response = &st.EventCreateResponse{
			Header: &st.Header{
				Code:    400,
				Message: fmt.Sprintf("Error: %v", err),
			},
		}
	} else {
		response.Header = &st.Header{
			Code: 200,
		}
	}

	buffer := &bytes.Buffer{}
	err = proto.MarshalText(buffer, response)
	if err != nil {
		fmt.Fprintf(w, "Error marshalling response body: %v", err)
		w.WriteHeader(500)
		return
	}

	w.Write(buffer.Bytes())
	w.WriteHeader(int(response.Header.Code))
}

func (handler *HTTPHandler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func (handler *HTTPHandler) handleDrivingConditionRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Fprintf(w, "/users/*/driving_conditions only supports POST methods")
		w.WriteHeader(400)
		return
	}

	response, err := func() (*st.DrivingConditionCreateResponse, error) {
		request := &st.DrivingConditionCreateRequest{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to parse body of request")
		}

		if err := proto.UnmarshalText(string(bodyBytes), request); err != nil {
			return nil, fmt.Errorf("unable to unmarshal request body into DrivingConditionCreateRequest")
		}

		drivingConditinWithID, err := handler.drivingconditionDAO.CreateDrivingCondition(context.TODO(), request.DrivingCondition)
		if err != nil {
			return nil, fmt.Errorf("error creating drivingCondition: %w", err)
		}

		return &st.DrivingConditionCreateResponse{
			UserId:           request.UserId,
			DrivingCondition: drivingConditinWithID,
		}, nil
	}()

	if err != nil {
		response = &st.DrivingConditionCreateResponse{
			Header: &st.Header{
				Code:    400,
				Message: fmt.Sprintf("Error: %v", err),
			},
		}
	} else {
		response.Header = &st.Header{
			Code: 200,
		}
	}

	buffer := &bytes.Buffer{}
	err = proto.MarshalText(buffer, response)
	if err != nil {
		fmt.Fprintf(w, "Error marshalling response body: %v", err)
		w.WriteHeader(500)
		return
	}

	w.Write(buffer.Bytes())
	w.WriteHeader(int(response.Header.Code))
}

func matchesRoute(route, requestURLPath string) bool {
	routeParts := strings.Split(route, "/")
	requestURLPathParts := strings.Split(requestURLPath, "/")

	if len(routeParts) != len(requestURLPathParts) {
		return false
	}

	for i := range routeParts {
		if routeParts[i] != "*" && requestURLPathParts[i] != routeParts[i] {
			return false
		}
	}

	return true
}
