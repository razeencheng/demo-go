package main

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Nickname  string    `json:"nickname,omitempty"`
	Sex       int       `json:"sex,string"`
	Age       int       `json:"age,omitempty"`
	AgeStr    int       `json:"age,omitempty,string"`
	Merried   bool      `json:"merried,omitempty"`
	Ms        bool      `json:"ms,omitempty,string"`
	Relation  *Relation `json:"relation"`
}

type Relation struct {
	Mma *Person `json:"mma"`
	Son *Person `json:"son,omitempty"`
}

func main() {
	xiaoming := &Person{
		FirstName: "xiaoming",
		Nickname:  "",
		AgeStr:    18,
		Ms:        true,
		Relation:  &Relation{},
	}
	buf, err := json.MarshalIndent(xiaoming, "", "")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf))
}
