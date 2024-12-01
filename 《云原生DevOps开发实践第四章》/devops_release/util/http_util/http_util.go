package httputil

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

func SendHttpRequest(method string, headers map[string]string, url string, body []byte) ([]byte, error) {
	ioreader := bytes.NewBuffer(body)
	req, err := http.NewRequest(method, url, ioreader)
	if err != nil {
		logrus.Errorf("创建请求错误！err:%v", err)
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		logrus.Errorf("发送请求失败！url:%s err:", url, err)
		return nil, err
	}
	respBody := resp.Body
	defer respBody.Close()
	data, err := ioutil.ReadAll(respBody)
	if err != nil {
		logrus.Error("读取响应体错误！err:", err)
		return nil, err
	}
	return data, nil
}
