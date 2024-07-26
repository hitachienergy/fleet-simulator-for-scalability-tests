package hawkbit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hitachienergy/scalability-test-client/examples/httppool"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
)

type DDIRestApi struct {
	tenant string
	id     string

	baseEndpoint string
	gatewayToken string

	usePool bool
}

// newDDIRestApi creates a new instance of DDIRestApi
func newDDIRestApi(id string, tenant string, baseEndpoint string, gatewayToken string, usePool bool) *DDIRestApi {
	r := DDIRestApi{
		id:           id,
		tenant:       tenant,
		baseEndpoint: baseEndpoint,
		gatewayToken: gatewayToken,
		usePool:      usePool,
	}

	return &r
}

// getActionId extracts the action id from the link
func getActionId(link string) (id int64, err error) {
	startIndex := strings.LastIndex(link, "/") + 1
	endIndex := strings.Index(link, "?")

	id, err = strconv.ParseInt(link[startIndex:endIndex], 10, 64)
	if err != nil {
		return -1, err
	}

	return id, nil
}

// getRequiredLink polls the server for the newest information and extracts the target link from the response
func (r *DDIRestApi) getRequiredLink(basekey string) (link string, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/%s/controller/v1/%s", r.baseEndpoint, r.tenant, r.id), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/hal+json")
	req.Header.Add("Authorization", fmt.Sprintf("GatewayToken %s", r.gatewayToken))
	res, err := r.send(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return "", xerrors.Errorf("Fail to get %s. Status code: %d (%s)", basekey, res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var data ControllerBase
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	return data.Links[basekey].Href, nil
}

// getActionWithDeployment gets a new deployment information by sending the related messages to the server
func (r *DDIRestApi) getActionWithDeployment(actionID int64) (deployment *Deployment, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/%s/controller/v1/%s/deploymentBase/%d", r.baseEndpoint, r.tenant, r.id, actionID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/hal+json")
	req.Header.Add("Authorization", fmt.Sprintf("GatewayToken %s", r.gatewayToken))
	res, err := r.send(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return nil, xerrors.Errorf("Fail to get action with deployment. Status code: %d (%s)", res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var data ActionWithDeployment
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data.Deployment, nil
}

// sendConfirmation sends a confirmation message to the server
func (r *DDIRestApi) sendConfirmation(actionID int64, status string) (err error) {
	payload := map[string]interface{}{
		"confirmation": status,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://%s/%s/controller/v1/%s/confirmationBase/%d/feedback", r.baseEndpoint, r.tenant, r.id, actionID),
		bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	req.Header.Add("Accept", "application/hal+json")
	req.Header.Add("Authorization", fmt.Sprintf("GatewayToken %s", r.gatewayToken))
	res, err := r.send(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return xerrors.Errorf("Fail to post confirmation feedback. Status code: %d (%s)", res.StatusCode, res.Status)
	}

	return nil
}

// reportUpdate reports update status to the server
func (r *DDIRestApi) reportUpdate(actionID int64, localStatus LocalUpdateStatus) (err error) {
	status := DDIUpdateStatus{
		Details: localStatus.StatusMsgs,
	}

	switch localStatus.Status {
	case SUCCESSFUL:
		status.Execution = "closed"
		status.Result.Finished = "success"
		status.Code = 200
	case ERROR:
		status.Execution = "closed"
		status.Result.Finished = "failure"
	case DOWNLOADING:
		status.Execution = "download"
		status.Result.Finished = "none"
	case DOWNLOADED:
		status.Execution = "downloaded"
		status.Result.Finished = "none"
	case RUNNING:
		status.Execution = "proceeding"
		status.Result.Finished = "none"
	}

	return r.sendUpdateFeedback(actionID, status)
}

// sendUpdateFeedback sends an update feedback message to the server
func (r *DDIRestApi) sendUpdateFeedback(actionID int64, status DDIUpdateStatus) (err error) {
	payload := DDIUpdateFeedback{
		Status: status,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://%s/%s/controller/v1/%s/deploymentBase/%d/feedback", r.baseEndpoint, r.tenant, r.id, actionID),
		bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	req.Header.Add("Accept", "application/hal+json")
	req.Header.Add("Authorization", fmt.Sprintf("GatewayToken %s", r.gatewayToken))

	res, err := r.send(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return xerrors.Errorf("Fail to post update feedback. Status code: %d (%s)", res.StatusCode, res.Status)
	}

	return nil
}

// sendHTTP sends a simple http request. It can either send directly or using HTTP Pool
func (r *DDIRestApi) send(request *http.Request) (*http.Response, error) {
	request.Close = true

	if !r.usePool {
		return httppool.Client.Do(request)
	}
	client := httppool.Pool.Get()
	defer httppool.Pool.Put(client)
	return client.Do(request)
}
