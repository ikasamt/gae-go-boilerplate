package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"
)

type Fulltext struct {
	Ngram     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func UniqString(src []string) []string {
	ret := make([]string, 0, len(src))
	srcMap := make(map[string]struct{}, len(src))
	for _, n := range src {
		if _, ok := srcMap[n]; !ok {
			srcMap[n] = struct{}{}
			ret = append(ret, n)
		}
	}
	return ret
}

func TextNgrams(text string, n int) []string {
	sep_text := strings.Split(text, "")
	var ngrams []string
	if len(sep_text) < n {
		return nil
	}
	for i := 0; i < (len(sep_text) - n + 1); i++ {
		ngrams = append(ngrams, strings.Join(sep_text[i:i+n], ""))
	}
	return ngrams
}

func TextNgramsRange(text string) (ngrams []string) {
	ngrams = append(ngrams, TextNgrams(text, 1)...)
	ngrams = append(ngrams, TextNgrams(text, 2)...)
	ngrams = append(ngrams, TextNgrams(text, 3)...)
	return
}
func WordToSplittedWords(word string) []string {
	qSize := len(strings.Split(word, ``))

	var splits []string
	switch qSize {
	case 0:
		splits = []string{}
	case 1, 2, 3:
		splits = []string{word}
	default:
		splits = TextNgrams(word, 3)
		splits = append(splits, word[len(word)-3:len(word)])
	}

	uniq := UniqString(splits)
	return uniq
}

func FindWithSearchAPI(ctx context.Context, idx string, words []string) []int {
	values := []string{}

	for _, word := range words {
		//　文字を指定字単位で分割し配列にする ngram
		ngrams := WordToSplittedWords(word)
		for _, s := range ngrams {
			values = append(values, fmt.Sprintf(`Ngram="%s"`, s))
		}
	}
	query := strings.Join(values, ` AND `)

	searchAPIIndex, _ := search.Open(idx)
	iterator := searchAPIIndex.Search(ctx, query, &search.SearchOptions{IDsOnly: true})
	log.Debugf(ctx, `%s`, values)
	log.Debugf(ctx, `%s`, query)
	// iterator := searchAPIIndex.Search(ctx, query, &search.SearchOptions{})

	var IDs []int
	for {
		sid, err := iterator.Next(nil)
		if err == search.Done {
			break
		} else if err != nil {
			break
		}
		ID, _ := strconv.Atoi(sid)
		IDs = append(IDs, ID)
	}
	return IDs
}
