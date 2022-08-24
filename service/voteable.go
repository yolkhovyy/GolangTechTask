package service

type Voteable struct {
	UUID     string `dynamo:"ID,hash"`
	Question string
	Answers  []string
	Votes    []int64
}
