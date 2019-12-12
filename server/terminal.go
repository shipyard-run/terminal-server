package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
)

// HandleTerminal creates a new terminal server.
func HandleTerminal(r *mux.Router) {
	r.HandleFunc("", handleWebsocket)
}

type windowSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
	X    uint16
	Y    uint16
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	workdir := r.URL.Query().Get("workdir")
	user := r.URL.Query().Get("user")
	target := r.URL.Query().Get("target")
	shell := r.URL.Query().Get("shell")

	log.Printf("docker exec -ti -w %s -u %s %s %s", workdir, user, target, shell)


	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(fmt.Errorf("Unable to upgrade connection"))
		return
	}

	// cmd := exec.Command("/bin/bash", "-l") // #nosec
	cmd := exec.Command("/usr/bin/docker", "exec", "-ti", "-w", workdir, "-u", user, target, shell)
	cmd.Env = append(os.Environ(), "TERM=xterm")

	tty, err := pty.Start(cmd)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		log.Println(fmt.Errorf("Unable to start pty/cmd"))
		return
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Process.Wait()
		tty.Close()
		conn.Close()
	}()

	go func() {
		for {
			buf := make([]byte, 1024)
			read, err := tty.Read(buf)
			if err != nil {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
				log.Println(fmt.Errorf("Unable to read from pty/cmd"))
				return
			}
			_ = conn.WriteMessage(websocket.BinaryMessage, buf[:read])
		}
	}()

	for {
		_, reader, err := conn.NextReader()
		if err != nil {
			log.Println(fmt.Errorf("Unable to grab next reader"))
			return
		}

		dataTypeBuf := make([]byte, 1)
		read, err := reader.Read(dataTypeBuf)
		if err != nil {
			log.Println(fmt.Errorf("Unable to read message type from reader"))
			_ = conn.WriteMessage(websocket.TextMessage, []byte("Unable to read message type from reader"))
			return
		}

		if read != 1 {
			log.Println(fmt.Errorf("Unexpected number of bytes read"))
			return
		}

		switch dataTypeBuf[0] {
		case 0:
			copied, err := io.Copy(tty, reader)
			if err != nil {
				log.Println(fmt.Errorf("Error after copying %d bytes", copied))
			}
		case 1:
			decoder := json.NewDecoder(reader)
			resizeMessage := windowSize{}
			err := decoder.Decode(&resizeMessage)
			if err != nil {
				_ = conn.WriteMessage(websocket.TextMessage, []byte("Error decoding resize message: "+err.Error()))
				continue
			}
			log.Println("Resizing terminal")
			// #nosec G103
			_, _, errno := syscall.Syscall(
				syscall.SYS_IOCTL,
				tty.Fd(),
				syscall.TIOCSWINSZ,
				uintptr(unsafe.Pointer(&resizeMessage)),
			)
			if errno != 0 {
				log.Println(fmt.Errorf("Unable to resize terminal"))
			}
		default:
			log.Println(fmt.Errorf("Unknown data type"))
		}
	}
}
