/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 * @flow
 */

import React, {Component} from 'react';
import {Platform, StyleSheet, Button, Text, TextInput, Image, View} from 'react-native';
import { AsyncStorage } from 'react-native'
import Slider from "react-native-slider";

global.Buffer = global.Buffer || require('buffer').Buffer

const MulticastUDPPort = "9191"
const MulticastUDPPortNumber = 9191
const UnicastUDPPort = "9190"
const RESTPort = "9192"

var dgram = require('dgram')
var socket = dgram.createSocket('udp4')

const StatusRed = require('./assets/red.png')
const StatusYellow = require('./assets/yellow.png')
const StatusGreen = require('./assets/green.png')


// schema for the prototype: simple PSK, better than nothing!
// @TODO send this to a env variable
const SilentEducationPSK = "4baUV/2T=1a4nGrDS43FGnv6100asRNa35+shd/2b42300aNUFHsdn2m3iUJ86B/d2"


const MulticastHost = '224.0.0.1';

var message = new Buffer('supplicant');

var client = dgram.createSocket('udp4');
client.send(message, 0, message.length, 9191, MulticastHost, function(err, bytes) {
  if (err) throw err;
  //console.log('UDP message sent to ' + HOST +':'+ PORT);
  client.close();
});

var connectionStatus = null
var statusLabel = null

const instructions = Platform.select({
  ios: 'Press Cmd+R to reload,\n' + 'Cmd+D or shake for dev menu',
  android:
    'Double tap R on your keyboard to reload,\n' +
    'Shake or press menu button for dev menu',
});

type Props = {};

var AppInfo = 
  {

	deviceIP : '',
	pairSecret: '',
	lastKnownName: '',
        serverIP: '',

  }

class MutableLabel extends Component<Props> {
	constructor(Props) {
		super(Props)
		this.state = {
			text : ''
		}
		this.setLabel = this.setLabel.bind(this)
	}

	setLabel(newText) {
		this.setState({text:newText})
	}

	render() {
		return (
			<Text>{this.state.text}</Text>
		);
	}
	
}

class ConnectionStatus extends Component<Props> {
        constructor(Props) {
                super(Props)
                this.state = {
                        image: StatusRed
                }
        }

        setStatus(state) {
                this.setState({image:state})
        }

        render() {
                return (
                        <Image source={this.state.image}/>
                );
        }
}

class PairRequest extends Component<Props> {
	
	thisContext = null
	PINTextInput = null

	constructor(Props) {
		super(Props)
		this.state = {
			pairStage: 'Not paired',
			shitState: 'shit',
			pin: '',
		}
		thisContext = this

		this.handleTextInputChange = this.handleTextInputChange.bind(this);
		this.attemptPairing = this.attemptPairing.bind(this);
		this.completePairing = this.completePairing.bind(this);
	}

	completePairing() {
		alert('complerting pairing. seding "Whois'+this.state.pin+'"')
		var message = new Buffer("Whois"+this.state.pin);
                socket.send(message, 0, message.length, MulticastUDPPortNumber, MulticastHost, function(err, bytes) {});
		this.setState({pairStage:'Paired', shitState:'shit', pin:this.state.pin});
	}	

	startPairing() {
		this.setState({pairStage:'Pairing', shitState:'shit', pin:this.state.pin});
	}

	cancelPairing() {
		this.setState({pairStage:'Not paired', shitState:'shit', pin:this.state.pin});
	}

	requestPairing = function() {
                  var message = new Buffer("PairRequest");
                  socket.send(message, 0, message.length, MulticastUDPPortNumber, MulticastHost, function(err, bytes) {});
                  thisContext.startPairing();
                }

	attemptPairing() {
		this.completePairing()
	}

	handleTextInputChange(newText) {
    		this.setState({pairStage:this.state.pairStage, shitState:'shit', pin:newText});
	}

	render() {
		if(this.state.pairStage == 'Not paired') {
			return (
				<View>
					<Text>Sin emparejar</Text>
					<Button onPress={this.requestPairing} title="Emparejar" color="#841584" accessibilityLabel="Request pairing to a device"/>
				</View>
			);
		}
		else if(this.state.pairStage == 'Pairing') {
			return (
                                <View>
                                        <Text>Introduce PIN del dispositivo:</Text>
					<TextInput value={this.state.pin} onChangeText={(text) => this.handleTextInputChange(text)}/>
                                        <Button onPress={this.attemptPairing} title="Introducir" color="#2188A3" accessibilityLabel="Submit PIN"/>
                                </View>
                        );
		}
		else {
			return null; 
		}
	}

}


export default class App extends Component<Props> {
  constructor(Props) {
	super(Props)
	this.state = {
		value: 0
	}
  }

  componentDidMount() {

	loadAppInfo().then( function() {
		if(AppInfo.pairSecret == '') {
			statusLabel.setLabel('No emparejado')
		} 
		else {
			statusLabel.setLabel('Emparejado con: ' + AppInfo.lastKnownName)
		}
	})

  }

  render() {
    return (
      <View style={styles.container}>
	<MutableLabel ref={(c) => statusLabel = c}/>
	<PairRequest/>
	<ConnectionStatus ref={(c) => connectionStatus = c}/>
	<Slider
          value={this.state.value}
          onValueChange={value => this.setState({ value })}
        />
        <Text style={styles.welcome}>Welcome to React Native!</Text>
        <Text style={styles.instructions}>To get started, edit App.js</Text>
        <Text style={styles.instructions}>{instructions}</Text>
      </View>
    );
  }


}

var socket = dgram.createSocket('udp4')
socket.bind(9190) //UnicastUDPPort)
socket.once('listening', function() {
	if(connectionStatus != null) {
		connectionStatus.setStatus(StatusRed);
	}
})


socket.on('message', function(msg, rinfo) {
        if(connectionStatus != null) {
		connectionStatus.setStatus(StatusYellow);
	}
	data = JSON.parse(msg)
	if(data.serverip !== undefined) {
		alert('Server responded with IP: ' + data.serverip)
		AppInfo.serverIP = data.serverip
	}
	if(data.deviceip !== undefined) {
		AppInfo.deviceIP = data.deviceip
		url = 'http://' + AppInfo.deviceIP + ':8000/pairing'
		//alert('Device responded with IP: ' + data.deviceip + '. Requesting ' + url)
		POST(url, "", function(data) {
			AppInfo.pairSecret = data.secret
			connectionStatus.setStatus(StatusGreen)
			statusUrl = 'http://' + AppInfo.deviceIP + ':8000/status'
			GET(statusUrl, function(status) {
				statusLabel.setLabel('Conectado a ' + status.name)
				AppInfo.lastKnownName = status.name
				saveAppInfo()
			})
		})
	} 
})

async function updateConnectionStatus() {
	// find out about server
	// find out about device
}

async function saveAppInfo() {
	try {
            await AsyncStorage.setItem('DeviceIP', AppInfo.deviceIP);
        } catch (error) {
           alert('Error guardando datos: ' + error) 
        }	

	try {
            await AsyncStorage.setItem('PairSecret', AppInfo.pairSecret);
        } catch (error) {
           alert('Error guardando datos: ' + error)
        }

	try {
            await AsyncStorage.setItem('LastKnownName', AppInfo.lastKnownName);
        } catch (error) {
           alert('Error guardando datos: ' + error)
        }
}

async function loadAppInfo() {
	try {
            const value = await AsyncStorage.getItem('DeviceIP');
            if (value !== null) {
            	AppInfo.deviceIP = value
	    }
        } catch (error) {
            alert('Error cargando datos: ' + error)
        }

	try {
            const value = await AsyncStorage.getItem('PairSecret');
            if (value !== null) {
                AppInfo.pairSecret = value
            }
        } catch (error) {
            alert('Error cargando datos: ' + error)
        }

 	try {
            const value = await AsyncStorage.getItem('LastKnownName');
            if (value !== null) {
                AppInfo.lastKnownName = value
            }
        } catch (error) {
            alert('Error cargando datos: ' + error)
        }
}

function GET(url, callback) {
	const options = {
		method: 'GET',
		headers: new Headers({'psk':SilentEducationPSK}),
	}
	fetch(url, options).then(response => response.json()).then( (data) => callback(data) )
}

function POST(url, body, callback) {
	const options = {
		method: 'POST',
		headers: new Headers(
			{
				'psk':SilentEducationPSK,
				'Content-Type': 'application/x-www-form-urlencoded'
			}),
		body: body
	}
	fetch(url, options).then(response => response.json()).then((data)=>callback(data))
}

function DELETE(url, callback) {

}

function PUT(url, body, callback) {

}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#F5FCFF',
  },
  welcome: {
    fontSize: 20,
    textAlign: 'center',
    margin: 10,
  },
  instructions: {
    textAlign: 'center',
    color: '#333333',
    marginBottom: 5,
  },
});
