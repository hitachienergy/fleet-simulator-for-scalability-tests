package hawkbit

var ConfirmationBase = "confirmationBase"
var DeploymentBase = "deploymentBase"

type Link struct {
	Href string `json:"href"`
}

type ControllerBase struct {
	Links map[string]Link `json:"_links"`
}

type ActionWithDeployment struct {
	ID         string     `json:"id"`
	Deployment Deployment `json:"deployment"`
}

type Deployment struct {
	Download string  `json:"download"`
	Update   string  `json:"update"`
	Chunks   []Chunk `json:"chunks"`
}

type Chunk struct {
	Part      string        `json:"part"`
	Version   string        `json:"version"`
	Name      string        `json:"name"`
	Artifacts []DDIArtifact `json:"artifacts"`
}

type DDIArtifact struct {
	Filename string            `json:"filename"`
	Hashes   map[string]string `json:"hashes"`
	Size     int64             `json:"size"`
	Links    map[string]Link   `json:"_links"`
}

type DDIUpdateFeedback struct {
	Time   string          `json:"time"`
	Status DDIUpdateStatus `json:"status"`
}

type DDIUpdateStatus struct {
	Execution string          `json:"execution"`
	Result    DDIUpdateResult `json:"result"`
	Code      int32           `json:"code"`
	Details   []string        `json:"details"`
}

type DDIUpdateResult struct {
	Finished string `json:"finished"`
}
