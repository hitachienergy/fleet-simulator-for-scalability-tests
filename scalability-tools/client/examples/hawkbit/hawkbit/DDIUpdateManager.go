package hawkbit

import (
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/xerrors"
)

/** Update Manager **/

type DDIUpdateManager struct {
	// UpdateManagerDDI

	*sync.Mutex
	CurrentActionID int64
	IsUpdating      bool
	// status          LocalUpdateStatus

	*DDIClient
}

// newDDIUpdateManager creates a new instance of DDI update manager
func newDDIUpdateManager(client *DDIClient) *DDIUpdateManager {
	return &DDIUpdateManager{
		// status:          LocalUpdateStatus{status: IDLE},
		CurrentActionID: -1,
		DDIClient:       client,
		IsUpdating:      false,
		Mutex:           &sync.Mutex{},
	}
}

// func (u *DDIUpdateManager) PrepareUpdate(actionID int64) (deployment *Deployment, err error) {
// 	status := atomic.LoadInt32(&u.isUpdating)
// 	if status != -1 {
// 		return nil, nil
// 	}

// 	deployment, err = u.GetActionWithDeployment(actionID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return deployment, nil
// }

// resetUpdate switch the firmware to the new version and clear the intermediate update states
func (u *DDIUpdateManager) resetUpdate(id int64) bool {
	u.Lock()
	defer u.Unlock()

	u.IsUpdating = false
	if id >= 0 {
		u.CurrentActionID = id
	}

	return true
}

// tryState checks if there is current an update happening and tries the lock
func (u *DDIUpdateManager) tryState(id int64) bool {
	u.Lock()
	defer u.Unlock()

	if u.IsUpdating || u.CurrentActionID == id {
		return false
	}
	u.IsUpdating = true
	return true
}

// startUpdate launches an update simulation
func (u *DDIUpdateManager) startUpdate(actionID int64, deployment *Deployment) (err error) {
	success := u.tryState(actionID)
	if !success {
		return nil
	}

	defer func() {
		// TODO-Option: only set actionID for SUCCESS or ERROR
		u.resetUpdate(actionID)
		u.controller.CompleteTask(err == nil)
	}()

	err = u.reportUpdate(actionID, LocalUpdateStatus{RUNNING, []string{"Simulation begins!"}})
	if err != nil {
		return err
	}

	err = u.simulateDownload(deployment.Chunks, actionID)
	if err != nil {
		return err
	}

	if deployment.Update != "skip" {
		err = u.reportUpdate(actionID, LocalUpdateStatus{SUCCESSFUL, []string{"Simulation complete!"}})
		return err
	}

	return nil
}

// simulateDownload downloads firmware from remote platform
func (u *DDIUpdateManager) simulateDownload(modules []Chunk, actionID int64) (err error) {
	messages := []string{}
	for _, chunk := range modules {
		for _, artifact := range chunk.Artifacts {
			messages = append(messages,
				fmt.Sprintf("Download starts for: %s with SHA1 hash %s and size %d",
					artifact.Filename, artifact.Hashes["sha1"], artifact.Size))
		}
	}
	err = u.reportUpdate(actionID, LocalUpdateStatus{DOWNLOADING, messages})
	if err != nil {
		return err
	}

	// log.Info().Msgf("[%s] Start downloading", u.id)

	result := LocalUpdateStatus{Status: DOWNLOADED}
	for _, chunk := range modules {
		for _, artifact := range chunk.Artifacts {
			status := u.handleArtifact(&artifact)
			result.StatusMsgs = append(result.StatusMsgs, status.StatusMsgs...)
			if status.Status == ERROR {
				err = xerrors.Errorf(result.StatusMsgs[0])
				result.Status = ERROR
			}
		}
	}
	reportErr := u.reportUpdate(actionID, LocalUpdateStatus{result.Status, result.StatusMsgs})

	// log.Info().Msgf("[%s] Finish downloading", u.id)

	if reportErr != nil {
		err = reportErr
	}

	return err
}

// handleArtifact handles the artifact including download and verification
func (u *DDIUpdateManager) handleArtifact(artifact *DDIArtifact) (status LocalUpdateStatus) {
	if url, ok := artifact.Links["download"]; ok {
		return u.downloadUrl(url.Href, artifact.Hashes["sha1"], artifact.Size)
	} else if url, ok := artifact.Links["download-http"]; ok {
		return u.downloadUrl(url.Href, artifact.Hashes["sha1"], artifact.Size)
	} else {
		return LocalUpdateStatus{Status: ERROR,
			StatusMsgs: []string{fmt.Sprintf("No supported url for artifact %s (Expected: HTTPS, HTTP)", artifact.Filename)}}
	}
}

// downloadUrl downloads a file and do verification
func (u *DDIUpdateManager) downloadUrl(url string, hash string, size int64) (status LocalUpdateStatus) {
	data, err := u.download(url)
	if err != nil {
		return LocalUpdateStatus{Status: ERROR, StatusMsgs: []string{fmt.Sprintf("Failed to download %s: %s", url, err)}}
	}

	contentSize := int64(len(data))
	if contentSize != size {
		return LocalUpdateStatus{Status: ERROR,
			StatusMsgs: []string{fmt.Sprintf("Download %s has wrong content length (Expected: %d, Got %d)", url, size, contentSize)}}
	}

	myhash := sha1.New()
	myhash.Write(data)
	hashString := strings.ToLower(hex.EncodeToString(myhash.Sum(nil)))
	if hashString != hash {
		return LocalUpdateStatus{Status: ERROR,
			StatusMsgs: []string{fmt.Sprintf("Download %s failed with SHA1 hash missmatch (Expected: %s, Got %s)", url, hash, hashString)}}
	}

	return LocalUpdateStatus{Status: SUCCESSFUL, StatusMsgs: []string{fmt.Sprintf("Download %s successfully (%d bytes)", url, contentSize)}}
}

// download does the real file downloading
func (u *DDIUpdateManager) download(url string) (body []byte, err error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("GatewayToken %s", u.gatewayToken))
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return nil, xerrors.Errorf("Fail to get action with deployment. Status code: %d (%s)", res.StatusCode, res.Status)
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
