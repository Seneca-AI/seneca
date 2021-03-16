package rawvideohandler

import (
	"fmt"
	"net/http"
)

// HandleRawVideoPostRequest accepts POST requests that include mp4 data and
// parses the metadata to gather timestamp info and, if possible, location info.
// The video itself is stored in simple storage and the metadata is stored in
// firestore.  If the video does not contain timestamp info, the server
// returns a 400 error.
func HandleRawVideoPostRequest(w http.ResponseWriter, r *http.Request) {
	video_name, err := getVideoName(r)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		return
	}
	w.WriteHeader(200)
	fmt.Fprintf(w, "The video's name is %q", video_name)
}

func getVideoName(r *http.Request) (string, error) {
	video_name := r.FormValue("video_name")
	if video_name == "" {
		return "", fmt.Errorf("'video_name' not found in request's form data")
	}
	return video_name, nil
}
