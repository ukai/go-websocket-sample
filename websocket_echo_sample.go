package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"code.google.com/p/go.net/websocket"
)

var port *int = flag.Int("p", 23456, "Port to listen.")

// copyServer echoes back messages sent from client using io.Copy.
func copyServer(ws *websocket.Conn) {
	fmt.Printf("copyServer %#v\n", ws.Config())
	io.Copy(ws, ws)
	fmt.Println("copyServer finished")
}

// readWriteServer echoes back messages sent from client using Read and Write.
func readWriteServer(ws *websocket.Conn) {
	fmt.Printf("readWriteServer %#v\n", ws.Config())
	for {
		buf := make([]byte, 100)
		// Read at most 100 bytes.  If client sends a message more than
		// 100 bytes, first Read just reads first 100 bytes.
		// Next Read will read next at most 100 bytes.
		n, err := ws.Read(buf)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("recv:%q\n", buf[:n])
		// Write send a message to the client.
		n, err = ws.Write(buf[:n])
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("send:%q\n", buf[:n])
	}
	fmt.Println("readWriteServer finished")
}

// sendRecvServer echoes back text messages sent from client
// using websocket.Message.
func sendRecvServer(ws *websocket.Conn) {
	fmt.Printf("sendRecvServer %#v\n", ws)
	for {
		var buf string
		// Receive receives a text message from client, since buf is string.
		err := websocket.Message.Receive(ws, &buf)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("recv:%q\n", buf)
		// Send sends a text message to client, since buf is string.
		err = websocket.Message.Send(ws, buf)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("send:%q\n", buf)
	}
	fmt.Println("sendRecvServer finished")
}

// sendRecvBinaryServer echoes back binary messages sent from clent
// using websocket.Message.
// Note that chrome supports binary messaging in 15.0.874.* or later.
func sendRecvBinaryServer(ws *websocket.Conn) {
	fmt.Printf("sendRecvBinaryServer %#v\n", ws)
	for {
		var buf []byte
		// Receive receives a binary message from client, since buf is []byte.
		err := websocket.Message.Receive(ws, &buf)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("recv:%#v\n", buf)
		// Send sends a binary message to client, since buf is []byte.
		err = websocket.Message.Send(ws, buf)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("send:%#v\n", buf)
	}
	fmt.Println("sendRecvBinaryServer finished")
}

type T struct {
	Msg  string
	Path string
}

// jsonServer echoes back json string sent from client using websocket.JSON.
func jsonServer(ws *websocket.Conn) {
	fmt.Printf("jsonServer %#v\n", ws.Config())
	for {
		var msg T
		// Receive receives a text message serialized T as JSON.
		err := websocket.JSON.Receive(ws, &msg)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("recv:%#v\n", msg)
		// Send send a text message serialized T as JSON.
		err = websocket.JSON.Send(ws, msg)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("send:%#v\n", msg)
	}
}

func MainServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, `<html>
<head>
<script type="text/javascript">
var path;
var ws;
function init() {
   console.log("init");
   if (ws != null) {
     ws.close();
     ws = null;
   }
   path = document.msgform.path.value;
   console.log("path:" + path);
   var div = document.getElementById("msg");
   div.innerText = "path:" + path + "\n" + div.innerText;
   ws = new WebSocket("ws://localhost:23456" + path);
   if (path == "/sendRecvBlob") {
     ws.binaryType = "blob";
   } else if (path == "/sendRecvArrayBuffer") {
     ws.binaryType = "arraybuffer";
   }
   ws.onopen = function () {
      div.innerText = "opened\n" + div.innerText;
   };
   ws.onmessage = function (e) {
      div.innerText = "msg:" + e.data + "\n" + div.innerText;
      if (e.data instanceof ArrayBuffer) {
        s = "ArrayBuffer: " + e.data.byteLength + "[";
        var view = new Uint8Array(e.data);
        for (var i = 0; i < view.length; ++i) {
          s += " " + view[i];
        }
        s += "]";
        div.innerText = s + "\n" + div.innerText;
      }
   };
   ws.onclose = function (e) {
      div.innerText = "closed\n" + div.innerText;
   };
   console.log("init");
   div.innerText = "init\n" + div.innerText;
};
function send() {
   console.log("send");
   var m = document.msgform.message.value;
   if (path == "/sendRecvArrayBuffer" || path == "/sendRecvBlob") {
     var t = m;
     if (t != "") {
       var array = new Uint8Array(t.length);
       for (var i = 0; i < t.length; i++) {
          array[i] = t.charCodeAt(i);
       }
       m = array.buffer;
     } else {
     m = document.msgform.file.files[0];
     }
   } else if (path == "/json") {
     m = JSON.stringify({Msg: m, Path: path})
   }
   console.log("send:" + m);
   if (m instanceof ArrayBuffer) {
     var s = "arrayBuffer:" + m.byteLength + "[";
     var view = new Uint8Array(m);
     for (var i = 0; i < m.byteLength; ++i) {
      s += " " + view[i];
     }
     s += "]";
     console.log(s);
   }
   ws.send(m);
   return false;
};
</script>
<body onLoad="init();">
<form name="msgform" action="#" onsubmit="return send();">
<select onchange="init()" name="path">
<option value="/copy" selected="selected">/copy</option>
<option value="/readWrite">/readWrite</option>
<option value="/sendRecvText">/sendRecvText</option>
<option value="/sendRecvArrayBuffer">/sendRecvArrayBuffer</option>
<option value="/sendRecvBlob">/sendRecvBlob</option>
<option value="/json">/json</option>
</select>
<input type="text" name="message" size="80" value="">
<input type="file" name="file" >
<input type="submit" value="send">
</form>
<div id="msg"></div>
</html>
`)
}

func main() {
	flag.Parse()
	http.Handle("/copy", websocket.Handler(copyServer))
	http.Handle("/readWrite", websocket.Handler(readWriteServer))
	http.Handle("/sendRecvText", websocket.Handler(sendRecvServer))
	http.Handle("/sendRecvArrayBuffer", websocket.Handler(sendRecvBinaryServer))
	http.Handle("/sendRecvBlob", websocket.Handler(sendRecvBinaryServer))
	http.Handle("/json", websocket.Handler(jsonServer))
	http.HandleFunc("/", MainServer)
	fmt.Printf("http://localhost:%d/\n", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		panic("ListenANdServe: " + err.Error())
	}
}
