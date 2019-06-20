/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 * @flow
 */

import React, {Component} from 'react';
import {Platform, StyleSheet, Text, View} from 'react-native';

//var httpServer = require('react-native-http-server');

const instructions = Platform.select({
  ios: 'Press Cmd+R to reload,\n' + 'Cmd+D or shake for dev menu',
  android:
    'Double tap R on your keyboard to reload,\n' +
    'Shake or press menu button for dev menu',
});

var iplabel = null

class TextLabel extends Component<Props> {
        constructor(Props) {
                super(Props)
                //const ipv4Address = await NetworkInfo.getIPV4Address()
		this.state = {
                text: 'ip here'
                }
        }

        setLabel(newText) {
                this.setState({text:newText})
        }

        render() {
                return (
                        <Text style={styles.instructions}>{this.state.text}</Text>
                );
        }
}

type Props = {};
export default class App extends Component<Props> {
/*	
 componentWillMount(){
 
      var options = {
        port: 9111, // note that 80 is reserved on Android - an exception will be thrown
      };
 
      // initalise the server (now accessible via localhost:1234)
      httpServer.create(options, function(request, send){
 
          // interpret the url
          let url = request.url.split('/');
          let ext = url[1];
          let data = JSON.stringify({data: "hello world!", extension: ext});
 
          //Build our response object (you can specify status, mime_type (type), data, and response headers)
          let res = {};
          res.status = "OK";
          res.type = "application/json";
          res.data = data;
          res.headers = {
            "Access-Control-Allow-Credentials": "true",
            "Access-Control-Allow-Headers": "Authorization, Content-Type, Accept, Origin, User-Agent, Cache-Control, Keep-Alive, If-Modified-Since, If-None-Match",
            "Access-Control-Allow-Methods": "GET, HEAD",
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Expose-Headers": "Content-Type, Cache-Control, ETag, Expires, Last-Modified, Content-Length",
            "Access-Control-Max-Age": "3000",
            "Cache-Control": "max-age=300",
            "Connection": "keep-alive",
            "Content-Encoding": "gzip",
            "Content-Length": data.length.toString(),
            "Date": (new Date()).toUTCString(),
            "Last-Modified": (new Date()).toUTCString(),
            "Server": "Fastly",
            "Vary": "Accept-Encoding"
          };
 
          send(res);
 
      });
 
    }
 */

  render() {
    return (
      <View style={styles.container}>
        <TextLabel ref={(r) => refer = r}/>
      </View>
    );
  }

  componentDidMount(){


  };
/*
  componentWillUnmount() {
    httpServer.stop();
  }
*/
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
