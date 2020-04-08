import React, { Component } from "react";
import request from 'superagent'
import { Tag, Icon } from "antd";
import OwnerReferences from './OwnerReferences';

export default class OwnerReference extends Component {
  constructor(props) {
    super(props);
    this.state = {
      open: false,
      current: {},
      ownerReferences: [],
    }
  }

  componentDidMount() {
    this.getRefernce();
  }

  getRefernce() {
    const { ownerReference, namespace } = this.props;
    request
      .get(`/api/cluster/namespaces/${namespace}/${ownerReference.kind.toLowerCase()}s/${ownerReference.name}`)
      .then((res) => {
        const current = res.body;
        this.setState({ current, ownerReferences: current.metadata.ownerReferences || [] });
      })
  }

  render() {
    const { ownerReference, namespace, type, targetPorts } = this.props;
    const { ownerReferences } = this.state;

    console.log(targetPorts);

    return (
      <div className="reference">
        {ownerReferences.length > 0 ? <Icon type={this.state.open ? "caret-down" : "caret-right"} onClick={() => { this.setState({ open: !this.state.open }) }} /> : <Icon type="line" />}
        <Tag>{ownerReference.kind}:{ownerReference.name}</Tag>
        {ownerReference.kind === 'Deployment' && type !== 'system' && <div >
          <pre className="command-line">{`ktctl -namespace ${namespace} mesh ${ownerReference.name} --expose=${targetPorts[0]}`}</pre>
          <pre className="command-line">{`ktctl -namespace ${namespace} exchange ${ownerReference.name} --expose=${targetPorts[0]}`}</pre>
        </div>
        }
        {this.state.open && ownerReferences.length > 0 && <OwnerReferences ownerReferences={ownerReferences} namespace={namespace} type={type} targetPorts={targetPorts} />}
      </div>
    )
  }
}