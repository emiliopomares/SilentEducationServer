/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 * @flow
 */

import React, {Component} from 'react';
import {Platform, StyleSheet, Button, Text, TextInput, Image, View} from 'react-native';

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


const MulticastHost = '224.0.0.1';

var message = new Buffer('supplicant');

var client = dgram.createSocket('udp4');
client.send(message, 0, message.length, 9191, MulticastHost, function(err, bytes) {
  if (err) throw err;
  //console.log('UDP message sent to ' + HOST +':'+ PORT);
  client.close();
});

var connectionStatus = null


const instructions = Platform.select({
  ios: 'Press Cmd+R to reload,\n' + 'Cmd+D or shake for dev menu',
  android:
    'Double tap R on your keyboard to reload,\n' +
    'Shake or press menu button for dev menu',
});

type Props = {};


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

  render() {
    return (
      <View style={styles.container}>
	<PairRequest/>
	<ConnectionStatus ref={(c) => connectionStatus = c}/>
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
	}
	if(data.deviceip !== undefined) {
		alert('Device responded with IP: ' + data.deviceip)
	} 
})

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
