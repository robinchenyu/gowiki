package main

import (
	"fmt"

	"github.com/russross/blackfriday"
	// md "github.com/shurcooL/github_flavored_markdown"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Link struct {
	Type int // 1: for image, 2: for h2, 3: for h3
	Name string
	Url  string
}

type Page struct {
	Title string
	Body  template.HTML
	Links []Link
}

func checkerr(e error, msg string) {
	if e != nil {
		panic(e)
	}
}

var chin = make(chan string)

func scanfiles(path string, info os.FileInfo, err error) error {
	checkerr(err, "scanfiles failed ")

	// fmt.Println("path: ", path, "fileinfo ", info.Name(), info.IsDir(), info.Size())
	if strings.HasSuffix(info.Name(), ".md") {
		unix_path := strings.Replace(path, "\\", "/", -1)
		chin <- fmt.Sprintf("[%s](%s)\n\n", unix_path, unix_path[:len(path)-3])
	}
	return nil
}

func build_index() {
	go func(ch chan string) {
		var data string
		fd, err := os.Create("index.md")
		checkerr(err, "create index.md failed! ")

		defer fd.Close()

		for {
			select {
			case data = <-ch:
				if data == "EOF" {
					break
				}
				fmt.Fprintf(fd, data)
			}
		}
	}(chin)

	err := filepath.Walk(".", scanfiles)
	checkerr(err, "walk file failed! ")
	fmt.Println("finish!")
	chin <- "EOF"
}

func (p *Page) save() error {
	filename := p.Title + ".md"
	return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func markdownRender(content []byte) []byte {
	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	extensions := 0
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	// extensions |= blackfriday.EXTENSION_AUTO_HEADER_IDS

	return blackfriday.Markdown(content, renderer, extensions)
}

func buildLinks(body string) (bodywithlink string, links []Link, err error) {
	// var links = make([]Link, 8, 8)

	re_h3 := regexp.MustCompile("<[h|H]3>([^<]+)</[h|H]3>")
	output2 := re_h3.ReplaceAllString(body, "<a name='$1'></a><h3>$1</h3>")

	for _, h3name := range re_h3.FindAllStringSubmatch(body, -1) {
		// fmt.Println("link: %q", h3name)
		links = append(links, Link{Type: 3, Name: h3name[1], Url: "#" + h3name[1]})
	}

	return output2, links, nil
}

func loadPage(title string) (*Page, error) {
	re := regexp.MustCompile("<img.+src=\"(.+png)\"")
	filename := title + ".md"
	// img_file := false
	// if strings.HasSuffix(title, ".png") {
	// 	filename = title
	// 	img_file = true
	// }
	path_list := strings.Split(title, "/")
	dir1 := "/static"
	if len(path_list) >= 2 {
		dir1 = "/static/" + path_list[len(path_list)-2]
	}
	// fmt.Print(dir1)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	output := markdownRender(body)

	output1 := re.ReplaceAllString(string(output), "<img alt='' src='"+dir1+"/$1' width='50' ondblclick='enlargeImage1(this);' onclick='dropImage1(this);' ")

	output2, links, err := buildLinks(output1)

	body = []byte(output2)
	output = body

	return &Page{Title: title, Body: template.HTML(output), Links: links}, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p, err := loadPage(title)
	if err != nil {
		fmt.Println("loadPage ", err)
	}
	t, err := template.ParseFiles("static/view.html")
	if err != nil {
		fmt.Println("ParseFiles ", err)
	}
	err = t.Execute(w, p)
	if err != nil {
		fmt.Println("ExecuteTemplate ", err)
	}
}

func main() {
	// p1 := &Page{Title: "TestPage", Body: template.HTML("This is a sample Page.")}
	// p1.save()
	// p2, err := loadPage("TestPage")
	// if err != nil {
	// 	fmt.Println("loadPage ", err)
	// }
	// fmt.Println(string(p2.Body))

	go build_index()

	// 处理图片等静态页面
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	// 默认首页，无用
	http.HandleFunc("/", handler)

	// markdown页面
	http.HandleFunc("/view/", viewHandler)
	http.ListenAndServe(":8080", nil)
}
