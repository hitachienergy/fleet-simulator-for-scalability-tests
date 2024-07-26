package thingsboard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hitachienergy/scalability-test-client/examples/httppool"
	"io"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/xerrors"
)

type HTTPService struct {
	baseEndpoint   string
	accessToken    string
	secureDownload bool

	useHTTPPool bool
}

// newHTTPService creates a new HTTPService
func newHTTPService(baseEndpoint string, accessToken string, useHTTPPool bool) *HTTPService {
	_, port, err := net.SplitHostPort(baseEndpoint)
	secureDownload := (err == nil && port == "443")
	s := &HTTPService{
		baseEndpoint:   baseEndpoint,
		accessToken:    accessToken,
		secureDownload: secureDownload,
		useHTTPPool:    useHTTPPool,
	}

	return s
}

// getID returns the access token of the device
// func (s *HTTPService) getID() string {
// 	return s.accessToken
// }

// getFirmwareInfo returns the firmware info pulled from the server
func (s *HTTPService) getFirmwareInfo() (info *FirmwareInfo, err error) {
	myurl, err := url.Parse(fmt.Sprintf("http://%s/api/v1/%s/attributes", s.baseEndpoint, s.accessToken))
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("sharedKeys", FIRMWARE_SHARED_KEYS)
	myurl.RawQuery = params.Encode()
	req, err := http.NewRequest("GET", myurl.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := s.send(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return nil, xerrors.Errorf("Fail to get firmware info. Status code: %d (%s)", res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var data HTTPAttributes
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	if len(data.Shared.Version) != 0 && len(data.Shared.Checksum) == 0 {
		fmt.Println(s.accessToken)
		fmt.Println(data.Shared)
	}
	return &data.Shared, nil
}

// getFirmware downloads the firmware from the server
func (s *HTTPService) getFirmware(fwinfo FirmwareInfo) (data []byte, err error) {
	httpsIndicator := ""
	if s.secureDownload {
		httpsIndicator = "s"
	}
	myurl, err := url.Parse(fmt.Sprintf("http%s://%s/api/v1/%s/firmware", httpsIndicator, s.baseEndpoint, s.accessToken))
	if err != nil {
		return nil, err
	}

	// TODO: do we need to support chunk-by-chunk downloading?
	params := url.Values{}
	params.Add("title", fwinfo.Title)
	params.Add("version", fwinfo.Version)
	myurl.RawQuery = params.Encode()
	req, err := http.NewRequest("GET", myurl.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := s.send(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return nil, xerrors.Errorf("Fail to get firmware. Status code: %d (%s)", res.StatusCode, res.Status)
	}

	data, err = io.ReadAll(res.Body)
	return data, err
}

// func (s *HTTPService) getFirmware(fwinfo FirmwareInfo) (data []byte, err error) {
// 	httpsIndicator := ""
// 	if s.secureDownload {
// 		httpsIndicator = "s"
// 	}

// 	if CHUNK_SIZE <= 0 {
// 		CHUNK_SIZE = int(fwinfo.Size)
// 	}

// 	num_chunks := int(math.Ceil(fwinfo.Size / float64(CHUNK_SIZE)))
// 	for i := 0; i < num_chunks; i++ {
// 		// log.Info().Msgf("Loading chunk %d / %d", i+1, num_chunks)

// 		myurl, err := url.Parse(fmt.Sprintf("http%s://%s/api/v1/%s/firmware", httpsIndicator, s.baseEndpoint, s.accessToken))
// 		if err != nil {
// 			return nil, err
// 		}
// 		params := url.Values{}
// 		params.Add("title", fwinfo.Title)
// 		params.Add("version", fwinfo.Version)
// 		params.Add("chunk", fmt.Sprintf("%d", i))
// 		params.Add("size", fmt.Sprintf("%d", CHUNK_SIZE))

// 		myurl.RawQuery = params.Encode()

// 		req, err := http.NewRequest("GET", myurl.String(), nil)
// 		if err != nil {
// 			return nil, err
// 		}

// 		res, err := s.send(req)
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer res.Body.Close()
// 		if res.StatusCode > 299 {
// 			return nil, xerrors.Errorf("Fail to get firmware. Status code: %d (%s)", res.StatusCode, res.Status)
// 		}

// 		data_chunk, err := io.ReadAll(res.Body)
// 		if err != nil {
// 			return nil, err
// 		}
// 		data = append(data, data_chunk...)
// 		// log.Info().Msgf("Loaded chunk %d / %d", i+1, num_chunks)
// 	}
// 	return data, err
// }

// reportUpdateState reports the update state the the server
func (s *HTTPService) reportUpdateState(state FWUpdateState) (err error) {
	jsonPayload, err := json.Marshal(state)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://%s/api/v1/%s/telemetry", s.baseEndpoint, s.accessToken),
		bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")

	res, err := s.send(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return xerrors.Errorf("Fail to post update feedback. Status code: %d (%s)", res.StatusCode, res.Status)
	}

	return nil
}

// send sends an HTTP requests using either a shared http client or the http pool based on the configuration
func (s *HTTPService) send(request *http.Request) (*http.Response, error) {
	request.Close = true

	if !s.useHTTPPool {
		return httppool.Client.Do(request)
	}
	client := httppool.Pool.Get()
	defer httppool.Pool.Put(client)
	return client.Do(request)
}
