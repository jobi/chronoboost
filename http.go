package chronoboost

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const loginPath = "/reg/login"
const getKeyPath = "/reg/getkey/"
const getKeyFramePath = "/keyframe"

func (state *CurrentState) ObtainAuthCookie() error {
	loginUrl := &url.URL{"http", "", nil, state.AuthHost, loginPath, "", ""}
	resp, err := http.PostForm(loginUrl.String(), url.Values{"email": {state.Email}, "password": {state.Password}})
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("Invalid HTTP response %d", resp.StatusCode))
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "USER" {
			state.Token = cookie.Value
			break
		}
	}
	return nil
}

func (state *CurrentState) ObtainDecryptionKey() {
	path := fmt.Sprintf("%s%d.asp", getKeyPath, state.EventState.Event.Number)
	query := fmt.Sprintf("auth=%s", state.Token)
	getKeyUrl := &url.URL{"http", "", nil, state.AuthHost, path, query, ""}
	resp, err := http.Get(getKeyUrl.String())
	if err != nil {
		fmt.Println("Failed to get decryption key", err)
		return
	}

	if resp.StatusCode >= 400 {
		fmt.Println("Error code retrieving decryption key", resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read decryption key body", err)
		return
	}

	key := string(bytes)
	if key == "INVALID" {
		fmt.Println("Failed to retrieve key")
		return
	}

	var decoded []byte
	decoded, err = hex.DecodeString(strings.ToLower(string(bytes)))

	if err != nil {
		fmt.Println("Error decoding key", err)
		return
	}

	state.CryptoState.Key = uint(decoded[0])<<24 | uint(decoded[1])<<16 | uint(decoded[2])<<8 | uint(decoded[3])
	fmt.Println("Successfully decoded decryption key", state.CryptoState.Key)
}

func (state *CurrentState) ObtainTotalLaps() {
}

func (state *CurrentState) ObtainKeyFrame(number KeyFrameNumber) ([]byte, error) {
	var path string

	if number == 0 {
		path = fmt.Sprintf("%s.bin", getKeyFramePath)
	} else {
		path = fmt.Sprintf("%s_%05d.bin", getKeyFramePath, number)
	}

	getKeyFrameUrl := &url.URL{"http", "", nil, state.Host, path, "", ""}
	resp, err := http.Get(getKeyFrameUrl.String())
	if err != nil {
		fmt.Println("Failed to get key frame", err)
		return nil, err
	}

	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	bytes, err := ioutil.ReadAll(reader)

	if err != nil {
		fmt.Println("Error reading keyframe", err)
		return nil, err
	}

	return bytes, nil
}
