package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

var (
	tmpl          = template.Must(template.ParseFiles("templates/index.html"))
	currentCmd    *exec.Cmd
	cmdMutex      sync.Mutex
	currentURL    string
	ipcSocketPath = "/tmp/mpvsocket"
)

func main() {
	port := flag.String("port", "8080", "port of web server interface")
	address := flag.String("address", "0.0.0.0", "address of web server interface")
	flag.Parse()

	http.HandleFunc("/", handleForm)
	http.HandleFunc("/stop", handleStop)
	http.HandleFunc("/toggle", handleToggle)
	http.HandleFunc("/seek", handleSeek)
	http.HandleFunc("/seek-backward", handleSeekBackward)
	http.HandleFunc("/seek-forward", handleSeekForward)
	http.HandleFunc("/show-osc", handleShowOsc)
	http.HandleFunc("/hide-osc", handleHideOsc)

	log.Printf("Server läuft auf http://%s:%s", *address, *port)
	log.Fatal(http.ListenAndServe(*address+":"+*port, nil))
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

	tmpl.Execute(w, struct {
		URL string
	}{currentURL})
}

func handleToggle(w http.ResponseWriter, r *http.Request) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	conn, err := net.Dial("unix", ipcSocketPath)
	if err != nil {
		log.Printf("error while opening IPC connection: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	defer conn.Close()

	cmd := map[string]any{
		"command": []any{"cycle", "pause"},
	}
	data, _ := json.Marshal(cmd)
	data = append(data, '\n')
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("error while requesting mpv: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleSeek(w http.ResponseWriter, r *http.Request) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	conn, err := net.Dial("unix", ipcSocketPath)
	if err != nil {
		log.Printf("error while opening IPC connection: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	defer conn.Close()

	position := r.URL.Query().Get("position")

	cmd := map[string]any{
		"command": []any{"seek", position, "absolute-percent"},
	}
	data, _ := json.Marshal(cmd)
	data = append(data, '\n')
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("error while requesting mpv: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleSeekBackward(w http.ResponseWriter, r *http.Request) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	conn, err := net.Dial("unix", ipcSocketPath)
	if err != nil {
		log.Printf("error while opening IPC connection: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	defer conn.Close()

	cmd := map[string]any{
		"command": []any{"seek", "-15"},
	}
	data, _ := json.Marshal(cmd)
	data = append(data, '\n')
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("error while requesting mpv: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleSeekForward(w http.ResponseWriter, r *http.Request) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	conn, err := net.Dial("unix", ipcSocketPath)
	if err != nil {
		log.Printf("error while opening IPC connection: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	defer conn.Close()

	cmd := map[string]any{
		"command": []any{"seek", "+15"},
	}
	data, _ := json.Marshal(cmd)
	data = append(data, '\n')
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("error while requesting mpv: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleShowOsc(w http.ResponseWriter, r *http.Request) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	conn, err := net.Dial("unix", ipcSocketPath)
	if err != nil {
		log.Printf("error while opening IPC connection: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	defer conn.Close()

	cmd := map[string]any{
		"command": []any{"script-message", "osc-visibility", "always"},
	}
	data, _ := json.Marshal(cmd)
	data = append(data, '\n')
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("error while requesting mpv: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleHideOsc(w http.ResponseWriter, r *http.Request) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	conn, err := net.Dial("unix", ipcSocketPath)
	if err != nil {
		log.Printf("error while opening IPC connection: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	defer conn.Close()

	cmd := map[string]any{
		"command": []any{"script-message", "osc-visibility", "auto"},
	}
	data, _ := json.Marshal(cmd)
	data = append(data, '\n')
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("error while requesting mpv: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func playWithMPV(url string) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	if currentCmd != nil && currentCmd.Process != nil {
		_ = currentCmd.Process.Kill()
	}

	_ = os.Remove(ipcSocketPath)

	cmd := exec.Command("mpv",
		url,
		"--ytdl-format=bestvideo[height<=?720]+bestaudio/best",
		"--fs",
		"--input-ipc-server="+ipcSocketPath,
	)

	err := cmd.Start()
	if err != nil {
		log.Printf("error while starting mpv: %v", err)
		return
	}

	currentCmd = cmd
	currentURL = url

	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Printf("mpv exited with error: %v", err)
		}
	}()
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	if currentCmd != nil && currentCmd.Process != nil {
		_ = currentCmd.Process.Kill()
		currentCmd = nil
	}
	currentURL = ""
	_ = os.Remove(ipcSocketPath)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
