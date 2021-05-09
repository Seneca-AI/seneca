package gcp

// const (
// 	// "kind" is a Cloud Datstore concept.
// 	rawVideoKind        = "RawVideo"
// 	rawVideoDirName     = "RawVideos"
// 	cutVideoKind        = "CutVideo"
// 	cutVideoDirName     = "CutVideos"
// 	rawMotionKind       = "RawMotion"
// 	rawMotionDirName    = "RawMotions"
// 	rawLocationKind     = "RawLocation"
// 	rawLocationDirName  = "RawLocations"
// 	userKind            = "User"
// 	userDirName         = "Users"
// 	directoryKind       = "Directory"
// 	createTimeFieldName = "CreateTimeMs"
// 	userIDFieldName     = "UserId"
// 	emailFieldName      = "Email"
// )

// var (
// 	rawVideoKey = datastore.Key{
// 		Kind: directoryKind,
// 		Name: rawVideoDirName,
// 	}
// 	cutVideoKey = datastore.Key{
// 		Kind: directoryKind,
// 		Name: cutVideoDirName,
// 	}
// 	rawMotionKey = datastore.Key{
// 		Kind: rawMotionKind,
// 		Name: rawMotionDirName,
// 	}
// 	rawLocationKey = datastore.Key{
// 		Kind: rawLocationKind,
// 		Name: rawLocationDirName,
// 	}
// 	userKey = datastore.Key{
// 		Kind: userKind,
// 		Name: userDirName,
// 	}
// )

// // GoogleCloudDatastoreClient implements NoSQLDatabaseInterface using the real
// // Google Cloud Datastore.
// type GoogleCloudDatastoreClient struct {
// 	client                *datastore.Client
// 	projectID             string
// 	createTimeQueryOffset time.Duration
// }

// // NewGoogleCloudDatastoreClient initializes a new Google datastore.Client with the given parameters.
// // Params:
// // 		ctx context.Context
// // 		projectID string: the project
// // Returns:
// //		*GoogleCloudDatastoreClient: the client
// // 		error
// func NewGoogleCloudDatastoreClient(ctx context.Context, projectID string, createTimeQueryOffset time.Duration) (*GoogleCloudDatastoreClient, error) {
// 	client, err := datastore.NewClient(ctx, projectID)
// 	if err != nil {
// 		return nil, senecaerror.NewCloudError(fmt.Errorf("error initializing new GoogleCloudDatastoreClient - err: %v", err))
// 	}
// 	return &GoogleCloudDatastoreClient{
// 		client:                client,
// 		projectID:             projectID,
// 		createTimeQueryOffset: createTimeQueryOffset,
// 	}, nil
// }

// // GetRawVideo gets the *st.RawVideo for the given user around the specified createTime.
// // We search datastore for videos +/-createTimeQueryOffset the specified createTime.
// func (gcdc *GoogleCloudDatastoreClient) GetRawVideo(userID string, createTime time.Time) (*st.RawVideo, error) {
// 	beginTimeQuery := createTime.Add(-gcdc.createTimeQueryOffset)
// 	endTimeQuery := createTime.Add(gcdc.createTimeQueryOffset)

// 	query := datastore.NewQuery(
// 		rawVideoKind,
// 	).Filter(
// 		fmt.Sprintf("%s%s", createTimeFieldName, ">="), util.TimeToMilliseconds(beginTimeQuery),
// 	).Filter(
// 		fmt.Sprintf("%s%s", createTimeFieldName, "<="), util.TimeToMilliseconds(endTimeQuery),
// 	).Filter(
// 		fmt.Sprintf("%s%s", userIDFieldName, "="), userID,
// 	)

// 	var rawVideoOut []*st.RawVideo

// 	_, err := gcdc.client.GetAll(context.TODO(), query, &rawVideoOut)
// 	if err != nil {
// 		return nil, senecaerror.NewCloudError(fmt.Errorf("error querying datastore for RawVideo entity for user ID %q and createTime %v - err: %v", userID, createTime, err))
// 	}

// 	if len(rawVideoOut) > 1 {
// 		return nil, senecaerror.NewBadStateError(fmt.Errorf("error querying datastore for RawVideo entity for user ID %q and createTime %v, more than one value returned", userID, createTime))
// 	}

// 	if len(rawVideoOut) < 1 {
// 		return nil, senecaerror.NewNotFoundError(fmt.Errorf("rawVideo with userID %q and createTimeMs %d was not found", userID, util.TimeToMilliseconds(createTime)))
// 	}

// 	return rawVideoOut[0], nil
// }

// // GetRawVideoByID gets the rawVideo with the given ID from the datastore.
// func (gcdc *GoogleCloudDatastoreClient) GetRawVideoByID(id string) (*st.RawVideo, error) {
// 	idInt, err := strconv.ParseInt(id, 10, 64)
// 	if err != nil {
// 		return nil, senecaerror.NewBadStateError(fmt.Errorf("error converting id to int64 - err: %v", err))
// 	}

// 	rawVideo := &st.RawVideo{}

// 	if err := gcdc.client.Get(context.TODO(), &datastore.Key{
// 		Kind:   rawVideoKind,
// 		ID:     idInt,
// 		Parent: &rawVideoKey,
// 	}, rawVideo); err != nil {
// 		return nil, senecaerror.NewCloudError(fmt.Errorf("error getting raw video by ID - err: %v", err))
// 	}
// 	return rawVideo, nil
// }

// // InsertRawVideo inserts the given *st.RawVideo into the RawVideos Directory.
// func (gcdc *GoogleCloudDatastoreClient) InsertRawVideo(rawVideo *st.RawVideo) (string, error) {
// 	key := datastore.IncompleteKey(rawVideoKind, &rawVideoKey)
// 	completeKey, err := gcdc.client.Put(context.TODO(), key, rawVideo)
// 	if err != nil {
// 		return "", senecaerror.NewCloudError(fmt.Errorf("error putting RawVideo entity for user ID %q - err: %v", rawVideo.UserId, err))
// 	}
// 	return fmt.Sprintf("%d", completeKey.ID), nil
// }

// // InsertUniqueRawVideo inserts the given *st.RawVideo into the RawVideos Directory if a
// // similar RawVideo doesn't already exist.
// func (gcdc *GoogleCloudDatastoreClient) InsertUniqueRawVideo(rawVideo *st.RawVideo) (string, error) {
// 	existingRawVideo, err := gcdc.GetRawVideo(rawVideo.UserId, util.MillisecondsToTime(rawVideo.CreateTimeMs))

// 	var nfe *senecaerror.NotFoundError
// 	if err != nil && !errors.As(err, &nfe) {
// 		return "", senecaerror.NewCloudError(fmt.Errorf("error checking if raw video already exists - err: %w", err))
// 	}

// 	if existingRawVideo != nil {
// 		return "", senecaerror.NewUserError(rawVideo.UserId, fmt.Errorf("rawVideo with CreateTimeMs %d already exists", rawVideo.CreateTimeMs), fmt.Sprintf("Video at time %v already exists.", util.MillisecondsToTime(rawVideo.CreateTimeMs)))
// 	}
// 	return gcdc.InsertRawVideo(rawVideo)
// }

// // DeleteRawVideoByID deletes the rawVideo with the given ID from the datastore.
// func (gcdc *GoogleCloudDatastoreClient) DeleteRawVideoByID(id string) error {
// 	idInt, err := strconv.ParseInt(id, 10, 64)
// 	if err != nil {
// 		return senecaerror.NewBadStateError(fmt.Errorf("error converting id to int64 - err: %v", err))
// 	}

// 	key := datastore.IDKey(rawVideoKind, idInt, &rawVideoKey)
// 	if err := gcdc.client.Delete(context.TODO(), key); err != nil {
// 		return senecaerror.NewCloudError(fmt.Errorf("error deleting raw video by key - err: %v", err))
// 	}
// 	return nil
// }

// // GetCutVideo gets the *st.CutVideo for the given user around the specified createTime.
// func (gcdc *GoogleCloudDatastoreClient) GetCutVideo(userID string, createTime time.Time) (*st.CutVideo, error) {
// 	query := datastore.NewQuery(cutVideoKind).Filter(fmt.Sprintf("%s%s", userIDFieldName, "="), userID)

// 	query = addTimeOffsetFilter(createTime, gcdc.createTimeQueryOffset, query)

// 	var cutVideoOut []*st.CutVideo

// 	_, err := gcdc.client.GetAll(context.TODO(), query, &cutVideoOut)
// 	if err != nil {
// 		return nil, senecaerror.NewCloudError(fmt.Errorf("error querying datastore for CutVideo entity for user ID %q and createTime %v - err: %v", userID, createTime, err))
// 	}

// 	if len(cutVideoOut) > 1 {
// 		return nil, senecaerror.NewBadStateError(fmt.Errorf("error querying datastore for CutVideo entity for user ID %q and createTime %v, more than one value returned", userID, createTime))
// 	}

// 	if len(cutVideoOut) < 1 {
// 		return nil, senecaerror.NewNotFoundError(fmt.Errorf("cutVideo with userID %q and createTimeMs %d was not found", userID, util.TimeToMilliseconds(createTime)))
// 	}

// 	return cutVideoOut[0], nil
// }

// // DeleteCutVideoByID deletes the raw video with the given ID from the datastore.
// func (gcdc *GoogleCloudDatastoreClient) DeleteCutVideoByID(id string) error {
// 	idInt, err := strconv.ParseInt(id, 10, 64)
// 	if err != nil {
// 		return senecaerror.NewBadStateError(fmt.Errorf("error converting id to int64 - err: %v", err))
// 	}

// 	key := datastore.IDKey(cutVideoKind, idInt, &cutVideoKey)
// 	if err := gcdc.client.Delete(context.TODO(), key); err != nil {
// 		return senecaerror.NewCloudError(fmt.Errorf("error deleting cut video by key - err: %v", err))
// 	}
// 	return nil
// }

// // InsertCutVideo inserts the given *st.CutVideo into the CutVideos directory of the datastore.
// func (gcdc *GoogleCloudDatastoreClient) InsertCutVideo(cutVideo *st.CutVideo) (string, error) {
// 	key := datastore.IncompleteKey(cutVideoKind, &cutVideoKey)
// 	completeKey, err := gcdc.client.Put(context.TODO(), key, cutVideo)
// 	if err != nil {
// 		return "", senecaerror.NewCloudError(fmt.Errorf("error putting CutVideo entity for user ID %q - err: %v", cutVideo.UserId, err))
// 	}
// 	return fmt.Sprintf("%d", completeKey.ID), nil
// }

// // InsertUniqueCutVideo inserts the given *st.CutVideo if a CutVideo with a similar creation time doesn't already exist.
// func (gcdc *GoogleCloudDatastoreClient) InsertUniqueCutVideo(cutVideo *st.CutVideo) (string, error) {
// 	existingCutVideo, err := gcdc.GetCutVideo(cutVideo.UserId, util.MillisecondsToTime(cutVideo.CreateTimeMs))
// 	var nfe *senecaerror.NotFoundError
// 	if err != nil && !errors.As(err, &nfe) {
// 		return "", fmt.Errorf("error checking if cut video already exists - err: %w", err)
// 	}

// 	if existingCutVideo != nil {
// 		return "", senecaerror.NewBadStateError(fmt.Errorf("cutVideo with CreateTimeMs %d for user %s already exists", cutVideo.CreateTimeMs, cutVideo.UserId))
// 	}
// 	return gcdc.InsertCutVideo(cutVideo)
// }

// // GetRawMotion gets the *st.RawMotion for the given user at the given timestamp.
// func (gcdc *GoogleCloudDatastoreClient) GetRawMotion(userID string, timestamp time.Time) (*st.RawMotion, error) {
// 	query := datastore.NewQuery(rawMotionKind).Filter(fmt.Sprintf("%s%s", userIDFieldName, "="), userID).Filter("TimestampMs=", util.TimeToMilliseconds(timestamp))

// 	var rawMotionOut []*st.RawMotion

// 	_, err := gcdc.client.GetAll(context.TODO(), query, &rawMotionOut)
// 	if err != nil {
// 		return nil, senecaerror.NewCloudError(fmt.Errorf("error querying datastore for RawMotion entity for user ID %q and timestamp %v - err: %v", userID, timestamp, err))
// 	}

// 	if len(rawMotionOut) > 1 {
// 		return nil, senecaerror.NewBadStateError(fmt.Errorf("error querying datastore for RawMotion entity for user ID %q and timestamp %v, more than one value returned", userID, timestamp))
// 	}

// 	if len(rawMotionOut) < 1 {
// 		return nil, senecaerror.NewNotFoundError(fmt.Errorf("st.RawMotion with userID %q and TimestampMs %d was not found", userID, util.TimeToMilliseconds(timestamp)))
// 	}

// 	return rawMotionOut[0], nil
// }

// // DeleteRawMotionByID deletes the raw motion with the given ID from the datastore.
// func (gcdc *GoogleCloudDatastoreClient) DeleteRawMotionByID(id string) error {
// 	idInt, err := strconv.ParseInt(id, 10, 64)
// 	if err != nil {
// 		return senecaerror.NewBadStateError(fmt.Errorf("error converting id to int64 - err: %v", err))
// 	}

// 	key := datastore.IDKey(rawMotionKind, idInt, &rawMotionKey)
// 	if err := gcdc.client.Delete(context.TODO(), key); err != nil {
// 		return senecaerror.NewCloudError(fmt.Errorf("error deleting raw motion by key - err: %v", err))
// 	}
// 	return nil
// }

// // InsertRawMotion inserts the given *st.RawMotion into the RawMotions directory in the datastore.
// func (gcdc *GoogleCloudDatastoreClient) InsertRawMotion(rawMotion *st.RawMotion) (string, error) {
// 	key := datastore.IncompleteKey(rawMotionKind, &rawMotionKey)
// 	completeKey, err := gcdc.client.Put(context.TODO(), key, rawMotion)
// 	if err != nil {
// 		return "", senecaerror.NewCloudError(fmt.Errorf("error putting RawMotion entity for user ID %q - err: %v", rawMotion.UserId, err))
// 	}

// 	return fmt.Sprintf("%d", completeKey.ID), nil
// }

// // InsertUniqueRawMotion inserts the given *st.RawMotion if a RawMotion with the same creation time doesn't already exist.
// func (gcdc *GoogleCloudDatastoreClient) InsertUniqueRawMotion(rawMotion *st.RawMotion) (string, error) {
// 	existingRawMotion, err := gcdc.GetRawMotion(rawMotion.UserId, util.MillisecondsToTime(rawMotion.TimestampMs))

// 	var nfe *senecaerror.NotFoundError
// 	if err != nil && !errors.As(err, &nfe) {
// 		return "", fmt.Errorf("error checking if raw motion already exists - err: %w", err)
// 	}

// 	if existingRawMotion != nil {
// 		return "", senecaerror.NewBadStateError(fmt.Errorf("rawMotion with timestamp %d for user %s already exists", rawMotion.TimestampMs, rawMotion.UserId))
// 	}

// 	return gcdc.InsertRawMotion(rawMotion)
// }

// // GetRawLocation gets the *st.RawLocation for the given user at the specified timestamp.
// func (gcdc *GoogleCloudDatastoreClient) GetRawLocation(userID string, timestamp time.Time) (*st.RawLocation, error) {
// 	query := datastore.NewQuery(rawLocationKind).Filter(fmt.Sprintf("%s %s", userIDFieldName, "="), userID).Filter("TimestampMs =", util.TimeToMilliseconds(timestamp))

// 	var rawLocationOut []*st.RawLocation

// 	_, err := gcdc.client.GetAll(context.TODO(), query, &rawLocationOut)
// 	if err != nil {
// 		return nil, senecaerror.NewCloudError(fmt.Errorf("error querying datastore for RawLocation entity for user ID %q and timestamp %v - err: %v", userID, timestamp, err))
// 	}

// 	if len(rawLocationOut) > 1 {
// 		return nil, senecaerror.NewBadStateError(fmt.Errorf("error querying datastore for RawLocation entity for user ID %q and timestamp %v, more than one value returned", userID, timestamp))
// 	}

// 	if len(rawLocationOut) < 1 {
// 		return nil, nil
// 	}

// 	return rawLocationOut[0], nil
// }

// // DeleteRawLocationByID deletes the raw location with the given ID from the datastore.
// func (gcdc *GoogleCloudDatastoreClient) DeleteRawLocationByID(id string) error {
// 	idInt, err := strconv.ParseInt(id, 10, 64)
// 	if err != nil {
// 		return senecaerror.NewBadStateError(fmt.Errorf("error converting id to int64 - err: %v", err))
// 	}

// 	key := datastore.IDKey(rawLocationKind, idInt, &rawLocationKey)
// 	if err := gcdc.client.Delete(context.TODO(), key); err != nil {
// 		return senecaerror.NewCloudError(fmt.Errorf("error deleting raw location by key - err: %v", err))
// 	}
// 	return nil
// }

// // InsertRawLocation inserts the given *st.RawLocation into the RawLocations directory.
// func (gcdc *GoogleCloudDatastoreClient) InsertRawLocation(rawLocation *st.RawLocation) (string, error) {
// 	key := datastore.IncompleteKey(rawLocationKind, &rawLocationKey)
// 	completeKey, err := gcdc.client.Put(context.TODO(), key, rawLocation)
// 	if err != nil {
// 		return "", senecaerror.NewCloudError(fmt.Errorf("error putting RawLocation entity for user ID %q - err: %v", rawLocation.UserId, err))
// 	}
// 	return fmt.Sprintf("%d", completeKey.ID), nil
// }

// // InsertUniqueRawLocation inserts the given *st.RawLocation if a RawLocation with the same creation time doesn't already exist.
// func (gcdc *GoogleCloudDatastoreClient) InsertUniqueRawLocation(rawLocation *st.RawLocation) (string, error) {
// 	existingRawLocation, err := gcdc.GetRawLocation(rawLocation.UserId, util.MillisecondsToTime(rawLocation.TimestampMs))
// 	if err != nil {
// 		return "", fmt.Errorf("error checking if RawLocation already exists - err: %w", err)
// 	}
// 	if existingRawLocation != nil {
// 		return "", senecaerror.NewBadStateError(fmt.Errorf("rawLocation with timestamp %d for user %s already exists", rawLocation.TimestampMs, rawLocation.UserId))
// 	}
// 	return gcdc.InsertRawLocation(rawLocation)
// }

// // 	ListAllUserIDs lists all user IDs in the Seneca DB.
// //	Params:
// //		pageToken string: page token to apply to request
// //		maxResults int: max results to return
// //	Returns:
// //		[]string: user IDs
// //		string: the next page token, if any, to use in subsequent requests
// //		error
// func (gcdc *GoogleCloudDatastoreClient) ListAllUserIDs(pageToken string, maxResults int) ([]string, string, error) {
// 	query := datastore.NewQuery(userKind).KeysOnly().Order("__key__")

// 	if maxResults > 0 {
// 		query = query.Limit(maxResults)
// 	}

// 	// The page token here is just the minumum userID to start from (non-inclusive).
// 	if pageToken != "" {
// 		pageTokenInt, err := strconv.ParseInt(pageToken, 10, 64)
// 		if err != nil {
// 			return nil, "", fmt.Errorf("error converting pageToken %q to int - err: %w", pageToken, err)
// 		}
// 		pageTokenKey := datastore.IDKey(userKind, pageTokenInt, nil)
// 		query = query.Filter("__key__ >", pageTokenKey)
// 	}

// 	keys, err := gcdc.client.GetAll(context.TODO(), query, nil)
// 	if err != nil {
// 		return nil, "", fmt.Errorf("error getting getting all user IDs from store - err: %w", err)
// 	}

// 	ids := []string{}
// 	for _, k := range keys {
// 		ids = append(ids, strconv.FormatInt(k.ID, 10))
// 	}

// 	nextPageToken := ""
// 	if maxResults > 0 && len(ids) >= maxResults {
// 		nextPageToken = ids[len(ids)-1]
// 	}

// 	return ids, nextPageToken, nil
// }

// // 	GetUserByID returns the user with the given ID.
// //	Params:
// //		id string
// //	Returns:
// //		*st.User
// //		error
// func (gcdc *GoogleCloudDatastoreClient) GetUserByID(id string) (*st.User, error) {
// 	idInt, err := strconv.ParseInt(id, 10, 64)
// 	if err != nil {
// 		return nil, senecaerror.NewBadStateError(fmt.Errorf("error converting id to int64 - err: %v", err))
// 	}
// 	key := datastore.IDKey(userKind, idInt, &userKey)
// 	user := &st.User{}
// 	if err := gcdc.client.Get(context.TODO(), key, user); err != nil {
// 		return nil, fmt.Errorf("error getting user %q by ID - err: %w", id, err)
// 	}
// 	return user, nil
// }

// func (gcdc *GoogleCloudDatastoreClient) GetUserByEmail(email string) (*st.User, error) {
// 	query := datastore.NewQuery(userKind).Filter(fmt.Sprintf("%s%s", emailFieldName, "="), email)
// 	var users []*st.User
// 	_, err := gcdc.client.GetAll(context.TODO(), query, &users)
// 	if err != nil {
// 		return nil, fmt.Errorf("error getting user %q by email - err: %w", email, err)
// 	}

// 	if len(users) > 1 {
// 		return nil, senecaerror.NewBadStateError(fmt.Errorf("%d users with email address %q", len(users), email))
// 	}

// 	if len(users) == 0 {
// 		return nil, senecaerror.NewNotFoundError(fmt.Errorf("user with email %q not found", email))
// 	}

// 	return users[0], nil
// }

// func (gcdc *GoogleCloudDatastoreClient) InsertUser(user *st.User) (string, error) {
// 	key := datastore.IncompleteKey(userKind, &userKey)
// 	completeKey, err := gcdc.client.Put(context.TODO(), key, user)
// 	if err != nil {
// 		return "", senecaerror.NewCloudError(fmt.Errorf("error putting User entity for user with email %q - err: %v", user.Email, err))
// 	}
// 	return fmt.Sprintf("%d", completeKey.ID), nil
// }

// func (gcdc *GoogleCloudDatastoreClient) InsertUniqueUser(user *st.User) (string, error) {
// 	if user.Email == "" {
// 		return "", fmt.Errorf("user %v does not have email set, cannot insert", user)
// 	}

// 	_, err := gcdc.GetUserByEmail(user.Email)
// 	if err == nil {
// 		return "", fmt.Errorf("user with email %q already exists", user.Email)
// 	}
// 	var nfe *senecaerror.NotFoundError
// 	if !errors.As(err, &nfe) {
// 		return "", fmt.Errorf("error getting user: %w", err)
// 	}

// 	return gcdc.InsertUser(user)
// }

// func addTimeOffsetFilter(createTime time.Time, offset time.Duration, query *datastore.Query) *datastore.Query {
// 	beginTimeQuery := createTime.Add(-offset)
// 	endTimeQuery := createTime.Add(offset)

// 	return query.Filter(
// 		fmt.Sprintf("%s%s", createTimeFieldName, ">="), util.TimeToMilliseconds(beginTimeQuery),
// 	).Filter(
// 		fmt.Sprintf("%s%s", createTimeFieldName, "<="), util.TimeToMilliseconds(endTimeQuery),
// 	)
// }
