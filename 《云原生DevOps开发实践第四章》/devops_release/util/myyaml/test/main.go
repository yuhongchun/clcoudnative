package main

import (
	"fmt"
	"os"

	. "devops_release/util/model"
	"devops_release/util/myyaml"
)

func main() {
	items := []Item{
		Item{
			Key:   "first1.second1.third1",
			Value: "11111",
		},
		Item{
			Key:   "first1.second2",
			Value: "22222",
		},
		Item{
			Key:   "first2",
			Value: "333333",
		},
		Item{
			Key:   "first3.second3[0].third2",
			Value: "444444",
		},
		Item{
			Key:   "first3.second3[0].third3",
			Value: "4444445",
		},
		Item{
			Key:   "first3.second4[1].third4",
			Value: "5555555",
		},
		Item{
			Key:   "first3.second4[1].third5",
			Value: "5555555",
		},
		Item{
			Key:   "first3.second4[2]",
			Value: "5555555",
		},
	}
	yaml := myyaml.NewYaml(items)
	fmt.Println(yaml)
	fmt.Println(yaml.ToString())
	testFileName := "view.yaml"
	file, err := os.Create(testFileName)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	file.WriteString(yaml.ToString())
}
