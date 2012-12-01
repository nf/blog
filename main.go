package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"code.google.com/p/go.talks/pkg/present"
)

type Doc struct {
	*present.Doc
	Path string
}

var (
	docs      []*Doc
	docPaths  = make(map[string]*Doc)
	httpAddr  = flag.String("http", ":8080", "HTTP listen address")
	rTemplate = template.Must(template.ParseFiles("root.tmpl"))
	pTemplate = template.Must(present.Template().ParseFiles("action.tmpl", "article.tmpl"))
)

func main() {
	flag.Parse()
	var err error
	docs, err = loadDocs()
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range docs {
		d.Template = pTemplate
		docPaths[d.Path] = d
	}
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.URL.Path == "/" {
		err = rTemplate.Execute(w, docs)
	} else {
		doc, ok := docPaths[r.URL.Path[1:]]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		err = pTemplate.ExecuteTemplate(w, "root", doc)
	}
	if err != nil {
		log.Println(err)
		fmt.Fprintln(w, err)
	}
}

func loadDocs() ([]*Doc, error) {
	var docs []*Doc
	fn := func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".article" {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		d, err := present.Parse(f, filepath.Base(path), 0)
		if err != nil {
			return err
		}
		docs = append(docs, &Doc{d, path})
		return nil
	}
	err := filepath.Walk(".", fn)
	if err != nil {
		return nil, err
	}
	sort.Sort(docsByTime(docs))
	return docs, nil
}

type docsByTime []*Doc

func (s docsByTime) Len() int           { return len(s) }
func (s docsByTime) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s docsByTime) Less(i, j int) bool { return s[i].Time.After(s[j].Time) }
