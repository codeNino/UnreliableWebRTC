package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"gameserver/helper"

	"github.com/pion/webrtc/v3"
)

type player struct {
	sent_messages [][]byte
	datachan      *webrtc.DataChannel
}

//Updates Map
var Players sync.Map

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade: ", err)
		return
	}
	defer c.Close()
	fmt.Println("User connected from: ", c.RemoteAddr())

	//===========This Player's Identity===================
	var playerTag string

	//===========WEBRTC====================================
	// Create a new RTCPeerConnection

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	//Setup dataChannel to act like UDP with ordered messages (no retransmits)
	//with the DataChannelInit struct
	var udpPls webrtc.DataChannelInit
	var retransmits uint16 = 0

	//DataChannel will drop any messages older than
	//the most recent one received if ordered = true && retransmits = 0
	//This is nice so we can always assume client
	//side that the message received from the server
	//is the most recent update, and not have to
	//implement logic for handling old messages
	var ordered = true

	udpPls.Ordered = &ordered
	udpPls.MaxRetransmits = &retransmits

	// Create a datachannel with label 'UDP' and options udpPls
	dataChannel, err := peerConnection.CreateDataChannel("UDP", &udpPls)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())

		//3 = ICEConnectionStateConnected
		if connectionState.String() == "connected" {
			playerTag = uuid.New().String()
			Players.Store(playerTag, &player{sent_messages: [][]byte{}, datachan: dataChannel})
			fmt.Printf("Player  added with id %s \n", playerTag)

		} else if connectionState.String() == "failed" || connectionState.String() == "disconnected" || connectionState.String() == "closed" {
			Players.Delete(playerTag)
			fmt.Printf("Player removed with id %s \n", playerTag)

			err := peerConnection.Close() //deletes all references to this peerconnection in mem and same for ICE agent (ICE agent releases the "closed" status)
			if err != nil {               //https://www.w3.org/TR/webrtc/#dom-rtcpeerconnection-close
				fmt.Println(err)
			}
		}
	})

	// Register channel opening handling
	dataChannel.OnOpen(func() {
		//Send Client their playerTag
		sendErr := dataChannel.Send([]byte(playerTag))
		if sendErr != nil {
			panic(err)
		}
	})

	// Register message handling (Data all served as a bytes slice []byte) for user controls
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		//fmt.Printf("Message from DataChannel '%s': '%s'\n", reliableChannel.Label(), string(msg.Data))
		if loaded, ok := Players.Load(playerTag); ok {
			// Assert the player data value to the player type
			if playerdata, ok := loaded.(*player); ok {
				playerdata.sent_messages = append(playerdata.sent_messages, msg.Data)
				Players.Store(playerTag, playerdata)
				// outputMessage := fmt.Sprintf("Received Data from Player with id '%s': '%s'\n", playerTag, string(msg.Data))
				fmt.Printf("Received Data from Player with id '%s': '%s'\n", playerTag, string(msg.Data))
				BroadCastMessageToPeers(fmt.Sprintf("Player with ID '%s' says :  '%s'\n", playerTag, string(msg.Data)),
					playerTag)
			} else {
				fmt.Println("Value found in Store is not of type player")
			}
		} else {
			fmt.Println("player data not found")
		}

	})

	//==============================================================================

	// Create an offer to send to the browser
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	fmt.Println(*peerConnection.LocalDescription())

	//Send the SDP with the final ICE candidate to the browser as our offer
	err = c.WriteMessage(1, []byte(helper.Encode(*peerConnection.LocalDescription()))) //write message back to browser, 1 means message in byte format?
	if err != nil {
		fmt.Println("write:", err)
	}

	//Wait for the browser to send an answer (its SDP)
	msgType, message, err2 := c.ReadMessage() //ReadMessage blocks until message received
	if err2 != nil {
		fmt.Println("read:", err)
	}

	answer := webrtc.SessionDescription{}

	helper.Decode(string(message), &answer) //set answer to the decoded SDP
	fmt.Println(answer, msgType)

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		panic(err)
	}

}

// }

//We'll have this marshalling function here so the multiple goroutines for each
//player will not be inefficient by all trying to marshall the same thing
func GetSyncMapReadyForSending(m *sync.Map) {
	for {
		time.Sleep(time.Millisecond)

		tmpMap := make(map[string]*player)
		m.Range(func(k, v interface{}) bool {
			tmpMap[k.(string)] = v.(*player)
			return true
		})

		_, err := json.Marshal(tmpMap)
		if err != nil {
			panic(err)
		}
	}
}

func BroadCastMessageToPeers(message, senderTag string) {
	Players.Range(func(key, value interface{}) bool {
		p := value.(*player)
		if key != senderTag {
			p.datachan.Send([]byte(message))
		}
		return true
	})
}
