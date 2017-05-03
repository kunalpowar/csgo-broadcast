# CSGO GOTV Broadcast sytem

This is a POC code to test the CSGO's [GOTV broadcast](https://developer.valvesoftware.com/wiki/Counter-Strike:_Global_Offensive_Broadcast) feature.

## Usage
* `go run main.go`
    - This will start the webserver on port 3090 (will be configurable soon)
* Use something like ngrok to get a public endpoint 
    - `ngrok http 3090`
    - Say the endpoint is `http://6b96d99b.ngrok.io`
* On your CSGO server set
    - `tv_broadcast_url "http://6b96d99b.ngrok.io"`
    - `tv_broadcast 1`
* Currently use the logs from the webserver to figure out the token
* In your CSGO client use ` playcast "http://gotv-cdn.example.com/match/<token>"` to watch the broadcast 

## TODO
* Change to make this a standalone executable 
* Maintain releases
* Add flags to configure the server
