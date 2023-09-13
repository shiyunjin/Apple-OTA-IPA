package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"regexp"
)

//go:embed static/*
var content embed.FS

func main() {
	port := "5567"

	indexT, err := template.ParseFS(content, "static/index.html")
	if err != nil {
		log.Printf("internal index template error: %v", err)
		return
	}

	plistT, err := template.ParseFS(content, "static/ipa.plist")
	if err != nil {
		log.Printf("internal plist template error: %v", err)
		return
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "text/html; charset=UTF-8")

		domain := request.Host
		var (
			name      string = request.URL.Query().Get("name")
			image     string = request.URL.Query().Get("image")
			bundle_id string = request.URL.Query().Get("bundle_id")
			ipa_file  string = request.URL.Query().Get("file")
			version   string = request.URL.Query().Get("version")
			size      string = request.URL.Query().Get("size")
		)

		if name == "" || image == "" || bundle_id == "" || ipa_file == "" || version == "" {
			writer.WriteHeader(404)

			_, _ = writer.Write([]byte("Example: https://example.com/?name=App&image=https://example.com/image.png&bundle_id=com.example.app&file=filename.ipa&version=1.0.0&size=100"))
			return
		}

		if size == "" {
			size = "未知"
		}

		writer.WriteHeader(200)

		if err := indexT.Execute(writer, struct {
			DOMAIN    string
			NAME      string
			IMAGE     string
			BUNDLE_ID string
			IPA_FILE  string
			VERSION   string
			SIZE      string
		}{
			DOMAIN:    domain,
			NAME:      name,
			IMAGE:     image,
			BUNDLE_ID: bundle_id,
			IPA_FILE:  ipa_file,
			VERSION:   version,
			SIZE:      size,
		}); err != nil {
			log.Printf("failed to execute template: %v", err)
		}
	})

	staticDir, err := fs.Sub(content, "static")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/assets/", http.FileServer(http.FS(staticDir)))
	http.Handle("/file/", http.StripPrefix("/file/", http.FileServer(http.Dir("./"))))

	http.HandleFunc("/ipa/", func(writer http.ResponseWriter, request *http.Request) {
		// regex param form url
		// https://{{.DOMAIN}}/ipa/{{.NAME}}/{{.BUNDLE_ID}}-{{.VERSION}}-{{.IPA_FILE}}/install.plist
		urlMatch := regexp.MustCompile(`^/ipa/([^/]+)/([^/]+)-([^/]+)-([^/]+)/install.plist$`)
		params := urlMatch.FindStringSubmatch(request.URL.Path)

		if len(params) != 5 {
			writer.WriteHeader(404)
			return
		}

		name := params[1]
		bundle_id := params[2]
		version := params[3]
		ipa_file := params[4]

		writer.Header().Add("Content-Type", "application/x-plist")
		writer.WriteHeader(200)
		domain := request.Host

		if err := plistT.Execute(writer, struct {
			DOMAIN    string
			NAME      string
			BUNDLE_ID string
			VERSION   string
			IPA_FILE  string
		}{
			DOMAIN:    domain,
			NAME:      name,
			BUNDLE_ID: bundle_id,
			VERSION:   version,
			IPA_FILE:  ipa_file,
		}); err != nil {
			log.Printf("failed to execute template: %v", err)
		}
	})

	log.Printf("Serving on HTTP port: %s\n", port)
	_ = http.ListenAndServe(":"+port, nil)
}
