/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 * @flow
 */

import React, {Component} from 'react';
import {Platform, StyleSheet, Text, Image, View} from 'react-native';

global.Buffer = global.Buffer || require('buffer').Buffer

const MulticastUDPPort = "9191"
const UnicastUDPPort = "9190"
const RESTPort = "9192"

var dgram = require('dgram')
// OR, if not shimming via package.json "browser" field:
// var dgram = require('react-native-udp')
var socket = dgram.createSocket('udp4')
//socket.bind(9190) //UnicastUDPPort)
//socket.once('listening', function() {
//  	alert("listening")
//})
 
//socket.on('message', function(msg, rinfo) {
//  alert("message: " + msg.toString())
//})

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


export default class App extends Component<Props> {

  render() {
    return (
      <View style={styles.container}>
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
