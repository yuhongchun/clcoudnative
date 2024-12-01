package apollo
import (
	"fmt"
	"gopkg.in/yaml.v3"
)

// 将apollo的K、V转化成ymal格式数据
func ApolloTransitionYaml(map[interface{}]interface{}) (yamlStr string) {
	return yamlStr
}

// 将ymal格式数据转化成apollo的K、V值
func YamlTransitionApollo(yamlStr string) (KV map[string]interface{}, err error) {
	//yaml文件内容转换成map[interface{}]interface{})
	resultMap := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(yamlStr), &resultMap); err != nil {
		return nil, err
	}
	fmt.Println("****", resultMap)
	// 遍历map ，此处取metadata.name值
	KVMap := make(map[string]interface{})
	recursionParseMap("", resultMap, KVMap)
	return KVMap, nil
}

// 将ymal格式Map转化成apollo的格式map
func recursionParseMap(prefix string, yamlMap map[string]interface{}, KVMap map[string]interface{}) {
	if yamlMap == nil || KVMap == nil {
		//入参为空则返回
		return
	}
	for key, value := range yamlMap {
		switch value := value.(type) {
		case string, int, int64, bool, float32, float64:
			if prefix != "" {
				KVMap[prefix+"."+key] = value
			} else {
				KVMap[key] = value
			}
		case []interface{}:
			for index, item := range value {
				switch item := item.(type) {
				case string, int, int64, bool, float32, float64:
					if prefix != "" {
						KVMap[prefix+"."+key+fmt.Sprintf("[%d]", index)] = fmt.Sprintf("%v", item)
					} else {
						KVMap[key+fmt.Sprintf("[%d]", index)] = fmt.Sprintf("%v", item)
					}
				case map[string]interface{}:
					if prefix != "" {
						recursionParseMap(prefix+"."+key+fmt.Sprintf("[%d]", index), item, KVMap)
					} else {
						recursionParseMap(key+fmt.Sprintf("[%d]", index), item, KVMap)
					}
				}
			}
		case map[string]interface{}:
			if prefix != "" {
				recursionParseMap(prefix+"."+key, value, KVMap)
			} else {
				recursionParseMap(key, value, KVMap)
			}

		default:
			if prefix != "" {
				KVMap[prefix+"."+key] = ""
			} else {
				KVMap[key] = ""
			}
		}
	}
}
