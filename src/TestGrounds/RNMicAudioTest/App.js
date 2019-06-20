/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 * @flow
 */

import React, {Component} from 'react';
import {Platform, StyleSheet, Text, View} from 'react-native';

import MicStream from 'react-native-microphone-stream';

const instructions = Platform.select({
  ios: 'Press Cmd+R to reload,\n' + 'Cmd+D or shake for dev menu',
  android:
    'Double tap R on your keyboard to reload,\n' +
    'Shake or press menu button for dev menu',
});

class UpdatableLabel extends Component<Props> {
	constructor(Props) {
		super(Props)
		this.state = { 
		  text: 'Something here' 
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

type Props = {};
export default class App extends Component<Props> {

  constructor(Props) {
	super(Props)
	this.componentDidMount = this.componentDidMount.bind(this)
  }

  componentDidMount() {
	this.label.setLabel('Starting mic...')
	const listener = MicStream.addListener(data => console.log(data));
	MicStream.init({
  		bufferSize: 4096,
  		sampleRate: 44100,
  		bitsPerChannel: 16,
  		channelsPerFrame: 1,
	});
	MicStream.start();
	setTimeout( () => this.label.setLabel('1 sec'), 1000)
	setInterval( () => console.log("ouch"), 2000)
        this.label.setLabel('Mic started...')
  }

  render() {
    return (
      <View style={styles.container}>
	<UpdatableLabel ref={(r) => {this.label = r}}/>
        <Text style={styles.welcome}>Welcome to React Native!</Text>
        <Text style={styles.instructions}>To get started, edit App.js</Text>
        <Text style={styles.instructions}>{instructions}</Text>
      </View>
    );
  }


  componentWillUnmount() {

  }

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
