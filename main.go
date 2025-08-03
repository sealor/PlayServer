package main

import (
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"sync"
)

var (
	tmpl       = template.Must(template.ParseFiles("templates/index.html"))
	currentCmd *exec.Cmd
	cmdMutex   sync.Mutex
)

func main() {
	http.HandleFunc("/", handleForm)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server läuft auf http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		url := r.FormValue("url")
		if url != "" {
			go playWithMPV(url)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl.Execute(w, nil)
}

func playWithMPV(url string) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	// Beende aktuellen mpv-Prozess, falls aktiv
	if currentCmd != nil && currentCmd.Process != nil {
		_ = currentCmd.Process.Kill()
		currentCmd = nil
	}

	// Starte neuen mpv-Prozess
	cmd := exec.Command("mpv", "--fs", url)
	err := cmd.Start()
	if err != nil {
		log.Printf("Fehler beim Starten von mpv: %v", err)
		return
	}
	currentCmd = cmd

	// Warte auf Beendigung
	err = cmd.Wait()
	if err != nil {
		log.Printf("mpv wurde beendet: %v", err)
	}
}
