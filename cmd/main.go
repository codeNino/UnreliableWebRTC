package main

import (
	"fmt"
	"gameserver/handler"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {

	r := mux.NewRouter()

	go handler.GetSyncMapReadyForSending(&handler.Players)
	go handler.StartSendingBinaryMessages()

	// Listen on UDP Port 80, will be used for all WebRTC traffic
	udpListener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IP{0, 0, 0, 0},
		Port: 80,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Listening for WebRTC traffic at %s\n", udpListener.LocalAddr())

	// Create a SettingEngine, this allows non-standard WebRTC behavior
	settingEngine := webrtc.SettingEngine{}

	//Our Public Candidate is declared here cause were not using a STUN server for discovery
	//and just hardcoding the open port, and port forwarding webrtc traffic on the router
	settingEngine.SetNAT1To1IPs([]string{}, webrtc.ICECandidateTypeHost)

	// Configure our SettingEngine to use our UDPMux. By default a PeerConnection has
	// no global state. The API+SettingEngine allows the user to share state between them.
	// In this case we are sharing our listening port across many.
	settingEngine.SetICEUDPMux(webrtc.NewICEUDPMux(nil, udpListener))

	fileServer := http.FileServer(http.Dir("./web"))
	r.HandleFunc("/echo", handler.Echo) //this request comes from webrtc.html
	r.Handle("/", fileServer)
	http.Handle("/", r)

	err = http.ListenAndServe(":80", nil) //Http server blocks
	if err != nil {
		panic(err)
	}
}
