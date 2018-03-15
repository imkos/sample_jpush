package jpush

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"jsonc"
	"net/http"
)

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func PostJSON(jc *JPushClient, uri string) ([]byte, error) {
	j_data, err := json.Marshal(jc.wrapper)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(j_data)
	req, e1 := http.NewRequest(HTTP_POST, uri, body)
	if e1 != nil {
		return nil, e1
	}
	if jc.headers != nil {
		for k, v := range jc.headers {
			req.Header.Set(k, v)
		}
	}
	client := &http.Client{}
	resp, e2 := client.Do(req)
	if e2 != nil {
		return nil, e2
	}
	defer resp.Body.Close()
	//对HTTP POST请求返回做状态码判断
	if resp.StatusCode != http.StatusOK {
		b_resp, e3 := ioutil.ReadAll(resp.Body)
		if e3 != nil {
			return nil, errors.New(resp.Status + "@" + e3.Error())
		} else {
			return nil, errors.New(resp.Status + "@" + string(b_resp))
		}
	}
	return ioutil.ReadAll(resp.Body)
}
