package static

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
)

type handler struct{}

var Handler handler

type StaticContent struct {
	Hash        bool
	FileName    string
	ContentType string
	Body        []byte
}

type PrefixedStaticContent struct {
	Path    string
	Content *StaticContent
}

func (c PrefixedStaticContent) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != c.Path {
		http.NotFound(w, req)
	} else {
		w.Header().Add("Content-Type", c.Content.ContentType)
		w.Write(c.Content.Body)
	}
}

var contents = []*StaticContent{
	&sortable_min_js,
	&vote_js,
	&api_js,
	&vote_css,
}

var contents_once sync.Once

func initContents() {
	contents_once.Do(func() {
		for _, c := range contents {
			if c.Hash {
				csum := sha1.Sum(c.Body)
				version := hex.EncodeToString(csum[:])
				c.FileName = strings.Replace(c.FileName, "##", version, 1)
			}
		}
	})
}

func BindServeMux(mux *http.ServeMux, prefix string) {
	initContents()
	for _, c := range contents {
		path := prefix + "/" + c.FileName
		mux.Handle(path, PrefixedStaticContent{
			Path:    path,
			Content: c,
		})
	}
	mux.HandleFunc(prefix+"/", IndexForPrefix(prefix))
}

func PathSortableJS(prefix string) string {
	initContents()
	return prefix + "/" + sortable_min_js.FileName
}

func PathVoteJS(prefix string) string {
	initContents()
	return prefix + "/" + vote_js.FileName
}

func PathApiJS(prefix string) string {
	initContents()
	return prefix + "/" + api_js.FileName
}

func PathVoteCSS(prefix string) string {
	initContents()
	return prefix + "/" + vote_css.FileName
}
