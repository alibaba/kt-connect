import React, { Component } from "react";
import { Steps, Tag } from 'antd';
const { Step } = Steps;

export class EndpointMenu extends Component {
  render() {
    return (<div>EndpointMenu</div>)
  }
}

export class PodMenu extends Component {
  render() {
    const { component } = this.props;
    if (!component) {
      console.log(this.props, "debug content");
      return null;
    }
    return (<div>
      <h2 className="sub-head">{component.kind} Status</h2>
      <h3>Component Info:</h3>
      <Tag>{component.name}</Tag>
    </div>)
  }
}

export class ComponentMenu extends Component {
  render() {
    const { component } = this.props;
    const type = component.metadata.labels['kt-component'];
    return (<div>
      <h2 className="sub-head">Component Status</h2>
      <h3>Info:</h3>
      <Tag color="blue">{component.status.phase}</Tag>
      <Tag color="blue">{component.metadata.labels['kt-component']}</Tag>
      <h3>Container Statuses:</h3>
      <div>
        {
          component.status.containerStatuses.map((item, index) => {
            return <Tag key={`tag-${index}`} color="blue">{item.image}-{item.ready}</Tag>
          })
        }
      </div>
      <br />
      <h3>Component Topo:</h3>
      <Steps direction="vertical" size="small" current={4}>
        <Step title="Request" description="" />
        {type === 'mesh' && <Step title={`To Version:${component.metadata.labels['version']} (Service Mesh)`} description="" />}
        <Step title={`Request To pod: ${component.metadata.name}`} description={`${component.status.podIP}`} />
        <Step title="Forward to local:" description={`${component.metadata.labels['remoteAddress']}`} />
      </Steps>
    </div>)
  }
}

export class ServiceMenu extends Component {
  render() {
    return (<div>Service Component</div>)
  }
}

export default function Menus({ type, ...restProps }) {
  switch (type) {
    case 'component':
      return <ComponentMenu {...restProps} />;
    case 'pod':
      return <PodMenu {...restProps} />;
    default:
      return null;
  }
}