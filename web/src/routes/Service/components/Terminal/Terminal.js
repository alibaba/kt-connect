import React, { Component } from "react";
import { Terminal } from 'xterm';
import * as fit from 'xterm/lib/addons/fit/fit';

import 'xterm/dist/xterm.css';
import './terminal.scss';

Terminal.applyAddon(fit);

export default class TerminalExexutor extends Component {
  constructor(props) {
    super(props);
    this.term = new Terminal({
      theme: {
        background: '#001529',
      }
    });
    this.state = {
      messages: [],
    }
  }

  componentDidMount() {
    this.startTerm(this.props)
  }

  componentWillReceiveProps(nextProps) {
    this.term.writeln('Reconnecting...')
    this.connect(nextProps)
  }

  componentWillUnmount() {
    this.close();
  }

  startTerm = (props) => {
    this.term.open(document.getElementById('terminal'))
    this.term.fit();
    this.term.focus();
    this.connect(props);
    this.term.on('data', (input) => {
      var msg = { type: "input", input: input }
      this.socket.send(JSON.stringify(msg))
    });
  }

  connect = (props) => {
    const { namespace, name, container, shell } = props;
    if (!namespace || !name || !container) {
      return;
    }
    this.close();
    this.socket = new WebSocket(`ws://${window.location.host}/ws/terminal?ns=${namespace}&p=${name}&c=${container}&s=${shell}`);
    this.socket.onmessage = (event) => {
      this.term.write(event.data)
    }
    this.socket.onopen = () => {
      console.log("onopen")
    }
    this.socket.onclose = () => {
      console.log('Good bye');
    }
    window.addEventListener("resize", () => {
      this.term.fit()
      var msg = { type: "resize", rows: this.term.rows, cols: this.term.cols }
      this.socket.send(JSON.stringify(msg))
    })
  }

  close = () => {
    if (this.socket) {
      this.socket.close()
    }
    if (this.keeplive) {
      clearInterval(this.keeplive)
    }
  }

  send(data) {
    this.socket.send(data)
  }

  render() {
    return (<div className="terminal-container">
      <div id="terminal" />
    </div>)
  }
}