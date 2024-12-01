package myyaml

import (
	"fmt"
	"strings"

	"devops_release/util/model"
	"github.com/sirupsen/logrus"
	y "gopkg.in/yaml.v3"
)

const (
	Newline = "\n"
	Tabs    = "\t"
	Space   = " "
	Colon   = ":"
	Hyphen  = "-"
)

type MyYamy struct {
	Key    string
	Childs []*MyYamy
	IsLeaf bool
	IsRoot bool
	Value  string
}

func NewYaml(items []model.Item) *MyYamy {
	yaml := &MyYamy{
		IsRoot: true,
	}
	for _, item := range items {
		key := item.Key
		if len(key) == 0 {
			continue
		}
		value := item.Value
		yaml.AddKV(key, value)
	}
	return yaml
}
func (yaml *MyYamy) ToString() string {
	yamlString := ""
	m := map[string]interface{}{}
	toStringHelper(yaml, &yamlString, 0)
	err := y.Unmarshal([]byte(yamlString), &m)
	if err != nil {
		fmt.Println(yamlString)
		logrus.Errorf("yaml格式错误！err:%v", err)
		return ""
	}
	return yamlString
}
func toStringHelper(node *MyYamy, yamlString *string, tabNum int) {
	if node.IsLeaf {
		return
	}
	if !node.IsRoot {
		*yamlString += Newline
	}
	for i := 0; i < tabNum; i++ {
		*yamlString = *yamlString + Space + Space
	}
	for n, c := range node.Childs {
		if strings.Index(c.Key, "]") != -1 {
			*yamlString = *yamlString + Hyphen
		} else {
			*yamlString = *yamlString + c.Key
			*yamlString = *yamlString + Colon
		}
		if c.IsLeaf {
			*yamlString = *yamlString + Space + c.Value
		}
		toStringHelper(c, yamlString, tabNum+1)
		if n < len(node.Childs)-1 {
			*yamlString = strings.TrimRight(*yamlString, Space)
			*yamlString += Newline
		}
		for i := 0; i < tabNum; i++ {
			*yamlString = *yamlString + Space + Space
		}
	}
}
func (yaml *MyYamy) AddKV(key string, value string) {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if strings.Contains(key, "[") {
		key = strings.ReplaceAll(key, "[", ".")
	}
	keyLevels := strings.Split(key, ".")
	addKVHelper(yaml, keyLevels, value, 0)
}
func addKVHelper(node *MyYamy, keylevels []string, value string, i int) {
	if i >= len(keylevels) {
		return
	}
	isleaf := false
	curVal := ""
	if len(keylevels) == i+1 {
		isleaf = true
		curVal = value
	}
	child := &MyYamy{
		Key:    keylevels[i],
		IsRoot: false,
		IsLeaf: isleaf,
		Value:  curVal,
	}
	for _, c := range node.Childs {
		if c.Key == child.Key {
			addKVHelper(c, keylevels, value, i+1)
			return
		}
	}
	node.Childs = append(node.Childs, child)
	addKVHelper(child, keylevels, value, i+1)
}
