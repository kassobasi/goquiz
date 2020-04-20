package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type quizItem struct {
	question string
	answer   string
}

func main() {
	f := flag.String("f", "program.csv", "name of the CSV file that contains questions and answers")
	t := flag.Int("t", 30, "timer in seconds")

	flag.Parse()

	items, err := getQuizItems(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	c := context.Background()
	tc, _ := context.WithTimeout(c, time.Duration(*t)*time.Second)

	s := make(chan int, 1)
	go askQuestions(tc, items, s)

	score := <-s
	fmt.Fprintf(os.Stdout, "Your score is %d/%d\n", score, len(items))
}

func askQuestions(c context.Context, items []quizItem, s chan<- int) {
	score := 0
	r := make(chan string)
	go func() {
		for {
			var response string
			_, err := fmt.Fscanf(os.Stdin, "%s", &response)
			if err != nil {
				fmt.Fprintln(os.Stderr, "can't read that")
				r <- ""
			}
			r <- response
		}
	}()

LOOP:
	for _, item := range items {
		fmt.Fprintf(os.Stdout, "%v? ", item.question)
		select {
		case <-c.Done():
			break LOOP
		case response := <-r:
			if response == item.answer {
				score++
			}
		}
	}

	s <- score
}

func getQuizItems(name *string) ([]quizItem, error) {
	f, err := os.Open(*name)
	if err != nil {
		return nil, fmt.Errorf("can't open file: %v %v", *name, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	lines, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%v is not a valid CSV file: %v", *name, err)
	}

	var items []quizItem
	for i, l := range lines {
		if len(l) != 2 {
			return nil, fmt.Errorf("not a valid question-answer at line %d: %v", i, l)
		}
		items = append(items, quizItem{question: l[0], answer: strings.TrimSpace(l[1])})
	}

	return items, nil
}
