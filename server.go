package amazing

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

func PreviewHandler(repo, example string) (http.Handler, error) {
	if !member(example, "customer-followup", "project-workbench") {
		return nil, fmt.Errorf("Unknown example")
	}
	files := map[string]string{"/": "index.html", "/index.html": "index.html", "/styles.css": "styles.css", "/app.js": "app.js"}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, ok := files[r.URL.Path]
		if !ok || r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		b, e := os.ReadFile(filepath.Join(repo, "examples", example, name))
		if e != nil {
			http.NotFound(w, r)
			return
		}
		mime := "text/html; charset=utf-8"
		if name == "styles.css" {
			mime = "text/css"
		}
		if name == "app.js" {
			mime = "text/javascript"
		}
		w.Header().Set("Content-Type", mime)
		_, _ = w.Write(b)
	}), nil
}
func Serve(repo, example, port string, out io.Writer) error {
	h, e := PreviewHandler(repo, example)
	if e != nil {
		return e
	}
	l, e := net.Listen("tcp", net.JoinHostPort("127.0.0.1", port))
	if e != nil {
		return e
	}
	defer l.Close()
	fmt.Fprintln(out, "Preview: http://"+l.Addr().String())
	return http.Serve(l, h)
}
