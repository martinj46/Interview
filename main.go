package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handler(wr http.ResponseWriter, r *http.Request) {

	file, err := os.Open("Documents.csv")
	if err != nil {
		fmt.Println("error 3", err)
		return
	}

	defer file.Close()

	data := csv.NewReader(file)

	read, err := data.ReadAll()

	html := "<html><body><table border=1>"

	for x, v := range read {
		html += "<tr>"
		for _, r := range v {
			html += "<td>" + "<a href=\"" + r + "\"> " + filepath.Base(r) + "</a>" + "</td>"

		}
		if x != 0 { // If its the header, don't generate a delete button
			html += "<td><form action=\"/delete\" style=\"display: inline;\">" +
				"<input type=\"hidden\" name=\"delete\"  value=\"" + v[1] + "\">" +
				"<input type=\"submit\" value=\"Delete\">" +
				"</form></td>"
		}
		html += "</tr>"
	}

	html += "</table></body><br>"
	html += `<form action="/upload" method="post"  enctype="multipart/form-data">
				<input type="file" name="upload">
				<br><br>
				<input type="text" name="category" placeholder="Category">
				<br><br>
				<input type="text" name="name" placeholder="File Name">
				<br><br>
				<input type="submit">
			</form></html>`

	fmt.Fprint(wr, html)
}

func pdf(w http.ResponseWriter, r *http.Request) {
	file := strings.TrimPrefix(r.URL.Path, "/")

	path := filepath.Join(".", file)

	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, path)
}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	// save form data
	category := r.FormValue("category")
	name := r.FormValue("name")
	file, handle, err := r.FormFile("upload")

	if err != nil {
		return
	}
	defer file.Close()

	// making the file path
	path := filepath.Join("./Docs/SD/", handle.Filename)
	dest, err := os.Create(path)
	if err != nil {
		return
	}
	defer dest.Close()

	_, err = io.Copy(dest, file)
	if err != nil {
		return
	}

	csvf, err := os.OpenFile("Documents.csv", os.O_APPEND|os.O_WRONLY, 0644) // Append perm, write only perm, and r/w, r, r octal encoding
	if err != nil {
		fmt.Println("Document.csv error")
		return
	}
	defer csvf.Close()

	write := csv.NewWriter(csvf)

	newRow := []string{name, path, category}
	err = write.Write(newRow)
	if err != nil {
		return
	}
	write.Flush()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func delete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	path := r.FormValue("delete")

	err := os.Remove(path)
	if err != nil {
		return
	}

	file, err := os.Open("Documents.csv")
	if err != nil {
		return
	}
	defer file.Close()

	rec, err := csv.NewReader(file).ReadAll()
	var update [][]string
	for _, x := range rec {
		if x[1] != path {
			update = append(update, x)
		}
	}

	csvf, err := os.OpenFile("Documents.csv", os.O_WRONLY|os.O_TRUNC, 0644) // write only perm | truncate perm, and r/w, r, r octal encoding, (0666) probably would work for everyone to r/w
	if err != nil {
		fmt.Println("Document.csv error")
		return
	}
	defer csvf.Close()

	write := csv.NewWriter(csvf)
	err = write.WriteAll(update)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/Docs/", pdf)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/delete", delete)

	fmt.Println("Server started on http://localhost:4200")
	http.ListenAndServe(":4200", nil)
}
