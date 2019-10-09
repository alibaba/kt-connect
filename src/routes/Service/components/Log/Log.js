import React, { Component } from 'react';
import request from 'superagent'

export default class PodLog extends Component {

  constructor(props) {
    super(props);
    this.state = {
      logs: [],
    }
  }

  componentDidMount() {
    this.getLogs(this.props);
  }

  componentWillReceiveProps(nextProps) {
    this.getLogs(nextProps);
  }

  getLogs = (props) => {
    this.setState({ logs: [{content: 'Loading'}] })
    const { namespace, name, container } = props;
    request.get(`/api/cluster/namespaces/${namespace}/pods/${name}/log?container=${container}`)
      .then(res => {
        this.setState({ ...res.body })
      })
  }

  render() {
    const { logs } = this.state;
    return (<div className="log-container">{
      logs.map((log, index) => {
        return (<p key={`log-${index}`}><span className="timestamp">{log.timestamp}</span><span className="content">{log.content}</span></p>)
      })
    }</div>)
  }
}