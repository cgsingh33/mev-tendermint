package p2p

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

type ResultRegisterNodeApi struct {
	PeerConnString string `json:"peer_conn_string"`
	Code           uint32 `json:"code"`
}

func RegisterWithSentinel(logger log.Logger, APIKey, peerID, sentinel string) {
	logger.Info(
		"[p2p.sentinel]: Registering with sentinel (first try)",
		"API Key", APIKey,
		"peerID", peerID,
		"sentinel string", sentinel,
	)

	jsonData, err := makePostRequestData(peerID, APIKey)
	if err != nil {
		logger.Info("[p2p.sentinel]: Err marshalling json data:", err)
		return
	}

	go postRequestRoutine(logger, sentinel, jsonData)
}

func attemptRegisterOnce(logger log.Logger, sentinel string, jsonData []byte) error {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Post(sentinel, "application/json", bytes.NewBuffer(jsonData)) //nolint:gosec
	if err != nil {
		logger.Info("[p2p.sentinel]: Err registering with sentinel", "err", err.Error())
		return err
	}
	if resp == nil {
		logger.Info("[p2p.sentinel]: No response from sentinel", "err", err.Error())
		return errors.New("no response from sentinel")
	}
	if resp.StatusCode != http.StatusOK {
		logger.Info("[p2p.sentinel]: Bad status code from sentinel", "status code", resp.StatusCode)
		return errors.New("bad status code from sentinel")
	}
	if resp.Body == nil {
		logger.Info("[p2p.sentinel]: No body in response from sentinel", "err", err.Error())

		return errors.New("no body in response from sentinel")
	}
	defer resp.Body.Close()

	unmarshalledResponse := &types.RPCResponse{}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Info("[p2p.sentinel]: error reading body", "err", err)
		return errors.New("error reading response body")
	}
	err = json.Unmarshal(bodyBytes, unmarshalledResponse)
	if err != nil {
		logger.Info("[p2p.sentinel]: error unmarshalling response body", "err", err)
		return errors.New("error unmarshalling response body")
	}
	if unmarshalledResponse.Error != nil {
		logger.Info("[p2p.sentinel]: error from sentinel rpc", "err", unmarshalledResponse.Error)
		return errors.New("error in response body")
	}
	fmt.Println("unmarshalledResponse")
	fmt.Println(unmarshalledResponse.Result)
	result := &ResultRegisterNodeApi{}
	err = json.Unmarshal(bodyBytes, result)
	if err != nil {
		logger.Info("[p2p.sentinel]: error unmarshalling response body", "err", err)
	} else {
		logger.Info("[p2p.sentinel]: successfully unmarshalled response body")
		fmt.Println(result)
	}
	return nil
}

func postRequestRoutine(logger log.Logger, sentinel string, jsonData []byte) {
	tries := 1
	for {
		logger.Info("[p2p.sentinel]: Attempt to reregister via Sentinel API",
			"try #", tries,
		)
		err := attemptRegisterOnce(logger, sentinel, jsonData)
		if err == nil {
			logger.Info("[p2p.sentinel]: Successfully registered with Sentinel API")
			return
		}
		logger.Info("[p2p.sentinel]: Failed to register with Sentinel API", "err", err)
		time.Sleep(30 * time.Second)
		tries++
	}
}

func makePostRequestData(peerID, APIKey string) ([]byte, error) {
	params := [2]string{peerID, APIKey}
	data := map[string]interface{}{
		"method": "register_node_api",
		"params": params,
		"id":     1,
	}

	return json.Marshal(data)
}
