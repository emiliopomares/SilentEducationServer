/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 * @flow
 */

import React, {Component} from 'react';
import {TouchableOpacity, Switch, Image, Platform, StyleSheet, Text, View} from 'react-native';
import Slider from "react-native-slider";
import { Akira } from 'react-native-textinput-effects';

const instructions = Platform.select({
  ios: 'Press Cmd+R to reload,\n' + 'Cmd+D or shake for dev menu',
  android:
    'Double tap R on your keyboard to reload,\n' +
    'Shake or press menu button for dev menu',
});

var DEFAULT_VALUE = 0.2
var SCALE = 100

class SliderContainer extends Component<Props> {

 constructor(Props) {
	super(Props)
	this.state = {
		value: DEFAULT_VALUE
	}
 }

  getInitialState() {
    return {
      value: DEFAULT_VALUE,
    };
  }

  computeValue() {
	return Math.floor(Number(this.props.scaleMin) + 
 		(Number(this.props.scaleMax) - Number(this.props.scaleMin)) * this.state.value)
  }

  render() {
    var value = this.state.value;

    return (
      <View>
        <View style={styles.titleContainer}>
	    <Text style={styles.caption} numberOfLines={1}>{this.props.caption}</Text>
	    <Text style={styles.value} numberOfLines={1}>{this.computeValue()}</Text>
	</View>
        {this._renderChildren()}
      </View>
    );
  }

  _renderChildren() {
    return React.Children.map(this.props.children, (child) => {
      if (child.type === Slider
          || child.type === ReactNative.Slider) {
        var value = this.state.value;
        return React.cloneElement(child, {
          value: value,
          onValueChange: (val) => this.setState({value: val}),
        });
      } else {
        return child;
      }
    });
  }
}

class PairStatus extends Component<Props> {
  constructor(Props) {
	super(Props)
	this.state = {
		pairStage: 'unpaired',
		deviceName: '(No emparejado)',
		PIN: ''
        }
	this.renderUnpaired = this.renderUnpaired.bind(this)
  	this.renderPairing = this.renderPairing.bind(this)
  	this.renderPaired = this.renderPaired.bind(this)
  	this.cancelPairing = this.cancelPairing.bind(this)
	this.sendPIN = this.sendPIN.bind(this)
	this.unpair = this.unpair.bind(this)
	this.pair = this.pair.bind(this)
	this.setPIN = this.setPIN.bind(this)
  }

  renderUnpaired() {
	return (
	   <View style={{alignItems: 'center', justifyContent: 'center'}}>
		<Image
                	style={{width:270,resizeMode: 'contain'}}
                	source={require('./assets/Logo.png')}
        	/>
		<Text>No emparejado</Text>
		<View style={{alignItems:'flex-end', justifyContent: 'flex-end'}}>
			<TouchableOpacity onPress={this.pair} style={styles.FacebookStyleSmall} activeOpacity={0.5}>
                        	<Text>Emparejar</Text>
                	</TouchableOpacity>
		</View>
	   </View>
	);
  }

  setPIN(pin) {
    this.setState({pairStage: this.state.pairStage, PIN: pin, deviceName: this.state.deviceName})
  }

  pair() {
    this.setState({pairStage:'pairing', PIN: this.state.PIN, deviceName: this.state.deviceName})
  }

  sendPIN() {
    this.setState({pairStage:'paired', PIN: this.state.PIN, deviceName: this.state.deviceName})
  }

  cancelPairing() {
    this.setState({pairStage:'unpaired', PIN: '', deviceName: this.state.deviceName})
  }

  unpair() {
	this.cancelPairing()
  }

  renderPairing() {
	return (
           <View style={{alignItems: 'center', justifyContent: 'center'}}>
                <Image
                        style={{width:270,resizeMode: 'contain'}}
                        source={require('./assets/Logo.png')}
                />
		<Akira
                	label={'PIN'}
                	value={this.state.PIN}
                	style={{width:260}}
                	borderColor={'#a5d1cc'}
                	onChangeText={(pin)=>{this.setPIN(pin)}}
                	inputPadding={16}
                	labelHeight={24}
                	labelStyle={{ color: '#ac83c4' }}
        	/>
                <View style={{flexDirection:'row', alignItems:'flex-end', justifyContent: 'flex-end'}}>
                        <TouchableOpacity onPress={this.sendPIN} style={styles.FacebookStyleSmall} activeOpacity={0.5}>
                                <Text>Enviar</Text>
                        </TouchableOpacity>
			<TouchableOpacity onPress={this.cancelPairing} style={styles.FacebookStyleSmall} activeOpacity={0.5}>
                                <Text>Cancelar</Text>
                        </TouchableOpacity>
                </View>
           </View>
        );
  }

  renderPaired() {
	return (
           <View style={{alignItems: 'center', justifyContent: 'center'}}>
                <Image
                        style={{width:270,resizeMode: 'contain'}}
                        source={require('./assets/Logo.png')}
                />
                <Text>{this.state.deviceName}</Text>
                <View style={{flexDirection:'row', alignItems:'flex-end', justifyContent: 'flex-end'}}>
                        <TouchableOpacity onPress={this.unpair} style={styles.FacebookStyleSmall} activeOpacity={0.5}>
                                <Text>Desemparejar</Text>
                        </TouchableOpacity>
                </View>
           </View>
        );	
  }

  render() {
	switch (this.state.pairStage) {
		case 'unpaired':
			return this.renderUnpaired()
		case 'pairing':
			return this.renderPairing()
		default:	
			return this.renderPaired()
	}
  }

}
/*
good old global references
*/

var MessageRef = null

type Props = {};
export default class App extends Component<Props> {

  constructor(Props) {
	super(Props)
 	this.toggleSwitch = this.toggleSwitch.bind(this)
	this.toDeviceButton = this.toDeviceButton.bind(this)
	this.toServerButton = this.toDeviceButton.bind(this)
	this.setMessage = this.setMessage.bind(this)
  }

  state = {
	message: '',
 	switches: {
		redSwitchValue: true,
		orangeSwitchValue: false,
		greenSwitchValue: false,
	}
  }

  setMessage(text) {
	this.setState({message:text,switches:this.state.switches})
  }

  toDeviceButton() {
	this.setState({message:'',switches:this.state.switches})
  }

  toServerButton() {
	this.setState({message:'',switches:this.state.switches})
  }

  toggleSwitch(which) {
	switch (which) {
		case 'red':
			this.setState({redSwitchValue : true, orangeSwitchValue : false, greenSwitchValue : false});
  			break;
		case 'orange':
			this.setState({redSwitchValue : false, orangeSwitchValue : true, greenSwitchValue : false});
			break;
  		case 'green':
			this.setState({redSwitchValue : false, orangeSwitchValue : false, greenSwitchValue : true});
                        break;
	}
  }
  

  render() {
    return (
      <View style={styles.container}>
       
	<PairStatus/>	
 
 	<SliderContainer caption='Volumen (%)' scaleMin='0' scaleMax='100'>
          <Slider
            trackStyle={customStyles3.track}
            thumbStyle={customStyles3.thumb}
            minimumTrackTintColor='#eecba8'
     	  />
	</SliderContainer>
	

	<SliderContainer caption='Frecuencia (Hz)' scaleMin='8000' scaleMax='20000'>
          <Slider
            trackStyle={customStyles3.track}
            thumbStyle={customStyles3.thumb}
            minimumTrackTintColor='#5b8bd8'
          />
        </SliderContainer>	


	<SliderContainer caption='Umbral (%)' scaleMin='0' scaleMax='100'>
          <Slider
            trackStyle={customStyles3.track}
            thumbStyle={customStyles3.thumb}
            minimumTrackTintColor='#8cd85b'
          />
        </SliderContainer>

	<SliderContainer caption='Duración (s)' scaleMin='1' scaleMax='30'>
          <Slider
            trackStyle={customStyles3.track}
            thumbStyle={customStyles3.thumb}
            minimumTrackTintColor='#d89d5b'
          />
        </SliderContainer>

	<View>
	  <View style={{flexDirection: 'row'}}>
		<TouchableOpacity style={styles.FacebookStyle} activeOpacity={0.5}>
    			<Image
     				source={require('./assets/Speaker.png')}
     				style={styles.ImageIconStyle}
    			/>
		</TouchableOpacity>
	
		<TouchableOpacity style={styles.FacebookStyle} activeOpacity={0.5}>
                        <Image
                                source={require('./assets/Walkie.png')}
                                style={styles.ImageIconStyle}
                        />
                </TouchableOpacity>
 	   </View>
	</View>

	<Akira
    		label={'Mensaje'}
		value={this.state.message}
    		style={{width:260}}
		borderColor={'#a5d1cc'}
		onChangeText={(text)=>{this.setMessage(text)}}    
		inputPadding={16}
    		labelHeight={24}
    		labelStyle={{ color: '#ac83c4' }}
  	/>	


	<View style={{flexDirection: 'row'}}>
          <Switch
            style={{marginTop:30}}
	    trackColor={{true:'#FF0000', false: null}}
            value = {this.state.switches.redSwitchValue}/>
          <Switch
            style={{marginTop:30}}
            trackColor={{true:'#FFAA00', false: null}}
            value = {this.state.switches.orangeSwitchValue}/>
   	  <Switch
            style={{marginTop:30}}
            trackColor={{true:'#00FF00', false: null}}
            value = {this.state.switches.greenSwitchValue}/>
	</View>

	<View>
          <View style={{flexDirection: 'column'}}>
                <TouchableOpacity onPress={this.toDeviceButton} style={styles.FacebookStyle2} activeOpacity={0.5}>
                	<Text>A dispositivo</Text>
		</TouchableOpacity>

                <TouchableOpacity onPress={this.toServerButton} style={styles.FacebookStyle2} activeOpacity={0.5}>
                	<Text>A servidor central</Text>
		</TouchableOpacity>
           </View>
        </View>

      </View>
    );
  }
}

var customStyles3 = StyleSheet.create({
  track: {
    height: 10, 
    width: 250,
    borderRadius: 5,
    backgroundColor: '#d0d0d0',
  },
  thumb: {
    width: 10,
    height: 30,
    borderRadius: 5,
    backgroundColor: '#eb6e1b',
  }
});

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
  FacebookStyle: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#aabbcc',
    borderWidth: 0.5,
    borderColor: '#fff',
    height: 40,
    width: 120,
    borderRadius: 5,
    margin: 5,
  },
  FacebookStyleSmall: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#aabbcc',
    borderWidth: 0.5,
    borderColor: '#fff',
    height: 35,
    width: 100,
    borderRadius: 5,
    margin: 5,
  },
  FacebookStyle2: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#aabbcc',
    borderWidth: 0.5,
    borderColor: '#fff',
    height: 40,
    width: 180,
    borderRadius: 5,
    margin: 5,
  },
  ImageIconStyle: {
    padding: 10,
    margin: 5,
    height: 25,
    width: 25,
    resizeMode: 'stretch',
  },
  instructions: {
    textAlign: 'center',
    color: '#333333',
    marginBottom: 5,
  },
});
